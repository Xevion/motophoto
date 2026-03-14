package server_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Xevion/motophoto/internal/server"
	"github.com/Xevion/motophoto/internal/testutil"
	"github.com/Xevion/motophoto/internal/testutil/dbfactory"
)

func TestHandleInitUpload(t *testing.T) {
	t.Parallel()
	env := testutil.NewEnv(t)
	ctx := t.Context()

	session, photographerID := testutil.LoginPhotographer(t, env.Handler, env.Pool)
	event := dbfactory.Event(ctx, t, env.Pool, env.Events, &dbfactory.EventOpts{
		PhotographerID: &photographerID,
	})
	gallery := dbfactory.Gallery(ctx, t, env.Galleries, event.ID, nil)

	url := fmt.Sprintf("/api/v1/events/%s/galleries/%s/photos/upload", event.ID, gallery.ID)

	t.Run("valid", func(t *testing.T) {
		t.Parallel()
		body := `{"filename":"test.jpg","content_type":"image/jpeg","size_bytes":1024}`
		rr := doRequestWithCookies(t, env.Handler, http.MethodPost, url, body, []*http.Cookie{session})

		assert.Equal(t, http.StatusCreated, rr.Code)

		var resp server.ItemResponse[server.InitUploadResponse]
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.NotEmpty(t, resp.Data.PhotoID)
		assert.NotEmpty(t, resp.Data.UploadURL)
	})

	t.Run("missing filename", func(t *testing.T) {
		t.Parallel()
		body := `{"content_type":"image/jpeg","size_bytes":1024}`
		rr := doRequestWithCookies(t, env.Handler, http.MethodPost, url, body, []*http.Cookie{session})
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("invalid content type", func(t *testing.T) {
		t.Parallel()
		body := `{"filename":"test.gif","content_type":"image/gif","size_bytes":1024}`
		rr := doRequestWithCookies(t, env.Handler, http.MethodPost, url, body, []*http.Cookie{session})
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("size too large", func(t *testing.T) {
		t.Parallel()
		body := `{"filename":"test.jpg","content_type":"image/jpeg","size_bytes":60000000}`
		rr := doRequestWithCookies(t, env.Handler, http.MethodPost, url, body, []*http.Cookie{session})
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("unauthenticated", func(t *testing.T) {
		t.Parallel()
		body := `{"filename":"test.jpg","content_type":"image/jpeg","size_bytes":1024}`
		rr := doRequest(t, env.Handler, http.MethodPost, url, body)
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("wrong owner", func(t *testing.T) {
		t.Parallel()
		otherSession, _ := testutil.LoginPhotographer(t, env.Handler, env.Pool)
		body := `{"filename":"test.jpg","content_type":"image/jpeg","size_bytes":1024}`
		rr := doRequestWithCookies(t, env.Handler, http.MethodPost, url, body, []*http.Cookie{otherSession})
		assert.Equal(t, http.StatusForbidden, rr.Code)
	})

	t.Run("event not found", func(t *testing.T) {
		t.Parallel()
		badURL := fmt.Sprintf("/api/v1/events/nonexistent/galleries/%s/photos/upload", gallery.ID)
		body := `{"filename":"test.jpg","content_type":"image/jpeg","size_bytes":1024}`
		rr := doRequestWithCookies(t, env.Handler, http.MethodPost, badURL, body, []*http.Cookie{session})
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("gallery not found", func(t *testing.T) {
		t.Parallel()
		badURL := fmt.Sprintf("/api/v1/events/%s/galleries/nonexistent/photos/upload", event.ID)
		body := `{"filename":"test.jpg","content_type":"image/jpeg","size_bytes":1024}`
		rr := doRequestWithCookies(t, env.Handler, http.MethodPost, badURL, body, []*http.Cookie{session})
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("customer role rejected", func(t *testing.T) {
		t.Parallel()
		customerSession, _ := testutil.LoginCustomer(t, env.Handler, env.Pool)
		body := `{"filename":"test.jpg","content_type":"image/jpeg","size_bytes":1024}`
		rr := doRequestWithCookies(t, env.Handler, http.MethodPost, url, body, []*http.Cookie{customerSession})
		assert.Equal(t, http.StatusForbidden, rr.Code)
	})
}

func TestHandleConfirmUpload(t *testing.T) {
	t.Parallel()
	env := testutil.NewEnv(t)
	ctx := t.Context()

	session, photographerID := testutil.LoginPhotographer(t, env.Handler, env.Pool)
	event := dbfactory.Event(ctx, t, env.Pool, env.Events, &dbfactory.EventOpts{
		PhotographerID: &photographerID,
	})
	gallery := dbfactory.Gallery(ctx, t, env.Galleries, event.ID, nil)

	t.Run("photo not found", func(t *testing.T) {
		t.Parallel()
		url := fmt.Sprintf("/api/v1/events/%s/galleries/%s/photos/nonexistent/confirm", event.ID, gallery.ID)
		rr := doRequestWithCookies(t, env.Handler, http.MethodPost, url, "", []*http.Cookie{session})
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("unauthenticated", func(t *testing.T) {
		t.Parallel()
		url := fmt.Sprintf("/api/v1/events/%s/galleries/%s/photos/someid/confirm", event.ID, gallery.ID)
		rr := doRequest(t, env.Handler, http.MethodPost, url, "")
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("wrong owner", func(t *testing.T) {
		t.Parallel()
		otherSession, _ := testutil.LoginPhotographer(t, env.Handler, env.Pool)
		url := fmt.Sprintf("/api/v1/events/%s/galleries/%s/photos/someid/confirm", event.ID, gallery.ID)
		rr := doRequestWithCookies(t, env.Handler, http.MethodPost, url, "", []*http.Cookie{otherSession})
		assert.Equal(t, http.StatusForbidden, rr.Code)
	})
}
