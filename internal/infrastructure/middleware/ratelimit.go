package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/aussenseiter-VsRB/JHIC-BE/internal/infrastructure/response"
)

const (
	maxBuckets    = 10000
	bucketIdleTTL = 5 * time.Minute
)

type ipBucket struct {
	tokens float64
	last   time.Time
}

type rateLimiter struct {
	mu    sync.Mutex
	rate  float64
	burst float64
	ips   map[string]*ipBucket
}

func newRateLimiter(perMinute int) *rateLimiter {
	return &rateLimiter{
		rate:  float64(perMinute) / 60,
		burst: float64(perMinute),
		ips:   make(map[string]*ipBucket),
	}
}

func (l *rateLimiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	b, ok := l.ips[key]
	if !ok {
		l.sweep(now)
		b = &ipBucket{tokens: l.burst, last: now}
		l.ips[key] = b
	}

	elapsed := now.Sub(b.last).Seconds()
	if elapsed > 0 {
		b.tokens = min(b.tokens+elapsed*l.rate, l.burst)
		b.last = now
	}
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

func (l *rateLimiter) sweep(now time.Time) {
	if len(l.ips) <= maxBuckets {
		return
	}
	for k, b := range l.ips {
		if now.Sub(b.last) > bucketIdleTTL {
			delete(l.ips, k)
		}
	}
}

func RateLimit(perMinute int) func(http.Handler) http.Handler {
	limiter := newRateLimiter(perMinute)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !limiter.Allow(clientIP(r)) {
				response.Error(w, http.StatusTooManyRequests, "rate limit exceeded")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
