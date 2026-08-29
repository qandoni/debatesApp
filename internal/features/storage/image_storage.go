package storage

import (
	"context"
	"io"
)

type ImageStorage interface {
	Upload(
		ctx context.Context,
		objectName string,
		reader io.Reader,
		size int64,
		contentType string,
	) error

	Delete(
		ctx context.Context,
		objectName string,
	) error

	GetURL(
		ctx context.Context,
		objectName string,
	) (string, error)
}
