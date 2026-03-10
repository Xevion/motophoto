package session

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/alexedwards/scs/pgxstore"
	"github.com/alexedwards/scs/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

// New creates a session manager backed by PostgreSQL.
// The store starts a background goroutine that cleans up expired sessions every 5 minutes.
// When production is true, the Secure flag is set on session cookies.
func New(pool *pgxpool.Pool, production bool) *scs.SessionManager {
	sm := scs.New()
	sm.Store = pgxstore.New(pool)
	sm.Lifetime = 24 * time.Hour
	sm.IdleTimeout = 30 * time.Minute
	sm.Cookie.Name = "session_id"
	sm.Cookie.HttpOnly = true
	sm.Cookie.Secure = production
	sm.Cookie.SameSite = http.SameSiteLaxMode
	sm.ErrorFunc = sessionErrorFunc
	return sm
}

// sessionErrorFunc replaces the default scs error handler to avoid noisy logging
// and double-writes when clients disconnect mid-request. The default handler
// logs via the stdlib log package (bypassing slog) and writes a 500 response
// on top of whatever the handler already wrote, triggering a superfluous
// WriteHeader warning.
func sessionErrorFunc(w http.ResponseWriter, r *http.Request, err error) {
	if r.Context().Err() != nil {
		slog.Debug("session commit skipped, context canceled", "error", err)
		return
	}
	slog.Error("session error", "error", err, "path", r.URL.Path)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}
