package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRateLimit_Burst(t *testing.T) {
	mw := RateLimit(2)
	var hits int
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 2; i++ {
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))
		require.Equal(t, http.StatusOK, rr.Code)
	}

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))
	require.Equal(t, http.StatusTooManyRequests, rr.Code)
	require.Equal(t, 2, hits)
}

func TestRateLimit_Refill(t *testing.T) {
	l := newRateLimiter(2)

	require.True(t, l.Allow("10.0.0.1"))
	require.True(t, l.Allow("10.0.0.1"))
	require.False(t, l.Allow("10.0.0.1"))

	b := l.ips["10.0.0.1"]
	b.tokens = 0
	b.last = time.Now().Add(-time.Minute)

	require.True(t, l.Allow("10.0.0.1"))
}

func TestRateLimit_IPsIndependent(t *testing.T) {
	l := newRateLimiter(1)

	require.True(t, l.Allow("10.0.0.1"))
	require.False(t, l.Allow("10.0.0.1"))
	require.True(t, l.Allow("10.0.0.2"))
}
