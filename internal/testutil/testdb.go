package testutil

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

// testDBConf returns the pgtestdb config using the local docker-compose defaults.
// In CI, the standard Postgres port (5432) is used instead of the docker-compose port.
func testDBConf() pgtestdb.Config {
	host := "localhost"
	port := "57512"
	user := "motophoto"
	password := "motophoto"
	// pgtestdb creates ephemeral template databases; this name is only used for
	// the initial connection, not for test data. CI creates "motophoto_test" but
	// pgtestdb doesn't rely on that matching.
	database := "motophoto"

	// CI uses the standard Postgres port and a different initial database name.
	if os.Getenv("CI") == "true" {
		port = "5432"
		database = "motophoto_test"
	}

	return pgtestdb.Config{
		DriverName: "pgx",
		Host:       host,
		Port:       port,
		User:       user,
		Password:   password,
		Database:   database,
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
		t.Fatalf("testutil.NewTestPool: create pool: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}
