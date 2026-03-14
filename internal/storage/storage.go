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

	// PublicURL returns the stable CDN URL for a public object.
	// Panics if called on a store with no public base URL configured.
	PublicURL(key string) string

	// Delete removes an object.
	Delete(ctx context.Context, key string) error
}
