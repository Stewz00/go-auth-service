package middleware

import (
	"net/http"

	"github.com/go-chi/httprate"
)

// RateLimiter creates a middleware that limits requests based on IP address
// It allows 100 requests per minute per IP address for regular endpoints
func RateLimiter() func(http.Handler) http.Handler {
	return httprate.LimitByIP(100, 60) // 100 requests per 60 seconds
}

// StrictRateLimiter creates a more restrictive rate limiter for sensitive endpoints
// like login and registration (10 requests per minute per IP)
func StrictRateLimiter() func(http.Handler) http.Handler {
	return httprate.LimitByIP(10, 60) // 10 requests per 60 seconds
}
