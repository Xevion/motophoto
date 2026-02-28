package testutil

import (
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Xevion/motophoto/internal/testutil/dbtest"
)

// NewTestPool creates an isolated Postgres database for a single test.
// It delegates to testutil/dbtest, which has no server dependency and can
// therefore be imported from test packages that would otherwise cycle through
// server (e.g. middleware_test).
func NewTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	return dbtest.NewTestPool(t)
}
