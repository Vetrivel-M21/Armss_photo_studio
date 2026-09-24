// Package storage wraps the object store used for template images, uploads,
// and renders. Two backends implement the same small interface: MinIO/S3 (the
// spec's default, run via docker-compose) and a local-disk backend for
// running the server without Docker — selected once at boot via config, so
// every handler just calls Store.Put/Get/SignedURL without knowing which one
// is underneath.
package storage

import (
	"context"
	"time"
)

type backend interface {
	EnsureBucket(ctx context.Context) error
	Ping(ctx context.Context) error
	Put(ctx context.Context, objectKey string, data []byte, contentType string) error
	Get(ctx context.Context, objectKey string) ([]byte, error)
	SignedURL(ctx context.Context, objectKey string, expiry time.Duration) (string, error)
}

type Store struct {
	backend backend
}

func (s *Store) EnsureBucket(ctx context.Context) error { return s.backend.EnsureBucket(ctx) }
func (s *Store) Ping(ctx context.Context) error         { return s.backend.Ping(ctx) }

func (s *Store) Put(ctx context.Context, objectKey string, data []byte, contentType string) error {
	return s.backend.Put(ctx, objectKey, data, contentType)
}

func (s *Store) Get(ctx context.Context, objectKey string) ([]byte, error) {
	return s.backend.Get(ctx, objectKey)
}

func (s *Store) SignedURL(ctx context.Context, objectKey string, expiry time.Duration) (string, error) {
	return s.backend.SignedURL(ctx, objectKey, expiry)
}
