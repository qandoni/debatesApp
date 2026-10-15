package images_service

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	core_config "github.com/qandoni/debatesApp/internal/core/config"
	"github.com/qandoni/debatesApp/internal/core/domain"
	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
)

func (s *ImagesService) CreatePostImages(
	ctx context.Context,
	userID int,
	postID int,
	files []*multipart.FileHeader,
) ([]domain.PostImage, error) {
	post, err := s.postsRepository.GetPost(ctx, postID)
	if err != nil {
		return []domain.PostImage{}, fmt.Errorf("get post: %w", err)
	}

	if post.AuthorID != userID {
		return []domain.PostImage{}, core_errors.ErrAccessForbidden
	}

	images := make([]domain.PostImage, 0, len(files))
	contents := make([]string, len(files))

	for i, fileHeader := range files {
		image, contentType, err := s.preparePostImage(postID, i+1, fileHeader)
		if err != nil {
			return []domain.PostImage{}, err
		}
		contents[i] = contentType
		images = append(images, image)
	}

	err = s.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		for i := range images {
			createdImage, err := s.postImagesRepository.CreatePostImage(
				txCtx,
				images[i],
			)
			if err != nil {
				return fmt.Errorf("create post image: %w", err)
			}

			images[i] = createdImage
		}

		return nil
	})
	if err != nil {
		return []domain.PostImage{}, fmt.Errorf(
			"create post images transaction: %w",
			err,
		)
	}

	for i, fileHeader := range files {
		if err := s.uploadImage(ctx, images[i].ImageURL, fileHeader, contents[i]); err != nil {

			if cleanupErr := s.minio.DeleteAll(ctx, imageNames(images)); cleanupErr != nil {
				return []domain.PostImage{}, fmt.Errorf(
					"upload post image: %v, minio cleanup failed: %v",
					err,
					cleanupErr,
				)
			}
			return []domain.PostImage{}, fmt.Errorf("upload post image: %w", err)
		}
	}

	err = s.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		return s.postImagesRepository.MarkActiveByIDs(txCtx, imageIDs(images))
	})
	if err != nil {
		return []domain.PostImage{}, fmt.Errorf("activate post images: %w", err)
	}

	return images, nil
}

func (s *ImagesService) preparePostImage(
	postID int,
	order int,
	fileHeader *multipart.FileHeader,
) (domain.PostImage, string, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return domain.PostImage{}, "", fmt.Errorf("open image: %w", err)
	}
	defer file.Close()

	buffer := make([]byte, 512)

	n, err := file.Read(buffer)
	if err != nil {
		return domain.PostImage{}, "", fmt.Errorf("read buffer: %w", err)
	}

	contentType := http.DetectContentType(buffer[:n])

	extension, ok := core_config.ExtensionByContentType(contentType)
	if !ok {
		return domain.PostImage{}, "", core_errors.ErrUnsupportedMediaType
	}

	objectName := s.minio.NewObjectName(
		fmt.Sprintf("posts/%d", postID),
		extension,
	)

	return domain.PostImage{
		PostID:       postID,
		ImageURL:     objectName,
		DisplayOrder: order,
	}, contentType, nil
}

func (s *ImagesService) uploadImage(
	ctx context.Context,
	objectName string,
	fileHeader *multipart.FileHeader,
	contentType string,
) error {
	file, err := fileHeader.Open()
	if err != nil {
		return fmt.Errorf("open image: %w", err)
	}
	defer file.Close()

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("seek: %w", err)
	}

	if err := s.minio.Upload(ctx, objectName, file, fileHeader.Size, contentType); err != nil {
		return err
	}
	return nil
}

func imageNames(images []domain.PostImage) []string {
	names := make([]string, 0, len(images))
	for _, image := range images {
		names = append(names, image.ImageURL)
	}
	return names
}

func imageIDs(images []domain.PostImage) []int {
	ids := make([]int, 0, len(images))
	for _, image := range images {
		ids = append(ids, image.ID)
	}
	return ids
}

func (s *ImagesService) GetByPostID(
	ctx context.Context,
	postID int,
) ([]domain.PostImage, error) {
	images, err := s.postImagesRepository.GetByPostID(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("get post images: %w", err)
	}

	return images, nil
}

func (s *ImagesService) GetByPostIDs(
	ctx context.Context,
	postIDs []int,
) (map[int][]domain.PostImage, error) {
	images, err := s.postImagesRepository.GetByPostIDs(ctx, postIDs)
	if err != nil {
		return nil, fmt.Errorf("get images by post ids: %w", err)
	}

	return images, nil
}

func (s *ImagesService) DeleteRecordsByPostID(
	ctx context.Context,
	postID int,
) error {
	if err := s.postImagesRepository.DeleteByPostID(ctx, postID); err != nil {
		return fmt.Errorf("delete post images: %w", err)
	}

	return nil
}

func (s *ImagesService) DeleteFromStorage(
	ctx context.Context,
	images []domain.PostImage,
) []string {
	return s.minio.DeleteAll(ctx, imageNames(images))
}
