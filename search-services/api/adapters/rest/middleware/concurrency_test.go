package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestConcurrencyLimiter(t *testing.T) {
	tests := []struct {
		name           string
		limit          int
		requests       int
		expectedStatus []int
	}{
		{
			name:           "single request with limit 1",
			limit:          1,
			requests:       1,
			expectedStatus: []int{http.StatusOK},
		},
		{
			name:           "two requests with limit 1",
			limit:          1,
			requests:       2,
			expectedStatus: []int{http.StatusOK, http.StatusServiceUnavailable},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(100 * time.Millisecond)
				w.WriteHeader(http.StatusOK)
			})

			limitedHandler := Concurrency(next, tt.limit)
			results := make(chan int, tt.requests)

			for i := 0; i < tt.requests; i++ {
				go func() {
					rr := httptest.NewRecorder()
					req := httptest.NewRequest("GET", "/", nil)
					limitedHandler.ServeHTTP(rr, req)
					results <- rr.Code
				}()
			}

			var gotStatus []int
			for i := 0; i < tt.requests; i++ {
				gotStatus = append(gotStatus, <-results)
			}

			if len(gotStatus) != len(tt.expectedStatus) {
				t.Errorf("expected %d responses, got %d", len(tt.expectedStatus), len(gotStatus))
			}

			hasServiceUnavailable := false
			for _, status := range gotStatus {
				if status == http.StatusServiceUnavailable {
					hasServiceUnavailable = true
					break
				}
			}

			if tt.requests > tt.limit && !hasServiceUnavailable {
				t.Error("expected at least one 503 response when exceeding limit")
			}
		})
	}
}
