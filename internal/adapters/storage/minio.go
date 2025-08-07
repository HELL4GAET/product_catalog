package storage

import (
	"context"
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"io"
)

type MinioStorage struct {
	client *minio.Client
	bucket string
}

func NewMinioStorage(cfg *Config) (*MinioStorage, error) {
	cli, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	if err = cli.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{}); err != nil {
		exists, _ := cli.BucketExists(ctx, cfg.Bucket)
		if !exists {
			return nil, fmt.Errorf("bucket %s inaccessible: %w", cfg.Bucket, err)
		}
	}
	return &MinioStorage{client: cli, bucket: cfg.Bucket}, nil
}

func (m *MinioStorage) Upload(
	ctx context.Context,
	key string,
	body io.Reader,
	size int64,
	contentType string,
) (string, error) {
	info, err := m.client.PutObject(ctx, m.bucket, key, body, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s/%s", m.client.EndpointURL(), m.bucket, info.Key), nil
}
