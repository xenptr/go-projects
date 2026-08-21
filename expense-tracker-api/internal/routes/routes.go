package routes

import (
	"net/http"

	"github.com/xenptr/go-projects/expense-tracker-api/internal/handlers"
)

func RegisterRoutes(mux *http.ServeMux, h *handlers.Handler) {
	mux.HandleFunc("GET /", h.Root)
}
