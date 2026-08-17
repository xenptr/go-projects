package routes

import (
	"net/http"
	"time"

	"github.com/xenptr/go-projects/todo-list-api/internal/handlers"
	"github.com/xenptr/go-projects/todo-list-api/internal/middleware"
	"github.com/xenptr/go-projects/todo-list-api/internal/ratelimit"
)

func RegisterRoutes(mux *http.ServeMux, h *handlers.Handler, secret []byte, rateLimit *ratelimit.Limiter) {
	mux.HandleFunc("GET /", h.Root)

	authRateLimit := middleware.RateLimit(
		rateLimit,
		ratelimit.Config{
			Limit:  5,
			Window: time.Minute,
		},
		"auth",
	)

	todoRateLimit := middleware.RateLimit(
		rateLimit,
		ratelimit.Config{
			Limit:  100,
			Window: time.Minute,
		},
		"todos",
	)

	authMiddleware := middleware.Auth(secret)

	mux.Handle("POST /register", authRateLimit(http.HandlerFunc(h.CreateUser)))
	mux.Handle("POST /login", authRateLimit(http.HandlerFunc(h.Login)))

	protected := func(handler http.Handler) http.Handler {
		return todoRateLimit(authMiddleware(handler))
	}

	mux.Handle("GET /todos", protected(http.HandlerFunc(h.ListTodos)))
	mux.Handle("POST /todos", protected(http.HandlerFunc(h.CreateTodo)))
	mux.Handle("PUT /todos/{id}", protected(http.HandlerFunc(h.UpdateTodo)))
	mux.Handle("DELETE /todos/{id}", protected(http.HandlerFunc(h.DeleteTodo)))
}
