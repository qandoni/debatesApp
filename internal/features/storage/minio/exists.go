package minio

import (
	"context"
	"errors"
	"fmt"

	"github.com/minio/minio-go/v7"
)

func (s *Storage) ObjectExists(
	ctx context.Context,
	objectName string,
) (bool, error) {
	_, err := s.client.StatObject(
		ctx,
		s.bucket,
		objectName,
		minio.StatObjectOptions{},
	)
	if err != nil {
		var minioErr minio.ErrorResponse
		if errors.As(err, &minioErr) && minioErr.Code == "NoSuchKey" {
			return false, nil
		}
		return false, fmt.Errorf("stat object %q: %w", objectName, err)
	}
	return true, nil
}
