package middleware

import (
	"net/http"
	"strings"

	"github.com/xenptr/go-projects/expense-tracker-api/internal/auth"
)

// Auth returns a middleware that validates JWT Bearer tokens and attaches
// the authenticated userID to the request context.
func Auth(secret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userToken := r.Header.Get("Authorization")

			if !strings.HasPrefix(userToken, "Bearer ") {
				http.Error(w, "missing or invalid authorization header", http.StatusUnauthorized)
				return
			}

			tokenString := strings.TrimPrefix(userToken, "Bearer ")

			userID, err := auth.ParseToken(tokenString, secret)
			if err != nil {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			// Store authenticated userID into request context for downstream handlers
			r = auth.SetUserID(r, userID)

			next.ServeHTTP(w, r)
		})
	}
}
