package middleware

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/Xevion/motophoto/internal/database/db"
)

type contextKey string

const (
	loggerKey  contextKey = "logger"
	userCtxKey contextKey = "user"
)

// Auth holds the session manager and DB queries needed to authenticate requests.
type Auth struct {
	sessions *scs.SessionManager
	queries  *db.Queries
}

// NewAuth creates an Auth middleware with the given session manager and queries.
func NewAuth(sessions *scs.SessionManager, queries *db.Queries) *Auth {
	return &Auth{sessions: sessions, queries: queries}
}

// UserFromContext retrieves the authenticated user stored by RequireAuth.
// Returns nil, false if no user is attached (e.g. on unauthenticated routes).
func UserFromContext(ctx context.Context) (*db.User, bool) {
	user, ok := ctx.Value(userCtxKey).(*db.User)
	return user, ok
}

// RequireAuth rejects unauthenticated requests with 401 and banned users with
// 403. On success it attaches the user to the request context so downstream
// handlers can call UserFromContext.
func (a *Auth) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := a.sessions.GetString(r.Context(), "user_id")
		if userID == "" {
			writeError(w, http.StatusUnauthorized, "authentication required")
			return
		}

		user, err := a.queries.GetUserByID(r.Context(), userID)
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusUnauthorized, "authentication required")
			return
		}
		if err != nil {
			LoggerFromContext(r.Context()).Error("looking up user by id", "error", err)
			writeError(w, http.StatusInternalServerError, "authentication failed")
			return
		}

		if user.BannedAt.Valid {
			writeError(w, http.StatusForbidden, "account is banned")
			return
		}

		ctx := context.WithValue(r.Context(), userCtxKey, &user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole wraps RequireAuth and additionally enforces that the authenticated
// user has the given role. Returns 403 if the role does not match.
func (a *Auth) RequireRole(role db.UserRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return a.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, _ := UserFromContext(r.Context())
			if user.Role != role {
				writeError(w, http.StatusForbidden, "insufficient permissions")
				return
			}
			next.ServeHTTP(w, r)
		}))
	}
}

// writeError writes a JSON error response. Duplicated here to avoid an import
// cycle between the middleware and server packages.
func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(`{"error":"` + msg + `"}`))
}

// LoggerFromContext retrieves the request-scoped logger stored by RequestLogger.
// Falls back to the default slog logger if none is found.
func LoggerFromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(loggerKey).(*slog.Logger); ok && l != nil {
		return l
	}
	return slog.Default()
}

// wrappedWriter captures the status code written by a handler so we can
// inspect it after the fact without consuming the response body.
type wrappedWriter struct {
	http.ResponseWriter
	status int
}

func (w *wrappedWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

// RequestID reads Railway's X-Railway-Request-Id header, falls back to the
// standard X-Request-Id, and generates a UUID if neither is present. The
// resolved ID is stored in the request context using chi's RequestIDKey so
// that chimiddleware.GetReqID continues to work throughout the handler chain.
//
// Railway's edge proxy sets X-Railway-Request-Id (not X-Request-Id), so chi's
// built-in middleware.RequestID would silently ignore it and generate a new ID,
// breaking end-to-end request tracing.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Railway-Request-Id")
		if id == "" {
			id = r.Header.Get("X-Request-Id")
		}
		if id == "" {
			id = uuid.New().String()
		}

		ctx := context.WithValue(r.Context(), chimiddleware.RequestIDKey, id)
		w.Header().Set("X-Request-Id", id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequestLogger is a chi-compatible middleware that logs every response using
// the default slog logger with structured fields. Successful responses are
// logged at Debug, client errors (4xx) at Warn, and server errors (5xx) at
// Error. A request-scoped logger carrying the request_id is stored in context
// so handlers can emit correlated log lines without repeating the field.
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := &wrappedWriter{ResponseWriter: w, status: http.StatusOK}

		reqID := chimiddleware.GetReqID(r.Context())
		logger := slog.Default().With("request_id", reqID)
		ctx := context.WithValue(r.Context(), loggerKey, logger)

		next.ServeHTTP(ww, r.WithContext(ctx))

		attrs := []any{
			"method", r.Method,
			"path", r.URL.Path,
			"status", ww.status,
			"duration_ms", time.Since(start).Milliseconds(),
			"request_id", reqID,
			"remote_addr", r.RemoteAddr,
		}

		switch {
		case ww.status >= 500:
			logger.Error("request error", attrs...)
		case ww.status >= 400:
			logger.Warn("request error", attrs...)
		default:
			logger.Debug("request", attrs...)
		}
	})
}
