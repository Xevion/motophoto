package server_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// doRequest is a small helper to reduce httptest boilerplate.
func doRequest(t *testing.T, handler http.Handler, method, path string, body string) *httptest.ResponseRecorder {
	t.Helper()
	return doRequestWithCookies(t, handler, method, path, body, nil)
}

// doRequestWithCookies is like doRequest but attaches the given cookies to the
// request. Use this for endpoints that require authentication.
func doRequestWithCookies(t *testing.T, handler http.Handler, method, path string, body string, cookies []*http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	var req *http.Request
	if body != "" {
		req = httptest.NewRequestWithContext(t.Context(), method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequestWithContext(t.Context(), method, path, nil)
	}
	for _, c := range cookies {
		req.AddCookie(c)
	}
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr
}
