package handlers

import "github.com/xenptr/go-projects/todo-list-api/internal/repository"

type Handler struct {
	repo *repository.Repo
}

func New(repo *repository.Repo) *Handler {
	return &Handler{
		repo: repo,
	}
}
