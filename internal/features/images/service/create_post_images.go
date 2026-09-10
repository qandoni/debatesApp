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
	for i, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			return []domain.PostImage{}, fmt.Errorf("open image: %w", err)
		}

		buffer := make([]byte, 512)
		n, err := file.Read(buffer)
		if err != nil {
			return []domain.PostImage{}, fmt.Errorf("read buffer: %w", err)
		}
		contentType := http.DetectContentType(buffer[:n])
		extension, ok := core_config.ExtensionByContentType(contentType)
		if !ok {
			return []domain.PostImage{}, core_errors.ErrUnsupportedMediaType
		}

		if _, err := file.Seek(0, io.SeekStart); err != nil {
			return []domain.PostImage{}, fmt.Errorf("seek: %w", err)
		}

		objectName := fmt.Sprintf(
			"posts/%d/%s%s",
			postID,
			uuid.New().String(),
			extension,
		)

		err = s.storage.Upload(ctx, objectName, file, fileHeader.Size, fileHeader.Header.Get("Content-Type"))
		closeErr := file.Close()
		if err != nil {
			return []domain.PostImage{}, fmt.Errorf("upload image: %w", err)
		}
		if closeErr != nil {
			return []domain.PostImage{}, fmt.Errorf("close image: %w", err)
		}
		image := domain.PostImage{
			PostID:       postID,
			ImageURL:     objectName,
			DisplayOrder: i + 1,
		}
		createdImage, err := s.postImagesRepository.CreatePostImage(ctx, image)
		if err != nil {
			return []domain.PostImage{}, fmt.Errorf("create post image: %w", err)
		}
		images = append(images, createdImage)
	}
	return images, nil

}
