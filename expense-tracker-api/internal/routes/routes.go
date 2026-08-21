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

	expenseRateLimit := middleware.RateLimit(
		rateLimit,
		ratelimit.Config{
			Limit:  100,
			Window: time.Minute,
		},
		"expenses",
	)

	authMiddleware := middleware.Auth(secret)

	mux.Handle("POST /register", authRateLimit(http.HandlerFunc(h.CreateUser)))
	mux.Handle("POST /login", authRateLimit(http.HandlerFunc(h.Login)))
	mux.Handle("POST /auth/refresh", authRateLimit(http.HandlerFunc(h.RefreshToken)))

	// Protected expense routes
	protected := func(handler http.Handler) http.Handler {
		return expenseRateLimit(authMiddleware(handler))
	}

	mux.Handle("GET /expenses", protected(http.HandlerFunc(h.ListExpenses)))
	mux.Handle("POST /expenses", protected(http.HandlerFunc(h.CreateExpense)))
	mux.Handle("GET /expenses/{id}", protected(http.HandlerFunc(h.GetExpense)))
	mux.Handle("PUT /expenses/{id}", protected(http.HandlerFunc(h.UpdateExpense)))
	mux.Handle("DELETE /expenses/{id}", protected(http.HandlerFunc(h.DeleteExpense)))
}
