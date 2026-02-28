// Package dbtest provides a lightweight test database helper that does not
// depend on the server package, making it safe to import from any test package
// including ones that server itself imports (e.g. middleware).
package dbtest

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/peterldowns/pgtestdb"
	"github.com/peterldowns/pgtestdb/migrators/goosemigrator"

	"github.com/Xevion/motophoto/internal/database"
)

func testDBConf() pgtestdb.Config {
	host := "localhost"
	port := "57512"
	user := "motophoto"
	password := "motophoto"
	dbname := "motophoto"

	if os.Getenv("CI") == "true" {
		port = "5432"
		dbname = "motophoto_test"
	}

	return pgtestdb.Config{
		DriverName: "pgx",
		Host:       host,
		Port:       port,
		User:       user,
		Password:   password,
		Database:   dbname,
		Options:    "sslmode=disable",
	}
}

var migrator = goosemigrator.New(
	"migrations",
	goosemigrator.WithFS(database.MigrationsFS()),
)

// NewTestPool creates an isolated Postgres database for a single test using
// pgtestdb's template-database cloning. Migrations run once and are cached.
// The database is dropped when the test finishes.
func NewTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	conf := pgtestdb.Custom(t, testDBConf(), migrator)

	pool, err := pgxpool.New(context.Background(), conf.URL())
	if err != nil {
		t.Fatalf("dbtest.NewTestPool: create pool: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}
