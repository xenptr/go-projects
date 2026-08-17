package handlers

import "github.com/xenptr/go-projects/todo-list-api/internal/repository"

type Handler struct {
	userRepo repository.UserRepository
	todoRepo repository.TodoRepository
	secret   []byte
}

func New(store repository.Store, secret []byte) *Handler {
	return &Handler{
		userRepo: store,
		todoRepo: store,
		secret:   secret,
	}
}

func NewWithRepos(userRepo repository.UserRepository, todoRepo repository.TodoRepository, secret []byte) *Handler {
	return &Handler{
		userRepo: userRepo,
		todoRepo: todoRepo,
		secret:   secret,
	}
}
