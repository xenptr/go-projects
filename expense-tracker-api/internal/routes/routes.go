package routes

import (
	"net/http"
	"time"

	"github.com/xenptr/go-projects/expense-tracker-api/internal/handlers"
	"github.com/xenptr/go-projects/expense-tracker-api/internal/middleware"
	"github.com/xenptr/go-projects/expense-tracker-api/internal/ratelimit"
)

func RegisterRoutes(mux *http.ServeMux, h *handlers.Handler, secret []byte, rateLimit ratelimit.Limiter) {
	mux.HandleFunc("GET /", h.Root)

	authRateLimit := middleware.RateLimit(
		rateLimit,
		ratelimit.Config{
			Limit:  5,
			Window: time.Minute,
		},
		"auth",
	)

	mux.Handle("POST /register", authRateLimit(http.HandlerFunc(h.CreateUser)))
	mux.Handle("POST /login", authRateLimit(http.HandlerFunc(h.Login)))
	mux.Handle("POST /auth/refresh", authRateLimit(http.HandlerFunc(h.RefreshToken)))
}
