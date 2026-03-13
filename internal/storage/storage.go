package storage

import (
	"context"
	"io"
	"time"
)

// Store represents an object storage bucket.
type Store interface {
	// Upload stores an object at the given key.
	Upload(ctx context.Context, key string, body io.Reader, contentType string) error

	// PresignedURL returns a temporary access URL for a private object.
	PresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error)

	// PublicURL returns the stable public URL for an object.
	PublicURL(key string) (string, error)

	// Delete removes an object.
	Delete(ctx context.Context, key string) error
}