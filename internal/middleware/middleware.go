package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
)

type contextKey string

const loggerKey contextKey = "logger"

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
