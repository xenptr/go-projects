package middleware

import (
	"net/http"
	"strconv"

	"github.com/xenptr/go-projects/expense-tracker-api/internal/ratelimit"
)

func RateLimit(limiter ratelimit.Limiter, cfg ratelimit.Config, keyPrefix string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := keyPrefix + ":" + r.RemoteAddr

			result, err := limiter.Allow(r.Context(), key, cfg)
			if err != nil {
				// Fail open for now.
				// Log this in a real application.
				next.ServeHTTP(w, r)
				return
			}

			w.Header().Set(
				"X-RateLimit-Limit",
				strconv.Itoa(result.Limit),
			)

			w.Header().Set(
				"X-RateLimit-Remaining",
				strconv.Itoa(result.Remaining),
			)

			if !result.Allowed {
				retryAfter := max(1, int(result.RetryAfter.Seconds()))

				w.Header().Set(
					"Retry-After",
					strconv.Itoa(retryAfter),
				)

				http.Error(
					w,
					"rate limit exceeded",
					http.StatusTooManyRequests,
				)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
