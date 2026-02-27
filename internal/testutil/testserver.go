package testutil

import (
	"net/http"
	"testing"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Xevion/motophoto/internal/server"
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
