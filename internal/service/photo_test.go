package service_test

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Xevion/motophoto/internal/database/db"
	"github.com/Xevion/motophoto/internal/service"
	"github.com/Xevion/motophoto/internal/testutil"
	"github.com/Xevion/motophoto/internal/testutil/dbfactory"
)

// fakeStore implements storage.Store for testing photo upload flows.
type fakeStore struct {
	uploaded   map[string][]byte
	downloadFn func(key string) (io.ReadCloser, error)
}

func newFakeStore() *fakeStore {
	return &fakeStore{uploaded: make(map[string][]byte)}
}

func (s *fakeStore) Upload(_ context.Context, key string, body io.Reader, _ string) error {
	data, err := io.ReadAll(body)
	if err != nil {
		return err
	}
	s.uploaded[key] = data
	return nil
}

func (s *fakeStore) PresignedURL(context.Context, string, time.Duration) (string, error) {
	return "http://test/presigned-get", nil
}

func (s *fakeStore) PresignedPUT(context.Context, string, string, time.Duration) (string, error) {
	return "http://test/presigned-put", nil
}

func (s *fakeStore) Download(_ context.Context, key string) (io.ReadCloser, error) {
	if s.downloadFn != nil {
		return s.downloadFn(key)
	}
	if data, ok := s.uploaded[key]; ok {
		return io.NopCloser(bytes.NewReader(data)), nil
	}
	return io.NopCloser(strings.NewReader("")), nil
}

func (s *fakeStore) PublicURL(key string) string          { return "http://test/" + key }
func (s *fakeStore) Delete(context.Context, string) error { return nil }

func makeTestJPEG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 100, 80))
	for x := range 100 {
		for y := range 80 {
			img.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}
	var buf bytes.Buffer
	require.NoError(t, jpeg.Encode(&buf, img, nil))
	return buf.Bytes()
}

func TestPhotoService_InitUpload(t *testing.T) {
	t.Parallel()
	env := testutil.NewEnv(t)
	ctx := t.Context()

	privateStore := newFakeStore()
	publicStore := newFakeStore()
	svc := service.NewPhotoService(env.Queries, privateStore, publicStore)

	userID := dbfactory.User(ctx, t, env.Pool, nil)
	event := dbfactory.Event(ctx, t, env.Pool, env.Events, &dbfactory.EventOpts{
		PhotographerID: &userID,
	})
	gallery := dbfactory.Gallery(ctx, t, env.Galleries, event.ID, nil)

	t.Run("valid", func(t *testing.T) {
		t.Parallel()
		result, err := svc.InitUpload(ctx, event.ID, gallery.ID, userID, service.InitUploadParams{
			Filename:    "test.jpg",
			ContentType: "image/jpeg",
			SizeBytes:   1024,
		})
		require.NoError(t, err)
		assert.NotEmpty(t, result.PhotoID)
		assert.Equal(t, "http://test/presigned-put", result.UploadURL)
	})

	t.Run("wrong owner", func(t *testing.T) {
		t.Parallel()
		otherUser := dbfactory.User(ctx, t, env.Pool, nil)
		_, err := svc.InitUpload(ctx, event.ID, gallery.ID, otherUser, service.InitUploadParams{
			Filename:    "test.jpg",
			ContentType: "image/jpeg",
			SizeBytes:   1024,
		})
		assert.ErrorIs(t, err, service.ErrForbidden)
	})

	t.Run("event not found", func(t *testing.T) {
		t.Parallel()
		_, err := svc.InitUpload(ctx, "nonexistent", gallery.ID, userID, service.InitUploadParams{
			Filename:    "test.jpg",
			ContentType: "image/jpeg",
			SizeBytes:   1024,
		})
		assert.ErrorIs(t, err, service.ErrNotFound)
	})

	t.Run("gallery not found", func(t *testing.T) {
		t.Parallel()
		_, err := svc.InitUpload(ctx, event.ID, "nonexistent", userID, service.InitUploadParams{
			Filename:    "test.jpg",
			ContentType: "image/jpeg",
			SizeBytes:   1024,
		})
		assert.ErrorIs(t, err, service.ErrNotFound)
	})

	t.Run("invalid content type", func(t *testing.T) {
		t.Parallel()
		_, err := svc.InitUpload(ctx, event.ID, gallery.ID, userID, service.InitUploadParams{
			Filename:    "test.gif",
			ContentType: "image/gif",
			SizeBytes:   1024,
		})
		assert.ErrorIs(t, err, service.ErrValidation)
	})

	t.Run("size too large", func(t *testing.T) {
		t.Parallel()
		_, err := svc.InitUpload(ctx, event.ID, gallery.ID, userID, service.InitUploadParams{
			Filename:    "test.jpg",
			ContentType: "image/jpeg",
			SizeBytes:   60_000_000,
		})
		assert.ErrorIs(t, err, service.ErrValidation)
	})

	t.Run("zero size", func(t *testing.T) {
		t.Parallel()
		_, err := svc.InitUpload(ctx, event.ID, gallery.ID, userID, service.InitUploadParams{
			Filename:    "test.jpg",
			ContentType: "image/jpeg",
			SizeBytes:   0,
		})
		assert.ErrorIs(t, err, service.ErrValidation)
	})
}

func TestPhotoService_ConfirmUpload(t *testing.T) {
	t.Parallel()
	env := testutil.NewEnv(t)
	ctx := t.Context()

	userID := dbfactory.User(ctx, t, env.Pool, nil)
	event := dbfactory.Event(ctx, t, env.Pool, env.Events, &dbfactory.EventOpts{
		PhotographerID: &userID,
	})
	gallery := dbfactory.Gallery(ctx, t, env.Galleries, event.ID, nil)

	testJPEG := makeTestJPEG(t)

	t.Run("valid", func(t *testing.T) {
		t.Parallel()
		privateStore := newFakeStore()
		publicStore := newFakeStore()
		privateStore.downloadFn = func(_ string) (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(testJPEG)), nil
		}
		svc := service.NewPhotoService(env.Queries, privateStore, publicStore)

		result, err := svc.InitUpload(ctx, event.ID, gallery.ID, userID, service.InitUploadParams{
			Filename:    "test.jpg",
			ContentType: "image/jpeg",
			SizeBytes:   1024,
		})
		require.NoError(t, err)

		photo, err := svc.ConfirmUpload(ctx, event.ID, gallery.ID, result.PhotoID, userID)
		require.NoError(t, err)
		assert.Equal(t, result.PhotoID, photo.ID)
		assert.Equal(t, "test.jpg", photo.Filename)
		assert.Equal(t, "image/jpeg", photo.ContentType)
		require.NotNil(t, photo.Width)
		require.NotNil(t, photo.Height)
		assert.Equal(t, int32(100), *photo.Width)
		assert.Equal(t, int32(80), *photo.Height)
		assert.NotEmpty(t, photo.PreviewURL)
		assert.Greater(t, photo.SizeBytes, int64(0))
	})

	t.Run("photo not found", func(t *testing.T) {
		t.Parallel()
		svc := service.NewPhotoService(env.Queries, newFakeStore(), newFakeStore())
		_, err := svc.ConfirmUpload(ctx, event.ID, gallery.ID, "nonexistent", userID)
		assert.ErrorIs(t, err, service.ErrNotFound)
	})

	t.Run("double confirm", func(t *testing.T) {
		t.Parallel()
		privateStore := newFakeStore()
		publicStore := newFakeStore()
		privateStore.downloadFn = func(_ string) (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(testJPEG)), nil
		}
		svc := service.NewPhotoService(env.Queries, privateStore, publicStore)

		result, err := svc.InitUpload(ctx, event.ID, gallery.ID, userID, service.InitUploadParams{
			Filename:    "double.jpg",
			ContentType: "image/jpeg",
			SizeBytes:   1024,
		})
		require.NoError(t, err)

		_, err = svc.ConfirmUpload(ctx, event.ID, gallery.ID, result.PhotoID, userID)
		require.NoError(t, err)

		_, err = svc.ConfirmUpload(ctx, event.ID, gallery.ID, result.PhotoID, userID)
		assert.ErrorIs(t, err, service.ErrConflict)
	})

	t.Run("wrong owner", func(t *testing.T) {
		t.Parallel()
		svc := service.NewPhotoService(env.Queries, newFakeStore(), newFakeStore())

		result, err := svc.InitUpload(ctx, event.ID, gallery.ID, userID, service.InitUploadParams{
			Filename:    "owner.jpg",
			ContentType: "image/jpeg",
			SizeBytes:   1024,
		})
		require.NoError(t, err)

		otherUser := dbfactory.User(ctx, t, env.Pool, &dbfactory.UserOpts{
			Role: new(string(db.UserRolePhotographer)),
		})
		_, err = svc.ConfirmUpload(ctx, event.ID, gallery.ID, result.PhotoID, otherUser)
		assert.ErrorIs(t, err, service.ErrForbidden)
	})
}
