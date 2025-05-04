package middleware

import "net/http"

type concurrencyLimiter struct {
	sema chan struct{}
}

func newConcurrencyLimiter(limit int) *concurrencyLimiter {
	return &concurrencyLimiter{
		sema: make(chan struct{}, limit),
	}
}

func Concurrency(next http.HandlerFunc, limit int) http.HandlerFunc {
	limiter := newConcurrencyLimiter(limit)

	return func(w http.ResponseWriter, r *http.Request) {
		select {
		case limiter.sema <- struct{}{}:
			defer func() { <-limiter.sema }()
			next.ServeHTTP(w, r)
		default:
			http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		}
	}
}
