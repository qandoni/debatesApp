package images_service

import (
	"context"
	"fmt"
	"io"

	"github.com/google/uuid"
)

func (s *ImagesService) UploadAvatar(
	ctx context.Context,
	userID int,
	reader io.Reader,
	size int64,
	contentType string,
	extension string,
) error {

	objectName := fmt.Sprintf(
		"users/%d/avatar/%s%s",
		userID,
		uuid.New().String(),
		extension,
	)

	if err := s.storage.Upload(
		ctx,
		objectName,
		reader,
		size,
		contentType,
	); err != nil {
		return fmt.Errorf("upload avatar: %w", err)
	}

	avatarURL, err := s.storage.GetURL(
		ctx,
		objectName,
	)
	if err != nil {
		_ = s.storage.Delete(ctx, objectName)

		return fmt.Errorf("get avatar url: %w", err)
	}

	if err := s.usersRepository.UpdateAvatarURL(
		ctx,
		userID,
		avatarURL,
	); err != nil {
		_ = s.storage.Delete(ctx, objectName)

		return fmt.Errorf("update avatar url: %w", err)
	}

	return nil
}

//TODO удалять старый аватар
