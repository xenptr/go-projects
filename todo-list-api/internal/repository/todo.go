package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/xenptr/go-projects/todo-list-api/internal/models"
)

func (r *Repo) CreateTodo(t models.Todo) (int64, error) {
	row := r.db.QueryRow(
		context.Background(),
		`INSERT INTO todos (user_id, title, description, completed)
			VALUES ($1, $2, $3, $4)
			RETURNING id`,
		t.UserID,
		t.Title,
		t.Description,
		t.Completed,
	)

	var id int64
	if err := row.Scan(&id); err != nil {
		return 0, fmt.Errorf("CreateTodo: %w", err)
	}

	return id, nil
}

func (r *Repo) GetTodoByID(id int64) (models.Todo, error) {
	var todo models.Todo
	row := r.db.QueryRow(
		context.Background(),
		`SELECT
			id 
			user_id
			title
			description
			completed
			created_at
		FROM todos WHERE id = $1`, id,
	)

	if err := row.Scan(&todo); err != nil {
		return todo, fmt.Errorf("GetTodoByID: %w", err)
	}

	return todo, nil
}

func (r *Repo) ListTodosByUser(userID int64, page *int64, limit *int64) ([]models.Todo, error) {
	var todos []models.Todo

	query := `
		SELECT
			id,
			user_id,
			title,
			description,
			completed,
			created_at
		FROM todos
		WHERE user_id = $1
	`

	args := []any{userID}

	if limit != nil {
		query += `LIMIT $2`
		args = append(args, *limit)
		if page != nil {
			offset := (*page - 1) * *limit

			query += `OFFSET $3`
			args = append(args, offset)
		}
	}

	rows, err := r.db.Query(context.Background(), query, args...)
	if err != nil {
		return nil, fmt.Errorf("ListTodosByUser query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var todo models.Todo

		if err := rows.Scan(&todo); err != nil {
			return nil, fmt.Errorf("ListTodosByUser scan: %w", err)
		}
		todos = append(todos, todo)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("ListTodosByUser rows: %w", err)
	}

	return todos, nil
}

func (r *Repo) UpdateTodo(id int64, t models.Todo) error {
	setClauses := make([]string, 3)
	args := make([]any, 3)
	i := 1

	if t.UserID == 0 {
		return fmt.Errorf("user_id is required")
	}
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

	ct, err := r.db.Exec(context.Background(), query, args...)
	if err != nil {
		return fmt.Errorf("UpdateTodo: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("todo not found")
	}
	return nil
}

func (r *Repo) DeleteTodo(id int64) error {
	ct, err := r.db.Exec(context.Background(), "DELETE FROM todos WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("DeleteTodo: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("todo not found")
	}
	return nil
}
