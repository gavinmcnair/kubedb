package auth

import (
	"net/http"
)

// Middleware is a simple authentication middleware that checks for a valid token.
func Middleware(token string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract the token from the "Authorization" header
			authHeader := r.Header.Get("Authorization")
			if authHeader != "Bearer "+token {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
			// Call the next handler in the chain
			next.ServeHTTP(w, r)
		})
	}
}

