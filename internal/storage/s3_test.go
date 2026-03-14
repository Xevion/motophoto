package storage_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Xevion/motophoto/internal/storage"
)

func newTestStore(t *testing.T, serverURL string, publicBase string) *storage.S3Store {
	t.Helper()

	store, err := storage.NewS3Store(t.Context(), storage.Config{
		Endpoint:   serverURL,
		Region:     "us-east-1",
		AccessKey:  "test",
		SecretKey:  "test",
		Bucket:     "test-bucket",
		PublicBase: publicBase,
	}, func(o *s3.Options) {
		o.Retryer = retry.NewStandard(func(so *retry.StandardOptions) {
			so.MaxAttempts = 1
		})
	})
	require.NoError(t, err)
	return store
}

func TestNewS3Store(t *testing.T) {
	t.Parallel()

	t.Run("succeeds with valid config", func(t *testing.T) {
		t.Parallel()
		store := newTestStore(t, "http://localhost:9000", "https://cdn.example.com")
		assert.NotNil(t, store)
	})

	t.Run("trims trailing slash from public base", func(t *testing.T) {
		t.Parallel()
		store := newTestStore(t, "http://localhost:9000", "https://cdn.example.com/")
		url := store.PublicURL("photos/abc.jpg")
		assert.Equal(t, "https://cdn.example.com/photos/abc.jpg", url)
	})

	t.Run("trims multiple trailing slashes", func(t *testing.T) {
		t.Parallel()
		store := newTestStore(t, "http://localhost:9000", "https://cdn.example.com///")
		url := store.PublicURL("photos/abc.jpg")
		assert.Equal(t, "https://cdn.example.com/photos/abc.jpg", url)
	})
}

func TestS3Store_PublicURL(t *testing.T) {
	t.Parallel()

	t.Run("returns concatenated URL", func(t *testing.T) {
		t.Parallel()
		store := newTestStore(t, "http://localhost:9000", "https://cdn.example.com")
		url := store.PublicURL("events/123/photo.jpg")
		assert.Equal(t, "https://cdn.example.com/events/123/photo.jpg", url)
	})

	t.Run("panics when no public base configured", func(t *testing.T) {
		t.Parallel()
		store := newTestStore(t, "http://localhost:9000", "")
		assert.PanicsWithValue(t,
			"PublicURL called on store with no public base URL configured",
			func() { store.PublicURL("key") },
		)
	})
}

func TestS3Store_Upload(t *testing.T) {
	t.Parallel()

	t.Run("sends PUT with correct key and content type", func(t *testing.T) {
		t.Parallel()

		var gotMethod string
		var gotKey string
		var gotContentType string
		var gotBody string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPut {
				gotMethod = r.Method
				gotKey = r.URL.Path
				gotContentType = r.Header.Get("Content-Type")
				body, _ := io.ReadAll(r.Body)
				gotBody = string(body)
			}
			w.WriteHeader(http.StatusOK)
		}))
		t.Cleanup(srv.Close)

		store := newTestStore(t, srv.URL, "https://cdn.example.com")
		err := store.Upload(t.Context(), "photos/test.jpg", strings.NewReader("image-data"), "image/jpeg")
		require.NoError(t, err)

		assert.Equal(t, http.MethodPut, gotMethod)
		assert.Equal(t, "/test-bucket/photos/test.jpg", gotKey)
		assert.Equal(t, "image/jpeg", gotContentType)
		assert.Equal(t, "image-data", gotBody)
	})

	t.Run("returns error on server failure", func(t *testing.T) {
		t.Parallel()

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		t.Cleanup(srv.Close)

		store := newTestStore(t, srv.URL, "https://cdn.example.com")
		err := store.Upload(t.Context(), "photos/test.jpg", strings.NewReader("data"), "image/jpeg")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "uploading object")
	})
}

func TestS3Store_Delete(t *testing.T) {
	t.Parallel()

	t.Run("sends DELETE with correct key", func(t *testing.T) {
		t.Parallel()

		var gotPath string
		var gotMethod string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotMethod = r.Method
			gotPath = r.URL.Path
			w.WriteHeader(http.StatusNoContent)
		}))
		t.Cleanup(srv.Close)

		store := newTestStore(t, srv.URL, "https://cdn.example.com")
		err := store.Delete(t.Context(), "photos/old.jpg")
		require.NoError(t, err)

		assert.Equal(t, http.MethodDelete, gotMethod)
		assert.Equal(t, "/test-bucket/photos/old.jpg", gotPath)
	})

	t.Run("returns error on server failure", func(t *testing.T) {
		t.Parallel()

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		t.Cleanup(srv.Close)

		store := newTestStore(t, srv.URL, "https://cdn.example.com")
		err := store.Delete(t.Context(), "photos/test.jpg")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "deleting object")
	})
}

func TestS3Store_PresignedURL(t *testing.T) {
	t.Parallel()

	t.Run("returns a signed URL", func(t *testing.T) {
		t.Parallel()

		// PresignedURL is a client-side operation (no HTTP call), so
		// the mock server isn't hit, but we need a valid endpoint for store creation.
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		t.Cleanup(srv.Close)

		store := newTestStore(t, srv.URL, "https://cdn.example.com")
		url, err := store.PresignedURL(t.Context(), "photos/private.jpg", 15*time.Minute)
		require.NoError(t, err)

		assert.Contains(t, url, "photos/private.jpg")
		assert.Contains(t, url, "X-Amz-Signature")
		assert.Contains(t, url, "X-Amz-Expires=900")
	})
}

func TestS3Store_PresignedPUT(t *testing.T) {
	t.Parallel()

	t.Run("returns a signed PUT URL", func(t *testing.T) {
		t.Parallel()

		// PresignedPUT is a client-side operation (no HTTP call), so
		// the mock server isn't hit, but we need a valid endpoint for store creation.
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		t.Cleanup(srv.Close)

		store := newTestStore(t, srv.URL, "https://cdn.example.com")
		url, err := store.PresignedPUT(t.Context(), "photos/upload.jpg", "image/jpeg", 15*time.Minute)
		require.NoError(t, err)

		assert.Contains(t, url, "photos/upload.jpg")
		assert.Contains(t, url, "X-Amz-Signature")
		assert.Contains(t, url, "X-Amz-Expires=900")
	})
}

func TestS3Store_Download(t *testing.T) {
	t.Parallel()

	t.Run("returns object body", func(t *testing.T) {
		t.Parallel()

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet {
				w.Header().Set("Content-Type", "image/jpeg")
				_, _ = w.Write([]byte("image-data"))
				return
			}
			w.WriteHeader(http.StatusOK)
		}))
		t.Cleanup(srv.Close)

		store := newTestStore(t, srv.URL, "")
		body, err := store.Download(t.Context(), "photos/test.jpg")
		require.NoError(t, err)
		t.Cleanup(func() { require.NoError(t, body.Close()) })

		data, err := io.ReadAll(body)
		require.NoError(t, err)
		assert.Equal(t, "image-data", string(data))
	})

	t.Run("returns error on server failure", func(t *testing.T) {
		t.Parallel()

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		t.Cleanup(srv.Close)

		store := newTestStore(t, srv.URL, "")
		_, err := store.Download(t.Context(), "photos/missing.jpg")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "downloading object")
	})
}
