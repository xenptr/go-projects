package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/xenptr/go-projects/todo-list-api/internal/models"
)

func (r *Repo) CreateUser(u models.User) (int64, error) {
	row := r.db.QueryRow(
		context.Background(),
		`INSERT INTO users (name, email, password_hash)
			VALUES ($1, $2, $3)
			RETURNING id`,
		u.Name,
		u.Email,
		u.PasswordHash,
	)

	var id int64
	if err := row.Scan(&id); err != nil {
		return 0, fmt.Errorf("CreateUser: %w", err)
	}

	return id, nil
}

func (r *Repo) GetUserByID(id int64) (models.User, error) {
	var user models.User
	row := r.db.QueryRow(
		context.Background(),
		`SELECT
			id 
			name
			email
			password_hash
			created_at
		FROM users WHERE id = $1`, id,
	)

	if err := row.Scan(&user); err != nil {
		return user, fmt.Errorf("GetUserByID: %w", err)
	}

	return user, nil
}

func (r *Repo) GetUserByEmail(email string) (models.User, error) {
	var user models.User
	row := r.db.QueryRow(
		context.Background(),
		`SELECT
			id 
			name
			email
			password_hash
			created_at
		FROM users WHERE email = $1`, email,
	)

	if err := row.Scan(&user); err != nil {
		return user, fmt.Errorf("GetUserByEmail: %w", err)
	}

	return user, nil
}

func (r *Repo) UpdateUser(id int64, u models.User) error {
	setClauses := make([]string, 3)
	args := make([]any, 3)
	i := 1

	if strings.TrimSpace(u.Name) != "" {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", i))
		args = append(args, u.Name)
		i++
	}
	if strings.TrimSpace(u.Email) != "" {
		setClauses = append(setClauses, fmt.Sprintf("email = $%d", i))
		args = append(args, u.Email)
		i++
	}
	if strings.TrimSpace(u.PasswordHash) != "" {
		setClauses = append(setClauses, fmt.Sprintf("email = $%d", i))
		args = append(args, u.PasswordHash)
		i++
	}

	setClauses = append(setClauses, "updated_at = CURRENT_TIMESTAMP")
	args = append(args, id)

	query := fmt.Sprintf(
		`UPDATE users SET %s WHERE id = $%d`,
		strings.Join(setClauses, ", "),
		i,
	)

	ct, err := r.db.Exec(context.Background(), query, args...)
	if err != nil {
		return fmt.Errorf("UpdateUser: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

func (r *Repo) DeleteUser(id int64) error {
	ct, err := r.db.Exec(context.Background(), "DELETE FROM users WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("DeleteUser: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}
