package handlers

import "github.com/xenptr/go-projects/todo-list-api/internal/repository"

type Handler struct {
	repo   *repository.Repo
	secret []byte
}

func New(repo *repository.Repo, secret []byte) *Handler {
	return &Handler{
		repo:   repo,
		secret: secret,
	}
}
