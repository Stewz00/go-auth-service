package middleware

import (
	"net/http"
	"sync"
	"time"
)

type visitor struct {
	count      int
	lastAccess time.Time
}

type rateLimiter struct {
	sync.RWMutex
	visitors  map[string]*visitor
	limit     int
	timeframe time.Duration
}

func newRateLimiter(limit int, timeframe time.Duration) *rateLimiter {
	return &rateLimiter{
		visitors:  make(map[string]*visitor),
		limit:     limit,
		timeframe: timeframe,
	}
}

func (rl *rateLimiter) isAllowed(ip string) bool {
	rl.Lock()
	defer rl.Unlock()

	now := time.Now()
	v, exists := rl.visitors[ip]

	if !exists {
		rl.visitors[ip] = &visitor{1, now}
		return true
	}

	// Reset if timeframe has passed
	if now.Sub(v.lastAccess) > rl.timeframe {
		v.count = 1
		v.lastAccess = now
		return true
	}

	if v.count >= rl.limit {
		return false
	}

	v.count++
	v.lastAccess = now
	return true
}

// RateLimiter creates a middleware that limits requests based on IP address
// It allows 100 requests per minute per IP address for regular endpoints
func RateLimiter() func(http.Handler) http.Handler {
	rl := newRateLimiter(100, time.Minute)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr
			if !rl.isAllowed(ip) {
				http.Error(w, "Too many requests", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// StrictRateLimiter creates a more restrictive rate limiter for sensitive endpoints
// like login and registration (10 requests per minute per IP)
func StrictRateLimiter() func(http.Handler) http.Handler {
	rl := newRateLimiter(10, time.Minute)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr
			if !rl.isAllowed(ip) {
				http.Error(w, "Too many requests", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
