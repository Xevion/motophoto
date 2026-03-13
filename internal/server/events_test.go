package server_test

import (
	"encoding/json"
	"fmt"
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
	return doRequestWithCookies(t, handler, method, path, body, nil)
}

// doRequestWithCookies is like doRequest but attaches the given cookies to the
// request. Use this for endpoints that require authentication.
func doRequestWithCookies(t *testing.T, handler http.Handler, method, path string, body string, cookies []*http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	var req *http.Request
	if body != "" {
		req = httptest.NewRequestWithContext(t.Context(), method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequestWithContext(t.Context(), method, path, nil)
	}
	for _, c := range cookies {
		req.AddCookie(c)
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

	session, _ := testutil.LoginPhotographer(t, env.Handler, env.Pool)

	t.Run("valid", func(t *testing.T) {
		t.Parallel()
		body := `{"name":"New Event","slug":"new-event","sport":"motocross"}`
		rr := doRequestWithCookies(t, env.Handler, http.MethodPost, "/api/v1/events", body, []*http.Cookie{session})

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
		rr := doRequestWithCookies(t, env.Handler, http.MethodPost, "/api/v1/events", body, []*http.Cookie{session})
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		t.Parallel()
		rr := doRequestWithCookies(t, env.Handler, http.MethodPost, "/api/v1/events", `{bad`, []*http.Cookie{session})
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestHandleUpdateEvent(t *testing.T) {
	t.Parallel()
	env := testutil.NewEnv(t)
	ctx := t.Context()

	event := dbfactory.Event(ctx, t, env.Pool, env.Events, nil)
	session, _ := testutil.LoginPhotographer(t, env.Handler, env.Pool)

	t.Run("valid partial update", func(t *testing.T) {
		t.Parallel()
		body := `{"name":"Updated Via API"}`
		rr := doRequestWithCookies(t, env.Handler, http.MethodPatch, "/api/v1/events/"+event.ID, body, []*http.Cookie{session})

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp server.ItemResponse[server.EventResponse]
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.Equal(t, "Updated Via API", resp.Data.Name)
		assert.Equal(t, event.Slug, resp.Data.Slug) // unchanged
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		body := `{"name":"Nope"}`
		rr := doRequestWithCookies(t, env.Handler, http.MethodPatch, "/api/v1/events/nonexistent", body, []*http.Cookie{session})
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		t.Parallel()
		rr := doRequestWithCookies(t, env.Handler, http.MethodPatch, "/api/v1/events/"+event.ID, `{bad`, []*http.Cookie{session})
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestHandleDeleteEvent(t *testing.T) {
	t.Parallel()
	env := testutil.NewEnv(t)
	ctx := t.Context()

	event := dbfactory.Event(ctx, t, env.Pool, env.Events, nil)
	session, _ := testutil.LoginPhotographer(t, env.Handler, env.Pool)

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		rr := doRequestWithCookies(t, env.Handler, http.MethodDelete, "/api/v1/events/"+event.ID, "", []*http.Cookie{session})
		assert.Equal(t, http.StatusNoContent, rr.Code)

		// Verify the event is actually gone.
		rr = doRequest(t, env.Handler, http.MethodGet, "/api/v1/events/"+event.ID, "")
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("nonexistent is idempotent", func(t *testing.T) {
		t.Parallel()
		rr := doRequestWithCookies(t, env.Handler, http.MethodDelete, "/api/v1/events/nonexistent", "", []*http.Cookie{session})
		assert.Equal(t, http.StatusNoContent, rr.Code)
	})
}

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
	session, _ := testutil.LoginPhotographer(t, env.Handler, env.Pool)

	t.Run("valid", func(t *testing.T) {
		t.Parallel()
		body := `{"name":"Podium Shots","slug":"podium-shots"}`
		rr := doRequestWithCookies(t, env.Handler, http.MethodPost, "/api/v1/events/"+event.ID+"/galleries", body, []*http.Cookie{session})

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
		rr := doRequestWithCookies(t, env.Handler, http.MethodPost, "/api/v1/events/"+event.ID+"/galleries", body, []*http.Cookie{session})
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		t.Parallel()
		rr := doRequestWithCookies(t, env.Handler, http.MethodPost, "/api/v1/events/"+event.ID+"/galleries", `{bad`, []*http.Cookie{session})
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("event not found", func(t *testing.T) {
		t.Parallel()
		body := `{"name":"Orphan","slug":"orphan"}`
		rr := doRequestWithCookies(t, env.Handler, http.MethodPost, "/api/v1/events/nonexistent/galleries", body, []*http.Cookie{session})
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

func TestHandleUpdateGallery(t *testing.T) {
	t.Parallel()
	env := testutil.NewEnv(t)
	ctx := t.Context()

	event := dbfactory.Event(ctx, t, env.Pool, env.Events, nil)
	gal := dbfactory.Gallery(ctx, t, env.Galleries, event.ID, nil)
	session, _ := testutil.LoginPhotographer(t, env.Handler, env.Pool)

	t.Run("valid partial update", func(t *testing.T) {
		t.Parallel()
		body := `{"name":"Renamed Gallery"}`
		rr := doRequestWithCookies(t, env.Handler, http.MethodPatch, "/api/v1/events/"+event.ID+"/galleries/"+gal.ID, body, []*http.Cookie{session})

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp server.ItemResponse[server.GalleryResponse]
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.Equal(t, "Renamed Gallery", resp.Data.Name)
		assert.Equal(t, gal.Slug, resp.Data.Slug)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		body := `{"name":"Nope"}`
		rr := doRequestWithCookies(t, env.Handler, http.MethodPatch, "/api/v1/events/"+event.ID+"/galleries/nonexistent", body, []*http.Cookie{session})
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		t.Parallel()
		rr := doRequestWithCookies(t, env.Handler, http.MethodPatch, "/api/v1/events/"+event.ID+"/galleries/"+gal.ID, `{bad`, []*http.Cookie{session})
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestHandleDeleteGallery(t *testing.T) {
	t.Parallel()
	env := testutil.NewEnv(t)
	ctx := t.Context()

	event := dbfactory.Event(ctx, t, env.Pool, env.Events, nil)
	gal := dbfactory.Gallery(ctx, t, env.Galleries, event.ID, nil)
	session, _ := testutil.LoginPhotographer(t, env.Handler, env.Pool)

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		rr := doRequestWithCookies(t, env.Handler, http.MethodDelete, "/api/v1/events/"+event.ID+"/galleries/"+gal.ID, "", []*http.Cookie{session})
		assert.Equal(t, http.StatusNoContent, rr.Code)

		// Verify the gallery is actually gone (list galleries is public).
		rr = doRequest(t, env.Handler, http.MethodGet, "/api/v1/events/"+event.ID+"/galleries", "")
		assert.Equal(t, http.StatusOK, rr.Code)

		var resp server.ListResponse[server.GalleryResponse]
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.Empty(t, resp.Data)
	})

	t.Run("nonexistent is idempotent", func(t *testing.T) {
		t.Parallel()
		rr := doRequestWithCookies(t, env.Handler, http.MethodDelete, "/api/v1/events/"+event.ID+"/galleries/nonexistent", "", []*http.Cookie{session})
		assert.Equal(t, http.StatusNoContent, rr.Code)
	})
}

func TestHealthEndpoint(t *testing.T) {
	t.Parallel()
	env := testutil.NewEnv(t)

	rr := doRequest(t, env.Handler, http.MethodGet, "/api/health", "")

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), `"status":"ok"`)
}

func TestHandleListGalleries_EventNotFound(t *testing.T) {
	t.Parallel()
	env := testutil.NewEnv(t)

	rr := doRequest(t, env.Handler, http.MethodGet, "/api/v1/events/nonexistent/galleries", "")
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestHandleListEvents_Pagination(t *testing.T) {
	t.Parallel()
	env := testutil.NewEnv(t)
	ctx := t.Context()

	for i := range 5 {
		dbfactory.Event(ctx, t, env.Pool, env.Events, &dbfactory.EventOpts{
			Status: new("published"),
			Name:   new(fmt.Sprintf("Page Event %d", i)),
			Slug:   new(fmt.Sprintf("page-event-%d", i)),
		})
	}

	// First page: limit=2
	rr := doRequest(t, env.Handler, http.MethodGet, "/api/v1/events?limit=2", "")
	require.Equal(t, http.StatusOK, rr.Code)

	var page1 server.ListResponse[server.EventResponse]
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &page1))
	assert.Len(t, page1.Data, 2)
	require.NotNil(t, page1.NextCursor, "expected NextCursor on first page")

	// Second page using cursor
	rr = doRequest(t, env.Handler, http.MethodGet, "/api/v1/events?limit=2&cursor="+*page1.NextCursor, "")
	require.Equal(t, http.StatusOK, rr.Code)

	var page2 server.ListResponse[server.EventResponse]
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &page2))
	assert.Len(t, page2.Data, 2)

	// Verify no duplicate IDs across pages.
	seen := map[string]bool{}
	for _, e := range page1.Data {
		seen[e.ID] = true
	}
	for _, e := range page2.Data {
		assert.False(t, seen[e.ID], "duplicate ID %s across pages", e.ID)
	}
}

func TestHandleCreateEvent_WithOptionalFields(t *testing.T) {
	t.Parallel()
	env := testutil.NewEnv(t)
	session, _ := testutil.LoginPhotographer(t, env.Handler, env.Pool)

	body := `{
		"name":"Full Event",
		"slug":"full-event",
		"sport":"motocross",
		"location":"Austin, TX",
		"description":"A great event",
		"date":"2026-06-15"
	}`
	rr := doRequestWithCookies(t, env.Handler, http.MethodPost, "/api/v1/events", body, []*http.Cookie{session})

	require.Equal(t, http.StatusCreated, rr.Code)

	var resp server.ItemResponse[server.EventResponse]
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	require.NotNil(t, resp.Data.Location, "Location should be set")
	assert.Equal(t, "Austin, TX", *resp.Data.Location)
	require.NotNil(t, resp.Data.Description, "Description should be set")
	assert.Equal(t, "A great event", *resp.Data.Description)
	require.NotNil(t, resp.Data.Date, "Date should be set")
	assert.Equal(t, "2026-06-15", *resp.Data.Date)
}
