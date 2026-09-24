package storage

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type minioBackend struct {
	client *minio.Client
	bucket string
}

// New builds a Store backed by MinIO/S3 (the spec's default — run via
// docker-compose). Use NewLocal instead to run without Docker.
func New(endpoint, accessKey, secretKey, bucket string, useSSL bool) (*Store, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("create minio client: %w", err)
	}
	return &Store{backend: &minioBackend{client: client, bucket: bucket}}, nil
}

func (b *minioBackend) EnsureBucket(ctx context.Context) error {
	exists, err := b.client.BucketExists(ctx, b.bucket)
	if err != nil {
		return fmt.Errorf("check bucket: %w", err)
	}
	if !exists {
		if err := b.client.MakeBucket(ctx, b.bucket, minio.MakeBucketOptions{}); err != nil {
			return fmt.Errorf("create bucket: %w", err)
		}
	}
	return nil
}

func (b *minioBackend) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	_, err := b.client.BucketExists(ctx, b.bucket)
	return err
}

func (b *minioBackend) Put(ctx context.Context, objectKey string, data []byte, contentType string) error {
	_, err := b.client.PutObject(ctx, b.bucket, objectKey, bytes.NewReader(data), int64(len(data)),
		minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return fmt.Errorf("put object %s: %w", objectKey, err)
	}
	return nil
}

func (b *minioBackend) Get(ctx context.Context, objectKey string) ([]byte, error) {
	obj, err := b.client.GetObject(ctx, b.bucket, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("get object %s: %w", objectKey, err)
	}
	defer obj.Close()
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(obj); err != nil {
		return nil, fmt.Errorf("read object %s: %w", objectKey, err)
	}
	return buf.Bytes(), nil
}

func (b *minioBackend) SignedURL(ctx context.Context, objectKey string, expiry time.Duration) (string, error) {
	u, err := b.client.PresignedGetObject(ctx, b.bucket, objectKey, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("sign url %s: %w", objectKey, err)
	}
	return u.String(), nil
}
