package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/xenptr/go-projects/expense-tracker-api/internal/models"
)

// createTestUser creates a throwaway user for expense tests and registers a
// cleanup that deletes it (along with its expenses via ON DELETE CASCADE).
func createTestUser(t *testing.T, repo *Repo, ctx context.Context) int64 {
	t.Helper()

	email := fmt.Sprintf("expense_test_user_%d@example.com", time.Now().UnixNano())
	u := models.User{
		Name:         "Expense Tester",
		Email:        email,
		PasswordHash: "hashed_pass",
	}

	userID, err := repo.CreateUser(ctx, u)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	t.Cleanup(func() {
		_, _ = repo.pool.Exec(ctx, "DELETE FROM users WHERE id = $1", userID)
	})

	return userID
}

// sampleExpense returns ready-to-use values for creating a test expense.
func sampleExpense() (string, decimal.Decimal, models.Category, time.Time, string) {
	title := "Test Expense"
	amount := decimal.NewFromFloat(42.50)
	category := models.CategoryGroceries
	date := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	description := "integration test expense"
	return title, amount, category, date, description
}

// ---- CreateExpense ----

func TestCreateExpense(t *testing.T) {
	repo, ctx := newTestRepo(t)
	userID := createTestUser(t, repo, ctx)

	t.Run("success creates expense", func(t *testing.T) {
		title, amount, category, date, description := sampleExpense()

		e, err := repo.CreateExpense(ctx, userID, title, amount, category, date, description)
		if err != nil {
			t.Fatalf("CreateExpense() returned error: %v", err)
		}
		if e.ID <= 0 {
			t.Fatalf("expected positive expense ID, got %d", e.ID)
		}
		if e.UserID != userID {
			t.Errorf("UserID = %d, want %d", e.UserID, userID)
		}
		if e.Title != title {
			t.Errorf("Title = %q, want %q", e.Title, title)
		}
		if !e.Amount.Equal(amount) {
			t.Errorf("Amount = %s, want %s", e.Amount, amount)
		}
		if e.Category != category {
			t.Errorf("Category = %q, want %q", e.Category, category)
		}
		if !e.Date.Equal(date) {
			t.Errorf("Date = %v, want %v", e.Date, date)
		}
		if e.Description != description {
			t.Errorf("Description = %q, want %q", e.Description, description)
		}
		if e.CreatedAt.IsZero() {
			t.Error("expected non-zero CreatedAt")
		}
	})

	t.Run("fails on non-existent user fk", func(t *testing.T) {
		title, amount, category, date, description := sampleExpense()

		_, err := repo.CreateExpense(ctx, 99999999, title, amount, category, date, description)
		if err == nil {
			t.Fatal("expected foreign key violation error, got nil")
		}
	})
}

// ---- GetExpense ----

func TestGetExpense(t *testing.T) {
	repo, ctx := newTestRepo(t)
	userID := createTestUser(t, repo, ctx)

	title, amount, category, date, description := sampleExpense()
	created, err := repo.CreateExpense(ctx, userID, title, amount, category, date, description)
	if err != nil {
		t.Fatalf("CreateExpense() returned error: %v", err)
	}

	t.Run("success retrieves expense by id and user", func(t *testing.T) {
		got, err := repo.GetExpense(ctx, created.ID, userID)
		if err != nil {
			t.Fatalf("GetExpense() returned error: %v", err)
		}
		if got.ID != created.ID {
			t.Errorf("ID = %d, want %d", got.ID, created.ID)
		}
		if got.Title != title {
			t.Errorf("Title = %q, want %q", got.Title, title)
		}
		if !got.Amount.Equal(amount) {
			t.Errorf("Amount = %s, want %s", got.Amount, amount)
		}
		if got.Category != category {
			t.Errorf("Category = %q, want %q", got.Category, category)
		}
	})

	t.Run("returns ErrNotFound for non-existent id", func(t *testing.T) {
		_, err := repo.GetExpense(ctx, 99999999, userID)
		if err != ErrNotFound {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("returns ErrNotFound when userID does not match", func(t *testing.T) {
		_, err := repo.GetExpense(ctx, created.ID, 99999999)
		if err != ErrNotFound {
			t.Fatalf("expected ErrNotFound for wrong user, got %v", err)
		}
	})
}

// ---- ListExpenses ----

func TestListExpenses(t *testing.T) {
	repo, ctx := newTestRepo(t)
	userID := createTestUser(t, repo, ctx)

	// Create expenses spread across different dates.
	dates := []time.Time{
		time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
	}
	for i, d := range dates {
		_, err := repo.CreateExpense(ctx, userID,
			fmt.Sprintf("Expense %d", i+1),
			decimal.NewFromFloat(float64((i+1)*10)),
			models.CategoryOthers,
			d,
			"",
		)
		if err != nil {
			t.Fatalf("CreateExpense() returned error: %v", err)
		}
	}

	t.Run("lists all expenses for user without filter", func(t *testing.T) {
		expenses, err := repo.ListExpenses(ctx, userID, ExpenseFilter{})
		if err != nil {
			t.Fatalf("ListExpenses() returned error: %v", err)
		}
		if len(expenses) != 3 {
			t.Errorf("len(expenses) = %d, want 3", len(expenses))
		}
	})

	t.Run("filters by start date", func(t *testing.T) {
		start := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
		expenses, err := repo.ListExpenses(ctx, userID, ExpenseFilter{StartDate: &start})
		if err != nil {
			t.Fatalf("ListExpenses() returned error: %v", err)
		}
		if len(expenses) != 2 {
			t.Errorf("len(expenses) = %d, want 2 (Feb and Mar)", len(expenses))
		}
	})

	t.Run("filters by end date", func(t *testing.T) {
		end := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
		expenses, err := repo.ListExpenses(ctx, userID, ExpenseFilter{EndDate: &end})
		if err != nil {
			t.Fatalf("ListExpenses() returned error: %v", err)
		}
		if len(expenses) != 2 {
			t.Errorf("len(expenses) = %d, want 2 (Jan and Feb)", len(expenses))
		}
	})

	t.Run("filters by start and end date range", func(t *testing.T) {
		start := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
		end := time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC)
		expenses, err := repo.ListExpenses(ctx, userID, ExpenseFilter{StartDate: &start, EndDate: &end})
		if err != nil {
			t.Fatalf("ListExpenses() returned error: %v", err)
		}
		if len(expenses) != 1 {
			t.Errorf("len(expenses) = %d, want 1 (Feb only)", len(expenses))
		}
		if expenses[0].Title != "Expense 2" {
			t.Errorf("Title = %q, want %q", expenses[0].Title, "Expense 2")
		}
	})

	t.Run("returns empty slice for user with no expenses", func(t *testing.T) {
		otherUser := createTestUser(t, repo, ctx)
		expenses, err := repo.ListExpenses(ctx, otherUser, ExpenseFilter{})
		if err != nil {
			t.Fatalf("ListExpenses() returned error: %v", err)
		}
		if len(expenses) != 0 {
			t.Errorf("len(expenses) = %d, want 0", len(expenses))
		}
	})

	t.Run("does not return expenses belonging to other users", func(t *testing.T) {
		otherUser := createTestUser(t, repo, ctx)
		_, _ = repo.CreateExpense(ctx, otherUser, "Other user expense", decimal.NewFromFloat(5), models.CategoryOthers, time.Now(), "")

		expenses, err := repo.ListExpenses(ctx, userID, ExpenseFilter{})
		if err != nil {
			t.Fatalf("ListExpenses() returned error: %v", err)
		}
		// Should still be exactly 3 (the original user's expenses).
		if len(expenses) != 3 {
			t.Errorf("len(expenses) = %d, want 3 (no cross-user leak)", len(expenses))
		}
	})
}

// ---- UpdateExpense ----

func TestUpdateExpense(t *testing.T) {
	repo, ctx := newTestRepo(t)
	userID := createTestUser(t, repo, ctx)

	title, amount, category, date, description := sampleExpense()
	created, err := repo.CreateExpense(ctx, userID, title, amount, category, date, description)
	if err != nil {
		t.Fatalf("CreateExpense() returned error: %v", err)
	}

	t.Run("success updates title only", func(t *testing.T) {
		newTitle := "Updated Title"
		updated, err := repo.UpdateExpense(ctx, created.ID, userID, &newTitle, nil, nil, nil, nil)
		if err != nil {
			t.Fatalf("UpdateExpense() returned error: %v", err)
		}
		if updated.Title != newTitle {
			t.Errorf("Title = %q, want %q", updated.Title, newTitle)
		}
		// Other fields should remain unchanged.
		if !updated.Amount.Equal(amount) {
			t.Errorf("Amount changed unexpectedly: %s", updated.Amount)
		}
	})

	t.Run("success updates amount only", func(t *testing.T) {
		newAmount := decimal.NewFromFloat(99.99)
		updated, err := repo.UpdateExpense(ctx, created.ID, userID, nil, &newAmount, nil, nil, nil)
		if err != nil {
			t.Fatalf("UpdateExpense() returned error: %v", err)
		}
		if !updated.Amount.Equal(newAmount) {
			t.Errorf("Amount = %s, want %s", updated.Amount, newAmount)
		}
	})

	t.Run("success updates category", func(t *testing.T) {
		newCategory := models.CategoryHealth
		updated, err := repo.UpdateExpense(ctx, created.ID, userID, nil, nil, &newCategory, nil, nil)
		if err != nil {
			t.Fatalf("UpdateExpense() returned error: %v", err)
		}
		if updated.Category != newCategory {
			t.Errorf("Category = %q, want %q", updated.Category, newCategory)
		}
	})

	t.Run("success updates date", func(t *testing.T) {
		newDate := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
		updated, err := repo.UpdateExpense(ctx, created.ID, userID, nil, nil, nil, &newDate, nil)
		if err != nil {
			t.Fatalf("UpdateExpense() returned error: %v", err)
		}
		if !updated.Date.Equal(newDate) {
			t.Errorf("Date = %v, want %v", updated.Date, newDate)
		}
	})

	t.Run("success updates description", func(t *testing.T) {
		newDesc := "Updated description"
		updated, err := repo.UpdateExpense(ctx, created.ID, userID, nil, nil, nil, nil, &newDesc)
		if err != nil {
			t.Fatalf("UpdateExpense() returned error: %v", err)
		}
		if updated.Description != newDesc {
			t.Errorf("Description = %q, want %q", updated.Description, newDesc)
		}
	})

	t.Run("returns ErrNotFound for non-existent id", func(t *testing.T) {
		newTitle := "Ghost"
		_, err := repo.UpdateExpense(ctx, 99999999, userID, &newTitle, nil, nil, nil, nil)
		if err != ErrNotFound {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("returns ErrNotFound when userID does not match", func(t *testing.T) {
		newTitle := "Hijack"
		_, err := repo.UpdateExpense(ctx, created.ID, 99999999, &newTitle, nil, nil, nil, nil)
		if err != ErrNotFound {
			t.Fatalf("expected ErrNotFound for wrong user, got %v", err)
		}
	})
}

// ---- DeleteExpense ----

func TestDeleteExpense(t *testing.T) {
	repo, ctx := newTestRepo(t)
	userID := createTestUser(t, repo, ctx)

	t.Run("success deletes expense", func(t *testing.T) {
		title, amount, category, date, description := sampleExpense()
		created, err := repo.CreateExpense(ctx, userID, title, amount, category, date, description)
		if err != nil {
			t.Fatalf("CreateExpense() returned error: %v", err)
		}

		if err := repo.DeleteExpense(ctx, created.ID, userID); err != nil {
			t.Fatalf("DeleteExpense() returned error: %v", err)
		}

		_, err = repo.GetExpense(ctx, created.ID, userID)
		if err != ErrNotFound {
			t.Fatalf("expected ErrNotFound after delete, got %v", err)
		}
	})

	t.Run("returns ErrNotFound for non-existent id", func(t *testing.T) {
		err := repo.DeleteExpense(ctx, 99999999, userID)
		if err != ErrNotFound {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("returns ErrNotFound when userID does not match", func(t *testing.T) {
		title, amount, category, date, description := sampleExpense()
		created, err := repo.CreateExpense(ctx, userID, title, amount, category, date, description)
		if err != nil {
			t.Fatalf("CreateExpense() returned error: %v", err)
		}

		err = repo.DeleteExpense(ctx, created.ID, 99999999)
		if err != ErrNotFound {
			t.Fatalf("expected ErrNotFound for wrong user, got %v", err)
		}

		// Original expense should still exist.
		_, err = repo.GetExpense(ctx, created.ID, userID)
		if err != nil {
			t.Errorf("expense should still exist after failed delete by wrong user: %v", err)
		}
	})
}
