package handlers

import "net/http"

func RegisterRoutes(mux *http.ServeMux, h *Handler) {
	mux.HandleFunc("GET /", h.Root)

	mux.HandleFunc("GET /posts", h.ListPosts)
	mux.HandleFunc("POST /posts", h.CreatePosts)

	mux.HandleFunc("GET /posts/{id}", h.GetPost)
	mux.HandleFunc("PUT /posts/{id}", h.UpdatePost)
	mux.HandleFunc("DELETE /posts/{id}", h.DeletePost)
}
