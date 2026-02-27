package server_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Xevion/motophoto/internal/server"
	"github.com/Xevion/motophoto/internal/testutil"
	"github.com/Xevion/motophoto/internal/testutil/dbfactory"
)

// doRequest is a small helper to reduce httptest boilerplate.
func doRequest(t *testing.T, handler http.Handler, method, path string, body string) *httptest.ResponseRecorder {
	t.Helper()
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr
}

func TestHandleListEvents(t *testing.T) {
	t.Parallel()
	env := testutil.NewEnv(t)
	ctx := t.Context()

	event := dbfactory.Event(ctx, t, env.Pool, env.Events, &dbfactory.EventOpts{
		Status: new("published"),
	})

	rr := doRequest(t, env.Handler, http.MethodGet, "/api/v1/events", "")

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp server.ListResponse[server.EventResponse]
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	require.Len(t, resp.Data, 1)
	assert.Equal(t, event.ID, resp.Data[0].ID)
	assert.Equal(t, event.Name, resp.Data[0].Name)
	assert.Equal(t, event.Sport, resp.Data[0].Sport)
}

func TestHandleGetEvent(t *testing.T) {
	t.Parallel()
	env := testutil.NewEnv(t)
	ctx := t.Context()

	event := dbfactory.Event(ctx, t, env.Pool, env.Events, &dbfactory.EventOpts{
		Status: new("published"),
	})

	t.Run("found", func(t *testing.T) {
		t.Parallel()
		rr := doRequest(t, env.Handler, http.MethodGet, "/api/v1/events/"+event.ID, "")

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp server.ItemResponse[server.EventResponse]
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.Equal(t, event.ID, resp.Data.ID)
		assert.Equal(t, event.Name, resp.Data.Name)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		rr := doRequest(t, env.Handler, http.MethodGet, "/api/v1/events/nonexistent", "")
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

func TestHandleCreateEvent(t *testing.T) {
	t.Parallel()
	env := testutil.NewEnv(t)
	ctx := t.Context()

	dbfactory.User(ctx, t, env.Pool, &dbfactory.UserOpts{ID: new("stub-user")})

	t.Run("valid", func(t *testing.T) {
		t.Parallel()
		body := `{"name":"New Event","slug":"new-event","sport":"motocross"}`
		rr := doRequest(t, env.Handler, http.MethodPost, "/api/v1/events", body)

		assert.Equal(t, http.StatusCreated, rr.Code)

		var resp server.ItemResponse[server.EventResponse]
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.Equal(t, "new-event", resp.Data.Slug)
		assert.Equal(t, "New Event", resp.Data.Name)
		assert.NotEmpty(t, resp.Data.ID)
	})

	t.Run("missing required fields", func(t *testing.T) {
		t.Parallel()
		body := `{"name":"Incomplete"}`
		rr := doRequest(t, env.Handler, http.MethodPost, "/api/v1/events", body)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		t.Parallel()
		rr := doRequest(t, env.Handler, http.MethodPost, "/api/v1/events", `{bad`)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestHandleUpdateEvent(t *testing.T) {
	t.Parallel()
	env := testutil.NewEnv(t)
	ctx := t.Context()

	event := dbfactory.Event(ctx, t, env.Pool, env.Events, nil)

	t.Run("valid partial update", func(t *testing.T) {
		t.Parallel()
		body := `{"name":"Updated Via API"}`
		rr := doRequest(t, env.Handler, http.MethodPatch, "/api/v1/events/"+event.ID, body)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp server.ItemResponse[server.EventResponse]
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.Equal(t, "Updated Via API", resp.Data.Name)
		assert.Equal(t, event.Slug, resp.Data.Slug) // unchanged
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		body := `{"name":"Nope"}`
		rr := doRequest(t, env.Handler, http.MethodPatch, "/api/v1/events/nonexistent", body)
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		t.Parallel()
		rr := doRequest(t, env.Handler, http.MethodPatch, "/api/v1/events/"+event.ID, `{bad`)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestHandleDeleteEvent(t *testing.T) {
	t.Parallel()
	env := testutil.NewEnv(t)
	ctx := t.Context()

	event := dbfactory.Event(ctx, t, env.Pool, env.Events, nil)

	rr := doRequest(t, env.Handler, http.MethodDelete, "/api/v1/events/"+event.ID, "")
	assert.Equal(t, http.StatusNoContent, rr.Code)

	// Verify the event is actually gone.
	rr = doRequest(t, env.Handler, http.MethodGet, "/api/v1/events/"+event.ID, "")
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

// Gallery handler tests

func TestHandleListGalleries(t *testing.T) {
	t.Parallel()
	env := testutil.NewEnv(t)
	ctx := t.Context()

	event := dbfactory.Event(ctx, t, env.Pool, env.Events, nil)
	g1 := dbfactory.Gallery(ctx, t, env.Galleries, event.ID, nil)
	g2 := dbfactory.Gallery(ctx, t, env.Galleries, event.ID, nil)

	rr := doRequest(t, env.Handler, http.MethodGet, "/api/v1/events/"+event.ID+"/galleries", "")

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp server.ListResponse[server.GalleryResponse]
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	require.Len(t, resp.Data, 2)

	ids := map[string]bool{resp.Data[0].ID: true, resp.Data[1].ID: true}
	assert.True(t, ids[g1.ID])
	assert.True(t, ids[g2.ID])
}

func TestHandleCreateGallery(t *testing.T) {
	t.Parallel()
	env := testutil.NewEnv(t)
	ctx := t.Context()

	event := dbfactory.Event(ctx, t, env.Pool, env.Events, nil)

	t.Run("valid", func(t *testing.T) {
		t.Parallel()
		body := `{"name":"Podium Shots","slug":"podium-shots"}`
		rr := doRequest(t, env.Handler, http.MethodPost, "/api/v1/events/"+event.ID+"/galleries", body)

		assert.Equal(t, http.StatusCreated, rr.Code)

		var resp server.ItemResponse[server.GalleryResponse]
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.Equal(t, "Podium Shots", resp.Data.Name)
		assert.Equal(t, "podium-shots", resp.Data.Slug)
		assert.NotEmpty(t, resp.Data.ID)
	})

	t.Run("missing required fields", func(t *testing.T) {
		t.Parallel()
		body := `{"name":"No Slug"}`
		rr := doRequest(t, env.Handler, http.MethodPost, "/api/v1/events/"+event.ID+"/galleries", body)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		t.Parallel()
		rr := doRequest(t, env.Handler, http.MethodPost, "/api/v1/events/"+event.ID+"/galleries", `{bad`)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("event not found", func(t *testing.T) {
		t.Parallel()
		body := `{"name":"Orphan","slug":"orphan"}`
		rr := doRequest(t, env.Handler, http.MethodPost, "/api/v1/events/nonexistent/galleries", body)
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

func TestHandleUpdateGallery(t *testing.T) {
	t.Parallel()
	env := testutil.NewEnv(t)
	ctx := t.Context()

	event := dbfactory.Event(ctx, t, env.Pool, env.Events, nil)
	gal := dbfactory.Gallery(ctx, t, env.Galleries, event.ID, nil)

	t.Run("valid partial update", func(t *testing.T) {
		t.Parallel()
		body := `{"name":"Renamed Gallery"}`
		rr := doRequest(t, env.Handler, http.MethodPatch, "/api/v1/events/"+event.ID+"/galleries/"+gal.ID, body)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp server.ItemResponse[server.GalleryResponse]
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.Equal(t, "Renamed Gallery", resp.Data.Name)
		assert.Equal(t, gal.Slug, resp.Data.Slug)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		body := `{"name":"Nope"}`
		rr := doRequest(t, env.Handler, http.MethodPatch, "/api/v1/events/"+event.ID+"/galleries/nonexistent", body)
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		t.Parallel()
		rr := doRequest(t, env.Handler, http.MethodPatch, "/api/v1/events/"+event.ID+"/galleries/"+gal.ID, `{bad`)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestHandleDeleteGallery(t *testing.T) {
	t.Parallel()
	env := testutil.NewEnv(t)
	ctx := t.Context()

	event := dbfactory.Event(ctx, t, env.Pool, env.Events, nil)
	gal := dbfactory.Gallery(ctx, t, env.Galleries, event.ID, nil)

	rr := doRequest(t, env.Handler, http.MethodDelete, "/api/v1/events/"+event.ID+"/galleries/"+gal.ID, "")
	assert.Equal(t, http.StatusNoContent, rr.Code)

	// Verify the gallery is actually gone.
	rr = doRequest(t, env.Handler, http.MethodGet, "/api/v1/events/"+event.ID+"/galleries", "")
	assert.Equal(t, http.StatusOK, rr.Code)

	var resp server.ListResponse[server.GalleryResponse]
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Empty(t, resp.Data)
}
