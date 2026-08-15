package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/xenptr/go-projects/todo-list-api/internal/models"
)

func (r *Repo) CreateTodo(ctx context.Context, t models.Todo) (int64, error) {
	var id int64
	err := r.pool.QueryRow(
		ctx,
		`INSERT INTO todos (user_id, title, description, completed)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id`,
		t.UserID,
		t.Title,
		t.Description,
		t.Completed,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("CreateTodo: %w", err)
	}
	return id, nil
}

func (r *Repo) GetTodoByID(ctx context.Context, id int64) (models.Todo, error) {
	var t models.Todo
	err := r.pool.QueryRow(
		ctx,
		`SELECT id, user_id, title, description, completed, created_at
		 FROM todos WHERE id = $1`,
		id,
	).Scan(&t.ID, &t.UserID, &t.Title, &t.Description, &t.Completed, &t.CreatedAt)
	if err != nil {
		return t, fmt.Errorf("GetTodoByID: %w", err)
	}
	return t, nil
}

func (r *Repo) ListTodosByUser(ctx context.Context, userID int64, page *int64, limit *int64, search *string) ([]models.Todo, int64, error) {
	where := `WHERE user_id = $1`
	args := []any{userID}
	argIndex := 2

	if search != nil {
		where += fmt.Sprintf(` AND (title ILIKE $%d OR description ILIKE $%d)`, argIndex, argIndex)
		args = append(args, "%"+*search+"%")
		argIndex++
	}

	var total int64
	err := r.pool.QueryRow(
		ctx,
		`SELECT COUNT(*) FROM todos `+where,
		args...,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("ListTodosByUser count: %w", err)
	}

	query := `SELECT id, user_id, title, description, completed, created_at FROM todos ` + where + ` ORDER BY created_at DESC`

	if limit != nil {
		query += fmt.Sprintf(` LIMIT $%d`, argIndex)
		args = append(args, *limit)
		argIndex++

		if page != nil {
			offset := (*page - 1) * *limit
			query += fmt.Sprintf(` OFFSET $%d`, argIndex)
			args = append(args, offset)
		}
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("ListTodosByUser query: %w", err)
	}
	defer rows.Close()

	var todos []models.Todo
	for rows.Next() {
		var t models.Todo
		if err := rows.Scan(&t.ID, &t.UserID, &t.Title, &t.Description, &t.Completed, &t.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("ListTodosByUser scan: %w", err)
		}
		todos = append(todos, t)
	}
	if rows.Err() != nil {
		return nil, 0, fmt.Errorf("ListTodosByUser rows: %w", rows.Err())
	}

	if todos == nil {
		todos = []models.Todo{}
	}

	return todos, total, nil
}

func (r *Repo) GetTodoOwner(ctx context.Context, id int64) (int64, error) {
	var ownerID int64
	err := r.pool.QueryRow(
		ctx,
		`SELECT user_id FROM todos WHERE id = $1`,
		id,
	).Scan(&ownerID)
	if err != nil {
		return 0, fmt.Errorf("GetTodoOwner: %w", err)
	}
	return ownerID, nil
}

func (r *Repo) UpdateTodo(ctx context.Context, id int64, t models.Todo) error {
	var setClauses []string
	var args []any
	i := 1

	if strings.TrimSpace(t.Title) != "" {
		setClauses = append(setClauses, fmt.Sprintf("title = $%d", i))
		args = append(args, t.Title)
		i++
	}
	if strings.TrimSpace(t.Description) != "" {
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", i))
		args = append(args, t.Description)
		i++
	}
	setClauses = append(setClauses, fmt.Sprintf("completed = $%d", i))
	args = append(args, t.Completed)
	i++

	setClauses = append(setClauses, "updated_at = CURRENT_TIMESTAMP")
	args = append(args, id)

	query := fmt.Sprintf(
		`UPDATE todos SET %s WHERE id = $%d`,
		strings.Join(setClauses, ", "),
		i,
	)

	ct, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("UpdateTodo: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("todo not found")
	}
	return nil
}

func (r *Repo) DeleteTodo(ctx context.Context, id int64) error {
	ct, err := r.pool.Exec(ctx, "DELETE FROM todos WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("DeleteTodo: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("todo not found")
	}
	return nil
}
