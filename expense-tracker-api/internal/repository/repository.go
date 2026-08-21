package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"github.com/xenptr/go-projects/expense-tracker-api/internal/models"
)

type UserRepository interface {
	CreateUser(ctx context.Context, u models.User) (int64, error)
	GetUserByID(ctx context.Context, id int64) (models.User, error)
	GetUserByEmail(ctx context.Context, email string) (models.User, error)
}

type ExpenseRepository interface {
	CreateExpense(ctx context.Context, userID int64, title string, amount decimal.Decimal, category models.Category, date time.Time, description string) (*models.Expense, error)
	GetExpense(ctx context.Context, id, userID int64) (*models.Expense, error)
	ListExpenses(ctx context.Context, userID int64, filter ExpenseFilter) ([]*models.Expense, error)
	UpdateExpense(ctx context.Context, id, userID int64, title *string, amount *decimal.Decimal, category *models.Category, date *time.Time, description *string) (*models.Expense, error)
	DeleteExpense(ctx context.Context, id, userID int64) error
}

type Store interface {
	UserRepository
	ExpenseRepository
}

type Repo struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repo {
	return &Repo{
		pool: pool,
	}
}
