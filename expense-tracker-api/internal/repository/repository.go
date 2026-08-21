package repository

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface{}

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
