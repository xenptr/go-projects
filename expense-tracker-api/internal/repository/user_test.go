package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/xenptr/go-projects/expense-tracker-api/internal/config"
	"github.com/xenptr/go-projects/expense-tracker-api/internal/models"
)

func newTestRepo(t *testing.T) (*Repo, context.Context) {
	t.Helper()

	if err := godotenv.Load("../../.env"); err != nil {
		t.Logf("could not load .env: %v", err)
	}

	cfg := config.Load()

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPass, cfg.DBName)

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Skipf("skipping db integration test: %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Skipf("skipping db integration test: %v", err)
	}

	t.Cleanup(func() {
		pool.Close()
	})

	return New(pool), ctx
}

func TestCreateUser(t *testing.T) {
	repo, ctx := newTestRepo(t)

	t.Run("success creates user", func(t *testing.T) {
		email := fmt.Sprintf("create_user_%d@example.com", time.Now().UnixNano())
		u := models.User{
			Name:         "Integration User",
			Email:        email,
			PasswordHash: "hashed_password_123",
		}

		id, err := repo.CreateUser(ctx, u)
		if err != nil {
			t.Fatalf("CreateUser() returned error: %v", err)
		}
		if id <= 0 {
			t.Fatalf("expected positive user ID, got %d", id)
		}

		t.Cleanup(func() {
			_, _ = repo.pool.Exec(ctx, "DELETE FROM users WHERE id = $1", id)
		})
	})

	t.Run("fails on duplicate email", func(t *testing.T) {
		email := fmt.Sprintf("dup_user_%d@example.com", time.Now().UnixNano())
		u := models.User{
			Name:         "Initial User",
			Email:        email,
			PasswordHash: "hashed_password_123",
		}

		id, err := repo.CreateUser(ctx, u)
		if err != nil {
			t.Fatalf("failed to insert initial user: %v", err)
		}
		t.Cleanup(func() {
			_, _ = repo.pool.Exec(ctx, "DELETE FROM users WHERE id = $1", id)
		})

		dup := models.User{
			Name:         "Duplicate User",
			Email:        email,
			PasswordHash: "hashed_password_456",
		}

		_, err = repo.CreateUser(ctx, dup)
		if err == nil {
			t.Fatal("expected error on duplicate email, got nil")
		}
	})
}

func TestGetUserByEmail(t *testing.T) {
	repo, ctx := newTestRepo(t)

	uniqueSuffix := fmt.Sprintf("%d", time.Now().UnixNano())
	email := fmt.Sprintf("get_user_%s@example.com", uniqueSuffix)

	u := models.User{
		Name:         "Search User",
		Email:        email,
		PasswordHash: "hashed_password_456",
	}

	id, err := repo.CreateUser(ctx, u)
	if err != nil {
		t.Fatalf("CreateUser() returned error: %v", err)
	}
	t.Cleanup(func() {
		_, _ = repo.pool.Exec(ctx, "DELETE FROM users WHERE id = $1", id)
	})

	t.Run("success returns user", func(t *testing.T) {
		got, err := repo.GetUserByEmail(ctx, email)
		if err != nil {
			t.Fatalf("GetUserByEmail() returned error: %v", err)
		}

		if got.ID != id {
			t.Errorf("ID = %d, want %d", got.ID, id)
		}
		if got.Name != u.Name {
			t.Errorf("Name = %q, want %q", got.Name, u.Name)
		}
		if got.Email != u.Email {
			t.Errorf("Email = %q, want %q", got.Email, u.Email)
		}
		if got.PasswordHash != u.PasswordHash {
			t.Errorf("PasswordHash = %q, want %q", got.PasswordHash, u.PasswordHash)
		}
		if got.CreatedAt.IsZero() {
			t.Error("expected non-zero CreatedAt")
		}
	})

	t.Run("fails when email does not exist", func(t *testing.T) {
		_, err := repo.GetUserByEmail(ctx, "nonexistent@example.com")
		if err == nil {
			t.Fatal("expected error for non-existent email, got nil")
		}
	})
}

func TestGetUserByID(t *testing.T) {
	repo, ctx := newTestRepo(t)

	email := fmt.Sprintf("get_by_id_%d@example.com", time.Now().UnixNano())
	u := models.User{
		Name:         "ID Lookup User",
		Email:        email,
		PasswordHash: "hashed_password_789",
	}

	id, err := repo.CreateUser(ctx, u)
	if err != nil {
		t.Fatalf("CreateUser() returned error: %v", err)
	}
	t.Cleanup(func() {
		_, _ = repo.pool.Exec(ctx, "DELETE FROM users WHERE id = $1", id)
	})

	t.Run("success returns user by ID", func(t *testing.T) {
		got, err := repo.GetUserByID(ctx, id)
		if err != nil {
			t.Fatalf("GetUserByID() returned error: %v", err)
		}
		if got.ID != id {
			t.Errorf("ID = %d, want %d", got.ID, id)
		}
		if got.Email != email {
			t.Errorf("Email = %q, want %q", got.Email, email)
		}
	})

	t.Run("fails when ID does not exist", func(t *testing.T) {
		_, err := repo.GetUserByID(ctx, -1)
		if err == nil {
			t.Fatal("expected error for non-existent ID, got nil")
		}
	})
}
