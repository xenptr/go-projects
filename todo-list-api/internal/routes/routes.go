package routes

import (
	"net/http"

	"github.com/xenptr/go-projects/todo-list-api/internal/handlers"
	"github.com/xenptr/go-projects/todo-list-api/internal/middleware"
)

func RegisterRoutes(mux *http.ServeMux, h *handlers.Handler, secret []byte) {
	mux.HandleFunc("GET /", h.Root)

	mux.HandleFunc("POST /register", h.CreateUser)
	mux.HandleFunc("POST /login", h.Login)

	authMiddleware := middleware.Auth(secret)

	mux.Handle("GET /todos", authMiddleware(http.HandlerFunc(h.ListTodos)))
	mux.Handle("POST /todos", authMiddleware(http.HandlerFunc(h.CreateTodo)))
	mux.Handle("PUT /todos/{id}", authMiddleware(http.HandlerFunc(h.UpdateTodo)))
	mux.Handle("DELETE /todos/{id}", authMiddleware(http.HandlerFunc(h.DeleteTodo)))
}
