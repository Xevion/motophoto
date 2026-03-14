package database

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Host returns the host portion of DATABASE_URL for logging.
// Returns "unknown" if the URL is unset or unparseable.
func Host() string {
	raw := os.Getenv("DATABASE_URL")
	if raw == "" {
		return "unknown"
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "unknown"
	}
	return u.Host
}

// New creates a new database connection pool.
func New(ctx context.Context) (*pgxpool.Pool, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return nil, fmt.Errorf("DATABASE_URL is not set -- copy .env.example to .env and configure it")
	}

	// Detect unresolved Railway/shell template syntax that would produce a malformed DSN.
	// Railway uses ${{ }} for variable references; if expansion fails the raw braces end up
	// in the connection string (e.g. database="railway}}") and pgx accepts the malformed name
	// silently until the server rejects it.
	if strings.Contains(dsn, "}}") || strings.Contains(dsn, "${{") {
		return nil, fmt.Errorf("DATABASE_URL contains unresolved template syntax (got %q): check Railway variable configuration", dsn)
	}

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse database config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return pool, nil
}
