package middleware

import (
	"net/http"

	"golang.org/x/time/rate"
)

func Rate(next http.HandlerFunc, rps int) http.HandlerFunc {
	rateLimiter := rate.NewLimiter(rate.Limit(rps), 1)

	return func(w http.ResponseWriter, r *http.Request) {
		if err := rateLimiter.Wait(r.Context()); err != nil {
			return
		}
		next.ServeHTTP(w, r)
	}
}
