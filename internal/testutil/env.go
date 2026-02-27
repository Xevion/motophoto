package testutil

import (
	"net/http"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Xevion/motophoto/internal/database/db"
	"github.com/Xevion/motophoto/internal/service"
)

// Env bundles all test dependencies so each test function needs only a
// single setup call. Pool, queries, services and an HTTP handler are all
// wired together against an isolated pgtestdb instance.
type Env struct {
	Pool      *pgxpool.Pool
	Queries   *db.Queries
	Events    *service.EventService
	Galleries *service.GalleryService
	Handler   http.Handler
}

// NewEnv creates a fully wired test environment backed by an isolated
// Postgres database. The database is dropped when t finishes.
func NewEnv(t *testing.T) *Env {
	t.Helper()

	pool := NewTestPool(t)
	q := db.New(pool)
	handler := NewTestServer(t, pool)

	return &Env{
		Pool:      pool,
		Queries:   q,
		Events:    service.NewEventService(q),
		Galleries: service.NewGalleryService(q),
		Handler:   handler,
	}
}
