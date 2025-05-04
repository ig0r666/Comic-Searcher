package middleware

import (
	"net/http"
	"strings"

	"yadro.com/course/api/core"
)

func Auth(next http.HandlerFunc, verifier core.TokenVerifier) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		tokenString := strings.Split(authHeader, " ")
		if len(tokenString) != 2 {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		token := tokenString[1]
		if err := verifier.Verify(token); err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next(w, r)
	}
}
