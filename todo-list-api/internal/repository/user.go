package repository

import (
	"context"
	"fmt"

	"github.com/xenptr/go-projects/todo-list-api/internal/models"
)

func (r *Repo) CreateUser(ctx context.Context, u models.User) (int64, error) {
	var id int64
	err := r.pool.QueryRow(
		ctx,
		`INSERT INTO users (name, email, password_hash)
		 VALUES ($1, $2, $3)
		 RETURNING id`,
		u.Name,
		u.Email,
		u.PasswordHash,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("CreateUser: %w", err)
	}
	return id, nil
}

func (r *Repo) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	var u models.User
	err := r.pool.QueryRow(
		ctx,
		`SELECT id, name, email, password_hash, created_at
		 FROM users WHERE email = $1`,
		email,
	).Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		return u, fmt.Errorf("GetUserByEmail: %w", err)
	}
	return u, nil
}
