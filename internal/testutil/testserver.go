package testutil

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Xevion/motophoto/internal/server"
	"github.com/Xevion/motophoto/internal/testutil/dbfactory"
)

// NewTestServer creates a Server backed by the given pool, with a minimal
// in-memory session manager. Returns the http.Handler for use with httptest.
func NewTestServer(t *testing.T, pool *pgxpool.Pool) http.Handler {
	t.Helper()

	sessions := scs.New()
	sessions.Lifetime = 24 * time.Hour

	srv, err := server.New(pool, sessions)
	if err != nil {
		t.Fatalf("testutil.NewTestServer: %v", err)
	}
	t.Cleanup(srv.Close)

	return srv.Router()
}

// loginAs creates a user with the given role, logs them in, and returns the
// session cookie and user ID.
func loginAs(t *testing.T, handler http.Handler, pool *pgxpool.Pool, role string) (*http.Cookie, string) {
	t.Helper()
	ctx := t.Context()

	password := "testpassword123"
	email := fmt.Sprintf("%s-%d@test.example", role, time.Now().UnixNano())
	dbfactory.User(ctx, t, pool, &dbfactory.UserOpts{
		Email:    &email,
		Password: &password,
		Role:     &role,
	})

	body := fmt.Sprintf(`{"email":%q,"password":%q}`, email, password)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("testutil.loginAs(%s): login failed with status %d: %s", role, rr.Code, rr.Body.String())
	}

	var resp server.ItemResponse[server.UserResponse]
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("testutil.loginAs(%s): decode response: %v", role, err)
	}

	var sessionCookie *http.Cookie
	for _, c := range rr.Result().Cookies() {
		if c.Name == "session" {
			sessionCookie = c
			break
		}
	}
	if sessionCookie == nil {
		t.Fatalf("testutil.loginAs(%s): no session cookie in login response", role)
	}

	return sessionCookie, resp.Data.ID
}

// LoginPhotographer creates a photographer user, logs them in, and returns the
// session cookie and user ID.
func LoginPhotographer(t *testing.T, handler http.Handler, pool *pgxpool.Pool) (*http.Cookie, string) {
	t.Helper()
	return loginAs(t, handler, pool, "photographer")
}

// LoginCustomer creates a customer user, logs them in, and returns the session
// cookie and user ID.
func LoginCustomer(t *testing.T, handler http.Handler, pool *pgxpool.Pool) (*http.Cookie, string) {
	t.Helper()
	return loginAs(t, handler, pool, "customer")
}
