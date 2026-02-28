package server_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Xevion/motophoto/internal/testutil"
)

// protectedEndpoint is a photographer-only write endpoint used to exercise both
// RequireAuth (no session, stale session, banned) and RequireRole (wrong role).
const protectedEndpoint = "/api/v1/events"

func TestRequireAuth_NoSession(t *testing.T) {
	t.Parallel()
	env := testutil.NewEnv(t)

	rr := doRequest(t, env.Handler, http.MethodPost, protectedEndpoint, "{}")
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestRequireAuth_StaleSession(t *testing.T) {
	t.Parallel()
	env := testutil.NewEnv(t)
	ctx := t.Context()

	session, userID := testutil.LoginPhotographer(t, env.Handler, env.Pool)

	_, err := env.Pool.Exec(ctx, "DELETE FROM users WHERE id = $1", userID)
	require.NoError(t, err)

	rr := doRequestWithCookies(t, env.Handler, http.MethodPost, protectedEndpoint, "{}", []*http.Cookie{session})
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestRequireAuth_BannedUser(t *testing.T) {
	t.Parallel()
	env := testutil.NewEnv(t)
	ctx := t.Context()

	session, userID := testutil.LoginPhotographer(t, env.Handler, env.Pool)

	require.NoError(t, env.Queries.BanUser(ctx, userID))

	rr := doRequestWithCookies(t, env.Handler, http.MethodPost, protectedEndpoint, "{}", []*http.Cookie{session})
	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestRequireRole_CustomerForbidden(t *testing.T) {
	t.Parallel()
	env := testutil.NewEnv(t)

	session, _ := testutil.LoginCustomer(t, env.Handler, env.Pool)

	rr := doRequestWithCookies(t, env.Handler, http.MethodPost, protectedEndpoint, "{}", []*http.Cookie{session})
	assert.Equal(t, http.StatusForbidden, rr.Code)
}
