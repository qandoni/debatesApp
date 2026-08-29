package minio

import (
	"context"
	"fmt"

	"github.com/minio/minio-go/v7"
)

func (s *Storage) Delete(
	ctx context.Context,
	objectName string,
) error {
	err := s.client.RemoveObject(
		ctx,
		s.bucket,
		objectName,
		minio.RemoveObjectOptions{},
	)
	if err != nil {
		return fmt.Errorf("delete object %q: %w", objectName, err)
	}
	return nil
}
