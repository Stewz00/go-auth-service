package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiter(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	limiter := StrictRateLimiter()(handler) // Use StrictRateLimiter which has lower limits

	tests := []struct {
		name           string
		requests       int
		wantStatusCode int
	}{
		{
			name:           "within limit",
			requests:       5,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "exceed limit",
			requests:       15,
			wantStatusCode: http.StatusTooManyRequests,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean state for each test
			time.Sleep(1 * time.Second)

			var lastStatus int
			for i := 0; i < tt.requests; i++ {
				req := httptest.NewRequest("GET", "/test", nil)
				req.RemoteAddr = "127.0.0.1:12345"
				w := httptest.NewRecorder()

				limiter.ServeHTTP(w, req)
				lastStatus = w.Code

				// Small delay between requests to simulate real traffic
				time.Sleep(10 * time.Millisecond)
			}

			if lastStatus != tt.wantStatusCode {
				t.Errorf("got status %v, want %v", lastStatus, tt.wantStatusCode)
			}
		})
	}
}
