package images_service

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
	core_config "github.com/qandoni/debatesApp/internal/core/config"
	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
)

func (s *AvatarService) UploadAvatar(
	ctx context.Context,
	userID int,
	reader io.Reader,
	size int64,
) error {

	if size > core_config.MaxAvatarSize {
		return core_errors.ErrInvalidArgument
	}

	buffer := make([]byte, 512)

	n, err := reader.Read(buffer)
	if err != nil {
		return fmt.Errorf("read avatar: %w", err)
	}

	contentType := http.DetectContentType(buffer[:n])

	extension, ok := core_config.ExtensionByContentType(contentType)
	if !ok {
		return core_errors.ErrUnsupportedMediaType
	}

	if seeker, ok := reader.(io.Seeker); ok {
		if _, err := seeker.Seek(0, io.SeekStart); err != nil {
			return fmt.Errorf("reset avatar reader: %w", err)
		}
	}

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

	oldAvatarURL, err := s.usersRepository.UpdateAvatarURL(
		ctx,
		userID,
		avatarURL,
	)
	if err != nil {
		_ = s.storage.Delete(ctx, objectName)

		return fmt.Errorf("update avatar url: %w", err)
	}

	if oldAvatarURL != nil && *oldAvatarURL != "" {
		oldObjectName := core_config.ExtractObjectName(*oldAvatarURL)

		if oldObjectName != "" {
			if err := s.storage.Delete(ctx, oldObjectName); err != nil {
				return fmt.Errorf("delete old avatar: %w", err)
			}
		}
	}

	return nil
}
