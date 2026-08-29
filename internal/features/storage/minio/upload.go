package minio

import (
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
)

func (s *Storage) Upload(
	ctx context.Context,
	objectName string,
	reader io.Reader,
	size int64,
	contentType string,
) error {
	_, err := s.client.PutObject(
		ctx,
		s.bucket,
		objectName,
		reader,
		size,
		minio.PutObjectOptions{
			ContentType: contentType,
		},
	)
	if err != nil {
		return fmt.Errorf("upload object %q: %w", objectName, err)
	}
	return err
}
