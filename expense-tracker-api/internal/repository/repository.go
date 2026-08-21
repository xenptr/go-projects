package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xenptr/go-projects/expense-tracker-api/internal/models"
)

type UserRepository interface {
	CreateUser(ctx context.Context, u models.User) (int64, error)
	GetUserByID(ctx context.Context, id int64) (models.User, error)
	GetUserByEmail(ctx context.Context, email string) (models.User, error)
}

type Store interface {
	UserRepository
}

type Repo struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repo {
	return &Repo{
		pool: pool,
	}
}
