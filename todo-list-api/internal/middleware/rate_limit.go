package middleware

import (
	"net/http"

	"github.com/xenptr/go-projects/todo-list-api/internal/ratelimit"
)

func RateLimit(limiter *ratelimit.Limiter, cfg ratelimit.Config, keyPrefix string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := keyPrefix + ":" + r.RemoteAddr

			allowed, err := limiter.Allow(r.Context(), key, cfg)
			if err != nil {
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}

			if !allowed {
				http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
