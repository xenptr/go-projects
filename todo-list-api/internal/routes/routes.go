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

	mux.Handle("GET /todos", middleware.Auth(secret)(http.HandlerFunc(h.ListTodos)))
	mux.Handle("POST /todos", middleware.Auth(secret)(http.HandlerFunc(h.CreateTodo)))
	mux.Handle("PUT /todos/{id}", middleware.Auth(secret)(http.HandlerFunc(h.UpdateTodo)))
	mux.Handle("DELETE /todos/{id}", middleware.Auth(secret)(http.HandlerFunc(h.DeleteTodo)))
}
