package handlers

import (
	"github.com/xenptr/go-projects/expense-tracker-api/internal/auth"
	"github.com/xenptr/go-projects/expense-tracker-api/internal/repository"
)

type Handler struct {
	userRepo     repository.UserRepository
	refreshStore auth.RefreshTokenStore
	secret       []byte
}

func New(store repository.Store, refreshStore auth.RefreshTokenStore, secret []byte) *Handler {
	return &Handler{
		userRepo:     store,
		refreshStore: refreshStore,
		secret:       secret,
	}
}

func NewWithRepos(userRepo repository.UserRepository, refreshStore auth.RefreshTokenStore, secret []byte) *Handler {
	return &Handler{
		userRepo:     userRepo,
		refreshStore: refreshStore,
		secret:       secret,
	}
}
