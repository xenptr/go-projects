package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xenptr/go-projects/todo-list-api/internal/models"
)

// UserRepository defines the data access contract for user entities.
type UserRepository interface {
	CreateUser(ctx context.Context, u models.User) (int64, error)
	GetUserByID(ctx context.Context, id int64) (models.User, error)
	GetUserByEmail(ctx context.Context, email string) (models.User, error)
}

// TodoRepository defines the data access contract for todo entities.
type TodoRepository interface {
	CreateTodo(ctx context.Context, t models.Todo) (int64, error)
	GetTodoByID(ctx context.Context, id int64) (models.Todo, error)
	ListTodosByUser(ctx context.Context, userID int64, page *int64, limit *int64, search *string) ([]models.Todo, int64, error)
	GetTodoOwner(ctx context.Context, id int64) (int64, error)
	UpdateTodo(ctx context.Context, id int64, t models.Todo) error
	DeleteTodo(ctx context.Context, id int64) error
}

// Store is a composite interface combining all sub-repositories.
type Store interface {
	UserRepository
	TodoRepository
}

// Repo is the concrete PostgreSQL implementation of the Store interface.
type Repo struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repo {
	return &Repo{
		pool: pool,
	}
}
