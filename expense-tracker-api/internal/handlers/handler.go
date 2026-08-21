package handlers

import "github.com/xenptr/go-projects/expense-tracker-api/internal/repository"

type Handler struct {
	userRepo repository.UserRepository
}

func New(store repository.Store) *Handler {
	return &Handler{
		userRepo: store,
	}
}
