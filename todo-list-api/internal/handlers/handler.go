package handlers

import (
	"github.com/xenptr/go-projects/todo-list-api/internal/auth"
	"github.com/xenptr/go-projects/todo-list-api/internal/repository"
)

type Handler struct {
	userRepo     repository.UserRepository
	todoRepo     repository.TodoRepository
	refreshStore auth.RefreshTokenStore
	secret       []byte
}

func New(store repository.Store, refreshStore auth.RefreshTokenStore, secret []byte) *Handler {
	return &Handler{
		userRepo:     store,
		todoRepo:     store,
		refreshStore: refreshStore,
		secret:       secret,
	}
}

func NewWithRepos(userRepo repository.UserRepository, todoRepo repository.TodoRepository, refreshStore auth.RefreshTokenStore, secret []byte) *Handler {
	return &Handler{
		userRepo:     userRepo,
		todoRepo:     todoRepo,
		refreshStore: refreshStore,
		secret:       secret,
	}
}
