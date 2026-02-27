package database

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var migrations embed.FS

// MigrationsFS returns the embedded migrations for use in test utilities.
func MigrationsFS() embed.FS { return migrations }

// Migrate runs all pending database migrations against the given pool.
// Migrations are embedded at build time from internal/database/migrations/*.sql.
func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	db := stdlib.OpenDBFromPool(pool)

	// fs.Sub strips the "migrations/" prefix so goose sees .sql files at the root.
	migrationsFS, err := fs.Sub(migrations, "migrations")
	if err != nil {
		return fmt.Errorf("get migrations sub-fs: %w", err)
	}

	provider, err := goose.NewProvider(goose.DialectPostgres, db, migrationsFS)
	if err != nil {
		return fmt.Errorf("create migration provider: %w", err)
	}

	results, err := provider.Up(ctx)
	if err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	for _, r := range results {
		if r.Error != nil {
			return fmt.Errorf("migration %s: %w", r.Source.Path, r.Error)
		}
		slog.Info("migration applied", "version", r.Source.Version, "path", r.Source.Path, "duration", r.Duration)
	}

	return nil
}
