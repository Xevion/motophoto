package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alexedwards/scs/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Xevion/motophoto/internal/database/db"
	"github.com/Xevion/motophoto/internal/middleware"
	"github.com/Xevion/motophoto/internal/testutil/dbfactory"
	"github.com/Xevion/motophoto/internal/testutil/dbtest"
)

// newTestAuth creates an Auth middleware backed by a real test database and an
// in-memory session manager.
func newTestAuth(t *testing.T) (*middleware.Auth, *scs.SessionManager, *pgxpool.Pool) {
	t.Helper()
	pool := dbtest.NewTestPool(t)
	sessions := scs.New()
	return middleware.NewAuth(sessions, db.New(pool)), sessions, pool
}

// makeSession runs a handler that stores userID in the session and returns the
// resulting session cookie. The session manager's LoadAndSave middleware must
// wrap the handler so the session token is committed and the cookie is set.
func makeSession(t *testing.T, sessions *scs.SessionManager, userID string) *http.Cookie {
	t.Helper()
	handler := sessions.LoadAndSave(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessions.Put(r.Context(), "user_id", userID)
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	for _, c := range rr.Result().Cookies() {
		if c.Name == "session" {
			return c
		}
	}
	t.Fatal("makeSession: no session cookie in response")
	return nil
}

// doAuthRequest runs a request through LoadAndSave -> handler with an optional
// cookie and returns the recorder.
func doAuthRequest(sessions *scs.SessionManager, handler http.Handler, cookie *http.Cookie) *httptest.ResponseRecorder {
	chain := sessions.LoadAndSave(handler)
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	if cookie != nil {
		req.AddCookie(cookie)
	}
	rr := httptest.NewRecorder()
	chain.ServeHTTP(rr, req)
	return rr
}

// okHandler is a trivial next handler that records a 200 when reached.
var okHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
})

func TestRequireAuth_NoSession(t *testing.T) {
	t.Parallel()
	auth, sessions, _ := newTestAuth(t)

	rr := doAuthRequest(sessions, auth.RequireAuth(okHandler), nil)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestRequireAuth_UnknownUser(t *testing.T) {
	t.Parallel()
	auth, sessions, _ := newTestAuth(t)

	// Session exists but the user_id references a non-existent row.
	cookie := makeSession(t, sessions, "does-not-exist")
	rr := doAuthRequest(sessions, auth.RequireAuth(okHandler), cookie)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestRequireAuth_BannedUser(t *testing.T) {
	t.Parallel()
	auth, sessions, pool := newTestAuth(t)
	ctx := t.Context()

	userID := dbfactory.User(ctx, t, pool, &dbfactory.UserOpts{Banned: true})
	cookie := makeSession(t, sessions, userID)
	rr := doAuthRequest(sessions, auth.RequireAuth(okHandler), cookie)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestRequireAuth_ValidUser(t *testing.T) {
	t.Parallel()
	auth, sessions, pool := newTestAuth(t)
	ctx := t.Context()

	userID := dbfactory.User(ctx, t, pool, nil)
	cookie := makeSession(t, sessions, userID)
	rr := doAuthRequest(sessions, auth.RequireAuth(okHandler), cookie)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestRequireAuth_AttachesUserToContext(t *testing.T) {
	t.Parallel()
	auth, sessions, pool := newTestAuth(t)
	ctx := t.Context()

	userID := dbfactory.User(ctx, t, pool, nil)
	cookie := makeSession(t, sessions, userID)

	var gotUser *db.User
	var gotOK bool
	capture := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUser, gotOK = middleware.UserFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	rr := doAuthRequest(sessions, auth.RequireAuth(capture), cookie)

	require.Equal(t, http.StatusOK, rr.Code)
	require.True(t, gotOK)
	assert.Equal(t, userID, gotUser.ID)
}

func TestRequireRole_CorrectRole(t *testing.T) {
	t.Parallel()
	auth, sessions, pool := newTestAuth(t)
	ctx := t.Context()

	role := string(db.UserRolePhotographer)
	userID := dbfactory.User(ctx, t, pool, &dbfactory.UserOpts{Role: &role})
	cookie := makeSession(t, sessions, userID)

	rr := doAuthRequest(sessions, auth.RequireRole(db.UserRolePhotographer)(okHandler), cookie)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestRequireRole_WrongRole(t *testing.T) {
	t.Parallel()
	auth, sessions, pool := newTestAuth(t)
	ctx := t.Context()

	role := string(db.UserRoleCustomer)
	userID := dbfactory.User(ctx, t, pool, &dbfactory.UserOpts{Role: &role})
	cookie := makeSession(t, sessions, userID)

	rr := doAuthRequest(sessions, auth.RequireRole(db.UserRolePhotographer)(okHandler), cookie)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestUserFromContext_Empty(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	user, ok := middleware.UserFromContext(req.Context())
	assert.False(t, ok)
	assert.Nil(t, user)
}
