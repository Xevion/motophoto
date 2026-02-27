package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/stretchr/testify/assert"

	"github.com/Xevion/motophoto/internal/middleware"
)

// realIPStack builds a minimal handler chain matching the production middleware
// order in server.go: RequestID -> chi.RealIP -> RequestLogger -> echo RemoteAddr.
// The final handler writes the resolved RemoteAddr so tests can assert on it.
func realIPStack() http.Handler {
	final := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Resolved-IP", r.RemoteAddr)
		w.WriteHeader(http.StatusOK)
	})

	// Mirror production order: RequestID -> RealIP -> RequestLogger -> handler
	return middleware.RequestID(chimw.RealIP(middleware.RequestLogger(final)))
}

func TestRealIP_TrueClientIPTakesPriority(t *testing.T) {
	t.Parallel()
	handler := realIPStack()

	// Simulate: Client (203.0.113.50) -> Cloudflare -> Fastly -> SvelteKit -> Go
	// Cloudflare sets True-Client-IP to the actual client.
	// Railway sets X-Real-IP to Cloudflare's edge IP (not the client).
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("True-Client-IP", "203.0.113.50")
	req.Header.Set("X-Real-IP", "162.158.0.1")                               // Cloudflare edge IP
	req.Header.Set("X-Forwarded-For", "203.0.113.50, 162.158.0.1, 10.0.0.1") // full chain

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, "203.0.113.50", rr.Header().Get("X-Resolved-IP"))
}

func TestRealIP_XRealIPUsedWhenNoTrueClientIP(t *testing.T) {
	t.Parallel()
	handler := realIPStack()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Real-IP", "198.51.100.10")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, "198.51.100.10", rr.Header().Get("X-Resolved-IP"))
}

func TestRealIP_XForwardedForFirstEntry(t *testing.T) {
	t.Parallel()
	handler := realIPStack()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.50, 10.0.0.1")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, "203.0.113.50", rr.Header().Get("X-Resolved-IP"))
}

func TestRealIP_FallsBackToRemoteAddr(t *testing.T) {
	t.Parallel()
	handler := realIPStack()

	// No IP headers at all -- should keep the default RemoteAddr from httptest
	// (which is "192.0.2.1:1234").
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Contains(t, rr.Header().Get("X-Resolved-IP"), "192.0.2.1")
}

func TestRealIP_CloudflareFastlyChain(t *testing.T) {
	t.Parallel()
	handler := realIPStack()

	// This is the exact scenario that was broken: Railway/Fastly sets X-Real-IP
	// to Cloudflare's edge, NOT the client. Without True-Client-IP forwarded,
	// chi picks up the wrong IP.
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Real-IP", "162.158.0.1") // Cloudflare edge -- WRONG
	req.Header.Set("X-Forwarded-For", "203.0.113.50, 162.158.0.1")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Without True-Client-IP, chi prefers X-Real-IP (Cloudflare edge).
	// This test documents the current (broken) behavior when True-Client-IP
	// is NOT forwarded by SvelteKit.
	assert.Equal(t, "162.158.0.1", rr.Header().Get("X-Resolved-IP"),
		"without True-Client-IP, chi uses X-Real-IP which contains the wrong IP")
}
