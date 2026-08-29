package minio

import (
	"context"
	"fmt"
)

func (s *Storage) GetURL(
	ctx context.Context,
	objectName string,
) (string, error) {
	scheme := "http"

	if s.secure {
		scheme = "https"
	}

	return fmt.Sprintf(
		"%s://%s/%s/%s",
		scheme,
		s.endpoint,
		s.bucket,
		objectName,
	), nil
}
