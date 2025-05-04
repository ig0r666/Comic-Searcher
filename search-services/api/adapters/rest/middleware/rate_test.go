package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiter(t *testing.T) {
	tests := []struct {
		name           string
		rps            int
		requests       int
		expectedStatus int
		shouldAllow    bool
	}{
		{
			name:           "req < rps",
			rps:            10,
			requests:       1,
			expectedStatus: http.StatusOK,
			shouldAllow:    true,
		},
		{
			name:           "req > rps (limiter blocked req)",
			rps:            1,
			requests:       2,
			expectedStatus: http.StatusOK,
			shouldAllow:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				w.WriteHeader(tt.expectedStatus)
			})

			limitedHandler := Rate(handler, tt.rps)

			for i := 0; i < tt.requests; i++ {
				rr := httptest.NewRecorder()
				req := httptest.NewRequest("GET", "/", nil)

				ctx, cancel := context.WithTimeout(req.Context(), 100*time.Millisecond)
				defer cancel()
				req = req.WithContext(ctx)

				limitedHandler.ServeHTTP(rr, req)

				if i == 0 {
					if rr.Code != tt.expectedStatus {
						t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
					}
					if !called {
						t.Error("handler was not called")
					}
				}
				called = false
			}

			if tt.requests > 1 {
				if tt.shouldAllow && !called {
					t.Error("handler should be called")
				}
				if !tt.shouldAllow && called {
					t.Error("handler shouldn't be called")
				}
			}
		})
	}
}
