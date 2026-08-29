package minio

import (
	"context"
	"fmt"

	minioSDK "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Storage struct {
	client   *minioSDK.Client
	bucket   string
	endpoint string
	secure   bool
}

func NewStorage(
	config Config,
) (*Storage, error) {
	client, err := minioSDK.New(config.Endpoint, &minioSDK.Options{
		Creds: credentials.NewStaticV4(
			config.AccessKey,
			config.SecretKey,
			"",
		),
		Secure: config.Secure,
	})
	if err != nil {
		return &Storage{}, fmt.Errorf("create minio client:")
	}
	return &Storage{
		client:   client,
		bucket:   config.Bucket,
		endpoint: config.Endpoint,
		secure:   config.Secure,
	}, nil
}

func (s *Storage) EnsureBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return fmt.Errorf("check bucket %q: %w", s.bucket, err)
	}
	if exists {
		return nil
	}
	if err := s.client.MakeBucket(
		ctx,
		s.bucket,
		minioSDK.MakeBucketOptions{},
	); err != nil {
		return fmt.Errorf("create bucket %q: %w", s.bucket, err)
	}
	return nil
}

func (s *Storage) SetPublicReadPolicy(ctx context.Context) error {
	policy := fmt.Sprintf(`{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Principal": {
					"AWS": ["*"]
				},
				"Action": [
					"s3:GetObject"
				],
				"Resource": [
					"arn:aws:s3:::%s/*"
				]
			}
		]
	}`, s.bucket)

	if err := s.client.SetBucketPolicy(
		ctx,
		s.bucket,
		policy,
	); err != nil {
		return fmt.Errorf("set bucket policy: %w", err)
	}

	return nil
}
