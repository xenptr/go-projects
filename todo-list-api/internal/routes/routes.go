package routes

import (
	"net/http"

	"github.com/xenptr/go-projects/todo-list-api/internal/handlers"
)

func RegisterRoutes(mux *http.ServeMux, h *handlers.Handler) {
	mux.HandleFunc("GET /", h.Root)
}
