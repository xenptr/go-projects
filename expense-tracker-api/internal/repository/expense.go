package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
	"github.com/xenptr/go-projects/expense-tracker-api/internal/models"
)

// ExpenseFilter holds the optional date range for listing expenses.
type ExpenseFilter struct {
	StartDate *time.Time
	EndDate   *time.Time
}

// CreateExpense inserts a new expense for the given user.
func (r *Repo) CreateExpense(ctx context.Context, userID int64, title string, amount decimal.Decimal, category models.Category, date time.Time, description string) (*models.Expense, error) {
	query := `
		INSERT INTO expenses (user_id, title, amount, category, date, description)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, user_id, title, amount, category, date, description, created_at, updated_at`

	e := &models.Expense{}
	err := r.pool.QueryRow(ctx, query, userID, title, amount, category, date, description).
		Scan(&e.ID, &e.UserID, &e.Title, &e.Amount, &e.Category, &e.Date, &e.Description, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create expense: %w", err)
	}
	return e, nil
}

// GetExpense returns a single expense owned by userID.
func (r *Repo) GetExpense(ctx context.Context, id, userID int64) (*models.Expense, error) {
	query := `
		SELECT id, user_id, title, amount, category, date, description, created_at, updated_at
		FROM expenses
		WHERE id = $1 AND user_id = $2`

	e := &models.Expense{}
	err := r.pool.QueryRow(ctx, query, id, userID).
		Scan(&e.ID, &e.UserID, &e.Title, &e.Amount, &e.Category, &e.Date, &e.Description, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get expense: %w", err)
	}
	return e, nil
}

// ListExpenses returns expenses owned by userID, optionally filtered by date range.
func (r *Repo) ListExpenses(ctx context.Context, userID int64, filter ExpenseFilter) ([]*models.Expense, error) {
	query := `
		SELECT id, user_id, title, amount, category, date, description, created_at, updated_at
		FROM expenses
		WHERE user_id = $1`

	args := []any{userID}
	argIdx := 2

	if filter.StartDate != nil {
		query += fmt.Sprintf(" AND date >= $%d", argIdx)
		args = append(args, filter.StartDate)
		argIdx++
	}
	if filter.EndDate != nil {
		query += fmt.Sprintf(" AND date <= $%d", argIdx)
		args = append(args, filter.EndDate)
		argIdx++
	}

	query += " ORDER BY date DESC, id DESC"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list expenses: %w", err)
	}
	defer rows.Close()

	var expenses []*models.Expense
	for rows.Next() {
		e := &models.Expense{}
		if err := rows.Scan(&e.ID, &e.UserID, &e.Title, &e.Amount, &e.Category, &e.Date, &e.Description, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan expense: %w", err)
		}
		expenses = append(expenses, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list expenses rows: %w", err)
	}
	return expenses, nil
}

// UpdateExpense applies partial updates to an expense owned by userID.
// Only non-nil fields are updated.
func (r *Repo) UpdateExpense(ctx context.Context, id, userID int64, title *string, amount *decimal.Decimal, category *models.Category, date *time.Time, description *string) (*models.Expense, error) {
	// Build a dynamic SET clause.
	setClauses := []string{"updated_at = NOW()"}
	args := []any{}
	argIdx := 1

	if title != nil {
		setClauses = append(setClauses, fmt.Sprintf("title = $%d", argIdx))
		args = append(args, *title)
		argIdx++
	}
	if amount != nil {
		setClauses = append(setClauses, fmt.Sprintf("amount = $%d", argIdx))
		args = append(args, *amount)
		argIdx++
	}
	if category != nil {
		setClauses = append(setClauses, fmt.Sprintf("category = $%d", argIdx))
		args = append(args, *category)
		argIdx++
	}
	if date != nil {
		setClauses = append(setClauses, fmt.Sprintf("date = $%d", argIdx))
		args = append(args, *date)
		argIdx++
	}
	if description != nil {
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", argIdx))
		args = append(args, *description)
		argIdx++
	}

	setStr := ""
	for i, c := range setClauses {
		if i > 0 {
			setStr += ", "
		}
		setStr += c
	}

	query := fmt.Sprintf(`
		UPDATE expenses
		SET %s
		WHERE id = $%d AND user_id = $%d
		RETURNING id, user_id, title, amount, category, date, description, created_at, updated_at`,
		setStr, argIdx, argIdx+1)

	args = append(args, id, userID)

	e := &models.Expense{}
	err := r.pool.QueryRow(ctx, query, args...).
		Scan(&e.ID, &e.UserID, &e.Title, &e.Amount, &e.Category, &e.Date, &e.Description, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("update expense: %w", err)
	}
	return e, nil
}

// DeleteExpense removes an expense owned by userID. Returns ErrNotFound if absent.
func (r *Repo) DeleteExpense(ctx context.Context, id, userID int64) error {
	query := `DELETE FROM expenses WHERE id = $1 AND user_id = $2`

	tag, err := r.pool.Exec(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("delete expense: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
