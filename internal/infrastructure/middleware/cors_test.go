package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCORS_AllowAll(t *testing.T) {
	mw := CORS([]string{"*"})
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	t.Run("simple request without origin", func(t *testing.T) {
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/v1/health", nil))
		require.Equal(t, http.StatusOK, rr.Code)
		require.Equal(t, "*", rr.Header().Get("Access-Control-Allow-Origin"))
		require.Empty(t, rr.Header().Get("Access-Control-Allow-Credentials"))
	})

	t.Run("request with origin reflects it", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
		r.Header.Set("Origin", "http://localhost:5173")
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, r)
		require.Equal(t, "http://localhost:5173", rr.Header().Get("Access-Control-Allow-Origin"))
		require.Equal(t, "true", rr.Header().Get("Access-Control-Allow-Credentials"))
		require.Contains(t, rr.Header().Values("Vary"), "Origin")
	})

	t.Run("preflight returns 204 with CORS headers", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodOptions, "/api/v1/auth/login", nil)
		r.Header.Set("Origin", "http://localhost:5173")
		r.Header.Set("Access-Control-Request-Method", "POST")
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, r)
		require.Equal(t, http.StatusNoContent, rr.Code)
		require.Equal(t, "http://localhost:5173", rr.Header().Get("Access-Control-Allow-Origin"))
		require.Equal(t, "GET, POST, PUT, DELETE, OPTIONS", rr.Header().Get("Access-Control-Allow-Methods"))
		require.Equal(t, "Content-Type, Authorization", rr.Header().Get("Access-Control-Allow-Headers"))
	})
}

func TestCORS_AllowSpecificOrigins(t *testing.T) {
	mw := CORS([]string{"https://admin.jhic.dev", "https://user.jhic.dev"})
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	t.Run("allowed origin is reflected", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
		r.Header.Set("Origin", "https://admin.jhic.dev")
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, r)
		require.Equal(t, "https://admin.jhic.dev", rr.Header().Get("Access-Control-Allow-Origin"))
		require.Equal(t, "true", rr.Header().Get("Access-Control-Allow-Credentials"))
	})

	t.Run("disallowed origin gets no allow-origin", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
		r.Header.Set("Origin", "https://evil.example.com")
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, r)
		require.Empty(t, rr.Header().Get("Access-Control-Allow-Origin"))
		require.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("preflight from disallowed origin is forbidden", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodOptions, "/api/v1/auth/login", nil)
		r.Header.Set("Origin", "https://evil.example.com")
		r.Header.Set("Access-Control-Request-Method", "POST")
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, r)
		require.Equal(t, http.StatusForbidden, rr.Code)
		require.Empty(t, rr.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("no origin gets no allow-origin when restricted", func(t *testing.T) {
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/v1/health", nil))
		require.Empty(t, rr.Header().Get("Access-Control-Allow-Origin"))
	})
}
