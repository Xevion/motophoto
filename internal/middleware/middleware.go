package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
)

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

// shouldLog reports whether a response status code warrants a log entry.
// All 5xx responses are logged. Meaningful 4xx codes are logged; codes like
// 301/304 that generate high volume and carry little diagnostic value are not.
func shouldLog(status int) bool {
	if status >= 500 {
		return true
	}
	switch status {
	case http.StatusBadRequest, // 400
		http.StatusUnauthorized,        // 401
		http.StatusForbidden,           // 403
		http.StatusNotFound,            // 404
		http.StatusMethodNotAllowed,    // 405
		http.StatusUnprocessableEntity, // 422
		http.StatusTooManyRequests:     // 429
		return true
	}
	return false
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

// RequestLogger is a chi-compatible middleware that logs error responses using
// the default slog logger, producing structured output consistent with the rest
// of the application. Only 5xx responses and meaningful 4xx codes are logged.
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := &wrappedWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(ww, r)

		if !shouldLog(ww.status) {
			return
		}

		slog.Error("request error",
			"method", r.Method,
			"path", r.URL.Path,
			"status", ww.status,
			"duration_ms", time.Since(start).Milliseconds(),
			"request_id", chimiddleware.GetReqID(r.Context()),
			"remote_addr", r.RemoteAddr,
		)
	})
}
