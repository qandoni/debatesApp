package images_service

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/google/uuid"
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
	uploadedObjects := make([]string, 0, len(files))

	for i, fileHeader := range files {
		objectName, err := s.uploadPostImage(
			ctx,
			postID,
			fileHeader,
		)
		if err != nil {

			_ = s.deleteUploadedObjects(ctx, uploadedObjects)

			return []domain.PostImage{}, fmt.Errorf(
				"upload post image: %w",
				err,
			)
		}

		uploadedObjects = append(uploadedObjects, objectName)

		images = append(images, domain.PostImage{
			PostID:       postID,
			ImageURL:     objectName,
			DisplayOrder: i + 1,
		})
	}

	err = s.txManager.WithinTransaction(ctx, func(ctx context.Context) error {
		for i := range images {
			createdImage, err := s.postImagesRepository.CreatePostImage(
				ctx,
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
		_ = s.deleteUploadedObjects(ctx, uploadedObjects)

		return []domain.PostImage{}, fmt.Errorf(
			"create post images transaction: %w",
			err,
		)
	}

	return images, nil
}

func (s *ImagesService) deleteUploadedObjects(
	ctx context.Context,
	objectNames []string,
) error {
	for _, objectName := range objectNames {
		if err := s.storage.Delete(ctx, objectName); err != nil {
			return fmt.Errorf(
				"delete uploaded image %q: %w",
				objectName,
				err,
			)
		}
	}

	return nil
}

func (s *ImagesService) uploadPostImage(
	ctx context.Context,
	postID int,
	fileHeader *multipart.FileHeader,
) (string, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("open image: %w", err)
	}
	defer file.Close()

	buffer := make([]byte, 512)

	n, err := file.Read(buffer)
	if err != nil {
		return "", fmt.Errorf("read buffer: %w", err)
	}

	contentType := http.DetectContentType(buffer[:n])

	extension, ok := core_config.ExtensionByContentType(contentType)
	if !ok {
		return "", core_errors.ErrUnsupportedMediaType
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("seek: %w", err)
	}

	objectName := fmt.Sprintf(
		"posts/%d/%s%s",
		postID,
		uuid.New().String(),
		extension,
	)

	if err := s.storage.Upload(
		ctx,
		objectName,
		file,
		fileHeader.Size,
		contentType,
	); err != nil {
		return "", fmt.Errorf("upload image: %w", err)
	}

	return objectName, nil
}

// func (s *ImagesService) GetByPostID(
// 	ctx context.Context,
// 	postID int,
// ) ([]domain.PostImage, error) {
// 	_, err := s.postsRepository.GetPost(ctx, postID)
// 	if err != nil {
// 		return nil, fmt.Errorf("get post: %w", err)
// 	}
// 	images, err := s.postImagesRepository.GetByPostID(ctx, postID)
// 	if err != nil {
// 		return []domain.PostImage{}, fmt.Errorf("get post images: %w", err)
// 	}
// 	return images, nil
// } нужно будет перенести проверку на существование поста в пост сервис

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
) error {
	for _, image := range images {
		if err := s.storage.Delete(ctx, image.ImageURL); err != nil {
			return fmt.Errorf(
				"delete image %q from storage: %w",
				image.ImageURL,
				err,
			)
		}
	}

	return nil
}
