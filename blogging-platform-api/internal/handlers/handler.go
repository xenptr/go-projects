package handlers

import "github.com/xenptr/go-projects/blogging-platform-api/internal/store"

type Handler struct {
	store store.PostStore
}

func New(store store.PostStore) *Handler {
	return &Handler{store: store}
}
