package storage

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// localBackend stores objects as plain files on disk, for running the server
// without Docker/MinIO. Files are served back over HTTP by a dedicated public
// route (see http/handlers/files.go) rather than a real presigned URL, since
// there's no S3-compatible signing to do for a local directory.
type localBackend struct {
	rootDir       string
	publicBaseURL string
}

// NewLocal builds a Store backed by the local filesystem under rootDir,
// serving files back at publicBaseURL + "/files/" + objectKey.
func NewLocal(rootDir, publicBaseURL string) (*Store, error) {
	if err := os.MkdirAll(rootDir, 0o755); err != nil {
		return nil, fmt.Errorf("create local storage dir %s: %w", rootDir, err)
	}
	return &Store{backend: &localBackend{
		rootDir:       rootDir,
		publicBaseURL: strings.TrimRight(publicBaseURL, "/"),
	}}, nil
}

func (b *localBackend) EnsureBucket(ctx context.Context) error {
	return os.MkdirAll(b.rootDir, 0o755)
}

func (b *localBackend) Ping(ctx context.Context) error {
	_, err := os.Stat(b.rootDir)
	return err
}

func (b *localBackend) Put(ctx context.Context, objectKey string, data []byte, contentType string) error {
	path, err := b.resolve(objectKey)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create directory for %s: %w", objectKey, err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write object %s: %w", objectKey, err)
	}
	return nil
}

func (b *localBackend) Get(ctx context.Context, objectKey string) ([]byte, error) {
	path, err := b.resolve(objectKey)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read object %s: %w", objectKey, err)
	}
	return data, nil
}

// SignedURL has no real expiry for local storage — it's a plain URL served by
// the /files route. The expiry parameter is accepted only to satisfy the
// shared backend interface.
func (b *localBackend) SignedURL(ctx context.Context, objectKey string, expiry time.Duration) (string, error) {
	if _, err := b.resolve(objectKey); err != nil {
		return "", err
	}
	segments := strings.Split(objectKey, "/")
	for i, s := range segments {
		segments[i] = url.PathEscape(s)
	}
	return fmt.Sprintf("%s/files/%s", b.publicBaseURL, strings.Join(segments, "/")), nil
}

// resolve maps an object key to a path under rootDir, rejecting any key that
// would escape it (object keys are generated internally with uuid segments,
// but this guards against a future caller passing something untrusted).
func (b *localBackend) resolve(objectKey string) (string, error) {
	cleaned := filepath.Clean("/" + objectKey)
	path := filepath.Join(b.rootDir, cleaned)
	if !strings.HasPrefix(path, filepath.Clean(b.rootDir)+string(filepath.Separator)) {
		return "", errors.New("invalid object key")
	}
	return path, nil
}
