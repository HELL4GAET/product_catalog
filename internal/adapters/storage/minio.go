package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioConfig struct {
	Endpoint       string
	PublicEndpoint string
	AccessKey      string
	SecretKey      string
	UseSSL         bool
	Bucket         string
	Region         string
}

type MinioStorage struct {
	client *minio.Client
	signer *minio.Client
	bucket string
}

func NewMinioStorage(cfg *MinioConfig) (*MinioStorage, error) {
	if cfg.Endpoint == "" {
		return nil, fmt.Errorf("storage endpoint required")
	}
	if cfg.AccessKey == "" || cfg.SecretKey == "" {
		return nil, fmt.Errorf("minio credentials required")
	}
	if cfg.Bucket == "" {
		return nil, fmt.Errorf("minio bucket required")
	}

	// 1) Internal client (uses internal endpoint, e.g. "minio:9000")
	internalClient, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("new minio internal client: %w", err)
	}

	ctx := context.Background()
	if err := internalClient.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{}); err != nil {
		exists, _ := internalClient.BucketExists(ctx, cfg.Bucket)
		if !exists {
			return nil, fmt.Errorf("bucket %s inaccessible: %w", cfg.Bucket, err)
		}
	}

	signerEndpoint := cfg.PublicEndpoint
	signerUseSSL := cfg.UseSSL

	var signerClient *minio.Client
	if signerEndpoint == "" || signerEndpoint == cfg.Endpoint {
		signerClient = internalClient
	} else {
		if strings.HasPrefix(signerEndpoint, "http://") || strings.HasPrefix(signerEndpoint, "https://") {
			u, err := url.Parse(signerEndpoint)
			if err == nil {
				signerUseSSL = u.Scheme == "https"
				signerEndpoint = u.Host
			}
		}
		if signerEndpoint == cfg.Endpoint {
			signerClient = internalClient
		} else {
			signCl, err := minio.New(signerEndpoint, &minio.Options{
				Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
				Secure: signerUseSSL,
				Region: cfg.Region,
			})
			if err != nil {
				return nil, fmt.Errorf("new minio signer client: %w", err)
			}
			signerClient = signCl
		}
	}

	return &MinioStorage{
		client: internalClient,
		signer: signerClient,
		bucket: cfg.Bucket,
	}, nil
}

func (m *MinioStorage) Upload(ctx context.Context, key string, body io.Reader, size int64, contentType string) error {
	_, err := m.client.PutObject(ctx, m.bucket, key, body, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

func (m *MinioStorage) GetPresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	reqParams := make(url.Values)
	presignedURL, err := m.signer.PresignedGetObject(ctx, m.bucket, key, expiry, reqParams)
	if err != nil {
		return "", fmt.Errorf("failed to get presigned URL: %w", err)
	}
	return presignedURL.String(), nil
}
