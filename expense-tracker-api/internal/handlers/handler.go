package handlers

import (
	"github.com/xenptr/go-projects/expense-tracker-api/internal/auth"
	"github.com/xenptr/go-projects/expense-tracker-api/internal/repository"
)

type Handler struct {
	userRepo     repository.UserRepository
	expenseRepo  repository.ExpenseRepository
	refreshStore auth.RefreshTokenStore
	secret       []byte
}

func New(store repository.Store, refreshStore auth.RefreshTokenStore, secret []byte) *Handler {
	return &Handler{
		userRepo:     store,
		expenseRepo:  store,
		refreshStore: refreshStore,
		secret:       secret,
	}
}

func NewWithRepos(userRepo repository.UserRepository, expenseRepo repository.ExpenseRepository, refreshStore auth.RefreshTokenStore, secret []byte) *Handler {
	return &Handler{
		userRepo:     userRepo,
		expenseRepo:  expenseRepo,
		refreshStore: refreshStore,
		secret:       secret,
	}
}
