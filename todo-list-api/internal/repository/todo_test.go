package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/xenptr/go-projects/todo-list-api/internal/models"
)

func createTestUser(t *testing.T, repo *Repo, ctx context.Context) int64 {
	t.Helper()

	uniqueEmail := fmt.Sprintf("todo_test_user_%d@example.com", time.Now().UnixNano())
	user := models.User{
		Name:         "Todo Tester",
		Email:        uniqueEmail,
		PasswordHash: "hashed_pass",
	}

	userID, err := repo.CreateUser(ctx, user)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	t.Cleanup(func() {
		_, _ = repo.pool.Exec(ctx, "DELETE FROM users WHERE id = $1", userID)
	})

	return userID
}

func TestCreateTodo(t *testing.T) {
	repo, ctx := newTestRepo(t)
	userID := createTestUser(t, repo, ctx)

	t.Run("success creates todo", func(t *testing.T) {
		todo := models.Todo{
			UserID:      userID,
			Title:       "Test Create Todo",
			Description: "Test Description",
			Completed:   false,
		}

		id, err := repo.CreateTodo(ctx, todo)
		if err != nil {
			t.Fatalf("CreateTodo() returned error: %v", err)
		}
		if id <= 0 {
			t.Fatalf("expected positive todo ID, got %d", id)
		}
	})

	t.Run("fails on non-existent user fk", func(t *testing.T) {
		todo := models.Todo{
			UserID:      99999999,
			Title:       "Invalid User Todo",
			Description: "Description",
			Completed:   false,
		}

		_, err := repo.CreateTodo(ctx, todo)
		if err == nil {
			t.Fatal("expected foreign key violation error, got nil")
		}
	})
}

func TestGetTodoByID(t *testing.T) {
	repo, ctx := newTestRepo(t)
	userID := createTestUser(t, repo, ctx)

	todo := models.Todo{
		UserID:      userID,
		Title:       "Fetch Todo",
		Description: "Detailed Description",
		Completed:   false,
	}

	id, err := repo.CreateTodo(ctx, todo)
	if err != nil {
		t.Fatalf("CreateTodo() returned error: %v", err)
	}

	t.Run("success retrieves todo", func(t *testing.T) {
		got, err := repo.GetTodoByID(ctx, id)
		if err != nil {
			t.Fatalf("GetTodoByID() returned error: %v", err)
		}

		if got.ID != id {
			t.Errorf("ID = %d, want %d", got.ID, id)
		}
		if got.UserID != userID {
			t.Errorf("UserID = %d, want %d", got.UserID, userID)
		}
		if got.Title != todo.Title {
			t.Errorf("Title = %q, want %q", got.Title, todo.Title)
		}
		if got.Description != todo.Description {
			t.Errorf("Description = %q, want %q", got.Description, todo.Description)
		}
		if got.Completed != todo.Completed {
			t.Errorf("Completed = %v, want %v", got.Completed, todo.Completed)
		}
		if got.CreatedAt.IsZero() {
			t.Error("expected non-zero CreatedAt")
		}
	})

	t.Run("fails on non-existent id", func(t *testing.T) {
		_, err := repo.GetTodoByID(ctx, 99999999)
		if err == nil {
			t.Fatal("expected error for non-existent todo, got nil")
		}
	})
}

func TestGetTodoOwner(t *testing.T) {
	repo, ctx := newTestRepo(t)
	userID := createTestUser(t, repo, ctx)

	id, err := repo.CreateTodo(ctx, models.Todo{
		UserID:      userID,
		Title:       "Owner Todo",
		Description: "Owner Desc",
	})
	if err != nil {
		t.Fatalf("CreateTodo() returned error: %v", err)
	}

	t.Run("success returns owner id", func(t *testing.T) {
		ownerID, err := repo.GetTodoOwner(ctx, id)
		if err != nil {
			t.Fatalf("GetTodoOwner() returned error: %v", err)
		}
		if ownerID != userID {
			t.Errorf("ownerID = %d, want %d", ownerID, userID)
		}
	})

	t.Run("fails on non-existent id", func(t *testing.T) {
		_, err := repo.GetTodoOwner(ctx, 99999999)
		if err == nil {
			t.Fatal("expected error for non-existent todo, got nil")
		}
	})
}

func TestUpdateTodo(t *testing.T) {
	repo, ctx := newTestRepo(t)
	userID := createTestUser(t, repo, ctx)

	id, err := repo.CreateTodo(ctx, models.Todo{
		UserID:      userID,
		Title:       "Initial Title",
		Description: "Initial Description",
		Completed:   false,
	})
	if err != nil {
		t.Fatalf("CreateTodo() returned error: %v", err)
	}

	t.Run("success updates todo", func(t *testing.T) {
		updateInput := models.Todo{
			Title:       "Modified Title",
			Description: "Modified Description",
			Completed:   true,
		}

		if err := repo.UpdateTodo(ctx, id, updateInput); err != nil {
			t.Fatalf("UpdateTodo() returned error: %v", err)
		}

		updated, err := repo.GetTodoByID(ctx, id)
		if err != nil {
			t.Fatalf("GetTodoByID() returned error: %v", err)
		}

		if updated.Title != "Modified Title" {
			t.Errorf("Title = %q, want %q", updated.Title, "Modified Title")
		}
		if updated.Description != "Modified Description" {
			t.Errorf("Description = %q, want %q", updated.Description, "Modified Description")
		}
		if !updated.Completed {
			t.Errorf("Completed = %v, want true", updated.Completed)
		}
	})

	t.Run("fails on non-existent id", func(t *testing.T) {
		err := repo.UpdateTodo(ctx, 99999999, models.Todo{Title: "None"})
		if err == nil {
			t.Fatal("expected error on non-existent todo update, got nil")
		}
	})
}

func TestDeleteTodo(t *testing.T) {
	repo, ctx := newTestRepo(t)
	userID := createTestUser(t, repo, ctx)

	id, err := repo.CreateTodo(ctx, models.Todo{
		UserID:      userID,
		Title:       "To Delete",
		Description: "Will be deleted",
	})
	if err != nil {
		t.Fatalf("CreateTodo() returned error: %v", err)
	}

	t.Run("success deletes todo", func(t *testing.T) {
		if err := repo.DeleteTodo(ctx, id); err != nil {
			t.Fatalf("DeleteTodo() returned error: %v", err)
		}

		_, err := repo.GetTodoByID(ctx, id)
		if err == nil {
			t.Fatal("expected error retrieving deleted todo, got nil")
		}
	})

	t.Run("fails on non-existent id", func(t *testing.T) {
		err := repo.DeleteTodo(ctx, 99999999)
		if err == nil {
			t.Fatal("expected error on non-existent todo delete, got nil")
		}
	})
}

func TestListTodosByUser(t *testing.T) {
	repo, ctx := newTestRepo(t)
	userID := createTestUser(t, repo, ctx)

	for i := 1; i <= 5; i++ {
		_, err := repo.CreateTodo(ctx, models.Todo{
			UserID:      userID,
			Title:       fmt.Sprintf("Item %d Go Guide", i),
			Description: fmt.Sprintf("Description for item %d", i),
			Completed:   i%2 == 0,
		})
		if err != nil {
			t.Fatalf("CreateTodo() returned error: %v", err)
		}
	}

	t.Run("lists all user todos without filters", func(t *testing.T) {
		todos, total, err := repo.ListTodosByUser(ctx, userID, nil, nil, nil)
		if err != nil {
			t.Fatalf("ListTodosByUser() returned error: %v", err)
		}

		if total != 5 {
			t.Errorf("total = %d, want 5", total)
		}
		if len(todos) != 5 {
			t.Errorf("len(todos) = %d, want 5", len(todos))
		}
	})

	t.Run("paginates todos with page and limit", func(t *testing.T) {
		page := int64(2)
		limit := int64(2)

		todos, total, err := repo.ListTodosByUser(ctx, userID, &page, &limit, nil)
		if err != nil {
			t.Fatalf("ListTodosByUser() returned error: %v", err)
		}

		if total != 5 {
			t.Errorf("total = %d, want 5", total)
		}
		if len(todos) != 2 {
			t.Errorf("len(todos) = %d, want 2", len(todos))
		}
	})

	t.Run("filters todos by search term", func(t *testing.T) {
		searchTerm := "item 3"

		todos, total, err := repo.ListTodosByUser(ctx, userID, nil, nil, &searchTerm)
		if err != nil {
			t.Fatalf("ListTodosByUser() returned error: %v", err)
		}

		if total != 1 {
			t.Errorf("total = %d, want 1", total)
		}
		if len(todos) != 1 {
			t.Fatalf("len(todos) = %d, want 1", len(todos))
		}
		if todos[0].Title != "Item 3 Go Guide" {
			t.Errorf("Title = %q, want %q", todos[0].Title, "Item 3 Go Guide")
		}
	})
}
