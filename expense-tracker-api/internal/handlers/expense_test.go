package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/xenptr/go-projects/expense-tracker-api/internal/auth"
	"github.com/xenptr/go-projects/expense-tracker-api/internal/models"
	"github.com/xenptr/go-projects/expense-tracker-api/internal/repository"
)

// ---- mock expense repository ----

type mockExpenseRepo struct {
	createExpenseFunc func(ctx context.Context, userID int64, title string, amount decimal.Decimal, category models.Category, date time.Time, description string) (*models.Expense, error)
	getExpenseFunc    func(ctx context.Context, id, userID int64) (*models.Expense, error)
	listExpensesFunc  func(ctx context.Context, userID int64, filter repository.ExpenseFilter) ([]*models.Expense, error)
	updateExpenseFunc func(ctx context.Context, id, userID int64, title *string, amount *decimal.Decimal, category *models.Category, date *time.Time, description *string) (*models.Expense, error)
	deleteExpenseFunc func(ctx context.Context, id, userID int64) error
}

func (m *mockExpenseRepo) CreateExpense(ctx context.Context, userID int64, title string, amount decimal.Decimal, category models.Category, date time.Time, description string) (*models.Expense, error) {
	if m.createExpenseFunc != nil {
		return m.createExpenseFunc(ctx, userID, title, amount, category, date, description)
	}
	return &models.Expense{ID: 1, UserID: userID, Title: title, Amount: amount, Category: category, Date: date, Description: description}, nil
}

func (m *mockExpenseRepo) GetExpense(ctx context.Context, id, userID int64) (*models.Expense, error) {
	if m.getExpenseFunc != nil {
		return m.getExpenseFunc(ctx, id, userID)
	}
	return &models.Expense{ID: id, UserID: userID, Title: "Sample", Amount: decimal.NewFromFloat(10.0), Category: models.CategoryOthers}, nil
}

func (m *mockExpenseRepo) ListExpenses(ctx context.Context, userID int64, filter repository.ExpenseFilter) ([]*models.Expense, error) {
	if m.listExpensesFunc != nil {
		return m.listExpensesFunc(ctx, userID, filter)
	}
	return []*models.Expense{}, nil
}

func (m *mockExpenseRepo) UpdateExpense(ctx context.Context, id, userID int64, title *string, amount *decimal.Decimal, category *models.Category, date *time.Time, description *string) (*models.Expense, error) {
	if m.updateExpenseFunc != nil {
		return m.updateExpenseFunc(ctx, id, userID, title, amount, category, date, description)
	}
	return &models.Expense{ID: id, UserID: userID}, nil
}

func (m *mockExpenseRepo) DeleteExpense(ctx context.Context, id, userID int64) error {
	if m.deleteExpenseFunc != nil {
		return m.deleteExpenseFunc(ctx, id, userID)
	}
	return nil
}

// ---- CreateExpense tests ----

func TestCreateExpense(t *testing.T) {
	secret := []byte("test-secret")

	tests := []struct {
		name           string
		setAuth        bool
		userID         int64
		body           any
		rawBody        string
		repo           *mockExpenseRepo
		expectedStatus int
	}{
		{
			name:           "unauthorized when no user id",
			setAuth:        false,
			body:           map[string]any{"title": "Coffee", "amount": "3.50", "date": "2026-01-01"},
			repo:           &mockExpenseRepo{},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "bad request on invalid json",
			setAuth:        true,
			userID:         1,
			rawBody:        "{invalid-json",
			repo:           &mockExpenseRepo{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:    "unprocessable entity on missing title",
			setAuth: true,
			userID:  1,
			body:    map[string]any{"amount": "3.50", "date": "2026-01-01"},
			repo:    &mockExpenseRepo{},
			// validation.ValidateCreateExpense returns plain error → writeError with 422
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name:    "unprocessable entity on invalid amount",
			setAuth: true,
			userID:  1,
			body:    map[string]any{"title": "Coffee", "amount": "-5.00", "date": "2026-01-01"},
			repo:    &mockExpenseRepo{},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name:    "unprocessable entity on invalid date format",
			setAuth: true,
			userID:  1,
			body:    map[string]any{"title": "Coffee", "amount": "3.50", "date": "01-01-2026"},
			repo:    &mockExpenseRepo{},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name:    "unprocessable entity on invalid category",
			setAuth: true,
			userID:  1,
			body:    map[string]any{"title": "Coffee", "amount": "3.50", "date": "2026-01-01", "category": "InvalidCat"},
			repo:    &mockExpenseRepo{},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name:    "internal server error on repo failure",
			setAuth: true,
			userID:  1,
			body:    map[string]any{"title": "Coffee", "amount": "3.50", "date": "2026-01-01"},
			repo: &mockExpenseRepo{
				createExpenseFunc: func(ctx context.Context, userID int64, title string, amount decimal.Decimal, category models.Category, date time.Time, description string) (*models.Expense, error) {
					return nil, errors.New("db error")
				},
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:    "success creates expense with explicit category",
			setAuth: true,
			userID:  1,
			body:    map[string]any{"title": "Groceries run", "amount": "42.00", "date": "2026-01-15", "category": "Groceries", "description": "Weekly shop"},
			repo: &mockExpenseRepo{
				createExpenseFunc: func(ctx context.Context, userID int64, title string, amount decimal.Decimal, category models.Category, date time.Time, description string) (*models.Expense, error) {
					if title != "Groceries run" {
						return nil, errors.New("unexpected title")
					}
					if category != models.CategoryGroceries {
						return nil, errors.New("unexpected category")
					}
					return &models.Expense{
						ID:          10,
						UserID:      userID,
						Title:       title,
						Amount:      amount,
						Category:    category,
						Date:        date,
						Description: description,
						CreatedAt:   time.Now(),
					}, nil
				},
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:    "success defaults category to Others when omitted",
			setAuth: true,
			userID:  1,
			body:    map[string]any{"title": "Mystery item", "amount": "1.99", "date": "2026-02-01"},
			repo: &mockExpenseRepo{
				createExpenseFunc: func(ctx context.Context, userID int64, title string, amount decimal.Decimal, category models.Category, date time.Time, description string) (*models.Expense, error) {
					if category != models.CategoryOthers {
						return nil, errors.New("expected default category Others")
					}
					return &models.Expense{ID: 11, UserID: userID, Title: title, Amount: amount, Category: category, Date: date}, nil
				},
			},
			expectedStatus: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewWithRepos(nil, tt.repo, newMockRefreshStore(), secret)

			var req *http.Request
			if tt.rawBody != "" {
				req = httptest.NewRequest(http.MethodPost, "/expenses", bytes.NewBufferString(tt.rawBody))
			} else {
				b, _ := json.Marshal(tt.body)
				req = httptest.NewRequest(http.MethodPost, "/expenses", bytes.NewReader(b))
			}

			if tt.setAuth {
				req = auth.SetUserID(req, tt.userID)
			}

			rec := httptest.NewRecorder()
			h.CreateExpense(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Fatalf("expected status %d, got %d. Body: %s", tt.expectedStatus, rec.Code, rec.Body.String())
			}
		})
	}
}

// ---- GetExpense tests ----

func TestGetExpense(t *testing.T) {
	secret := []byte("test-secret")

	tests := []struct {
		name           string
		pathID         string
		setAuth        bool
		userID         int64
		repo           *mockExpenseRepo
		expectedStatus int
	}{
		{
			name:           "unauthorized when no user id",
			pathID:         "1",
			setAuth:        false,
			repo:           &mockExpenseRepo{},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "bad request on invalid path id",
			pathID:         "invalid",
			setAuth:        true,
			userID:         1,
			repo:           &mockExpenseRepo{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:    "not found when expense does not exist",
			pathID:  "99",
			setAuth: true,
			userID:  1,
			repo: &mockExpenseRepo{
				getExpenseFunc: func(ctx context.Context, id, userID int64) (*models.Expense, error) {
					return nil, repository.ErrNotFound
				},
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:    "not found when expense belongs to another user",
			pathID:  "5",
			setAuth: true,
			userID:  1,
			repo: &mockExpenseRepo{
				getExpenseFunc: func(ctx context.Context, id, userID int64) (*models.Expense, error) {
					// repo enforces ownership — returns ErrNotFound for wrong user
					return nil, repository.ErrNotFound
				},
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:    "internal server error on repo failure",
			pathID:  "1",
			setAuth: true,
			userID:  1,
			repo: &mockExpenseRepo{
				getExpenseFunc: func(ctx context.Context, id, userID int64) (*models.Expense, error) {
					return nil, errors.New("db error")
				},
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:    "success returns expense",
			pathID:  "7",
			setAuth: true,
			userID:  1,
			repo: &mockExpenseRepo{
				getExpenseFunc: func(ctx context.Context, id, userID int64) (*models.Expense, error) {
					return &models.Expense{
						ID:       id,
						UserID:   userID,
						Title:    "Electricity",
						Amount:   decimal.NewFromFloat(55.00),
						Category: models.CategoryUtilities,
						Date:     time.Now(),
					}, nil
				},
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewWithRepos(nil, tt.repo, newMockRefreshStore(), secret)

			req := httptest.NewRequest(http.MethodGet, "/expenses/"+tt.pathID, nil)
			req.SetPathValue("id", tt.pathID)

			if tt.setAuth {
				req = auth.SetUserID(req, tt.userID)
			}

			rec := httptest.NewRecorder()
			h.GetExpense(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Fatalf("expected status %d, got %d. Body: %s", tt.expectedStatus, rec.Code, rec.Body.String())
			}
		})
	}
}

// ---- ListExpenses tests ----

func TestListExpenses(t *testing.T) {
	secret := []byte("test-secret")

	tests := []struct {
		name           string
		query          string
		setAuth        bool
		userID         int64
		repo           *mockExpenseRepo
		expectedStatus int
		verifyBody     func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name:           "unauthorized when no user id",
			query:          "",
			setAuth:        false,
			repo:           &mockExpenseRepo{},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:    "bad request on custom filter without start_date",
			query:   "?filter=custom&end_date=2026-01-31",
			setAuth: true,
			userID:  1,
			repo:    &mockExpenseRepo{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:    "bad request on custom filter without end_date",
			query:   "?filter=custom&start_date=2026-01-01",
			setAuth: true,
			userID:  1,
			repo:    &mockExpenseRepo{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:    "bad request on invalid start_date format",
			query:   "?filter=custom&start_date=01-01-2026&end_date=2026-01-31",
			setAuth: true,
			userID:  1,
			repo:    &mockExpenseRepo{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:    "bad request on invalid end_date format",
			query:   "?filter=custom&start_date=2026-01-01&end_date=31-01-2026",
			setAuth: true,
			userID:  1,
			repo:    &mockExpenseRepo{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:    "internal server error on repo failure",
			query:   "",
			setAuth: true,
			userID:  1,
			repo: &mockExpenseRepo{
				listExpensesFunc: func(ctx context.Context, userID int64, filter repository.ExpenseFilter) ([]*models.Expense, error) {
					return nil, errors.New("db error")
				},
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:    "success returns empty list when no expenses",
			query:   "",
			setAuth: true,
			userID:  1,
			repo: &mockExpenseRepo{
				listExpensesFunc: func(ctx context.Context, userID int64, filter repository.ExpenseFilter) ([]*models.Expense, error) {
					return nil, nil
				},
			},
			expectedStatus: http.StatusOK,
			verifyBody: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var res []any
				if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
					t.Fatalf("failed to decode json: %v", err)
				}
				if len(res) != 0 {
					t.Errorf("expected empty array, got %d items", len(res))
				}
			},
		},
		{
			name:    "success passes week filter with start date",
			query:   "?filter=week",
			setAuth: true,
			userID:  1,
			repo: &mockExpenseRepo{
				listExpensesFunc: func(ctx context.Context, userID int64, filter repository.ExpenseFilter) ([]*models.Expense, error) {
					if filter.StartDate == nil {
						return nil, errors.New("expected StartDate to be set for week filter")
					}
					return []*models.Expense{}, nil
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:    "success passes month filter with start date",
			query:   "?filter=month",
			setAuth: true,
			userID:  1,
			repo: &mockExpenseRepo{
				listExpensesFunc: func(ctx context.Context, userID int64, filter repository.ExpenseFilter) ([]*models.Expense, error) {
					if filter.StartDate == nil {
						return nil, errors.New("expected StartDate to be set for month filter")
					}
					return []*models.Expense{}, nil
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:    "success passes 3months filter with start date",
			query:   "?filter=3months",
			setAuth: true,
			userID:  1,
			repo: &mockExpenseRepo{
				listExpensesFunc: func(ctx context.Context, userID int64, filter repository.ExpenseFilter) ([]*models.Expense, error) {
					if filter.StartDate == nil {
						return nil, errors.New("expected StartDate to be set for 3months filter")
					}
					return []*models.Expense{}, nil
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:    "success passes custom filter with both dates",
			query:   "?filter=custom&start_date=2026-01-01&end_date=2026-01-31",
			setAuth: true,
			userID:  1,
			repo: &mockExpenseRepo{
				listExpensesFunc: func(ctx context.Context, userID int64, filter repository.ExpenseFilter) ([]*models.Expense, error) {
					if filter.StartDate == nil || filter.EndDate == nil {
						return nil, errors.New("expected both dates for custom filter")
					}
					return []*models.Expense{
						{ID: 1, Title: "Test", Amount: decimal.NewFromFloat(5.0), Category: models.CategoryOthers},
					}, nil
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:    "success no filter returns all expenses",
			query:   "",
			setAuth: true,
			userID:  1,
			repo: &mockExpenseRepo{
				listExpensesFunc: func(ctx context.Context, userID int64, filter repository.ExpenseFilter) ([]*models.Expense, error) {
					if filter.StartDate != nil || filter.EndDate != nil {
						return nil, errors.New("expected no date filters")
					}
					return []*models.Expense{
						{ID: 1, Title: "A", Amount: decimal.NewFromFloat(1.0), Category: models.CategoryOthers},
						{ID: 2, Title: "B", Amount: decimal.NewFromFloat(2.0), Category: models.CategoryGroceries},
					}, nil
				},
			},
			expectedStatus: http.StatusOK,
			verifyBody: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var res []any
				if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
					t.Fatalf("failed to decode json: %v", err)
				}
				if len(res) != 2 {
					t.Errorf("expected 2 items, got %d", len(res))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewWithRepos(nil, tt.repo, newMockRefreshStore(), secret)

			req := httptest.NewRequest(http.MethodGet, "/expenses"+tt.query, nil)
			if tt.setAuth {
				req = auth.SetUserID(req, tt.userID)
			}

			rec := httptest.NewRecorder()
			h.ListExpenses(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Fatalf("expected status %d, got %d. Body: %s", tt.expectedStatus, rec.Code, rec.Body.String())
			}
			if tt.verifyBody != nil {
				tt.verifyBody(t, rec)
			}
		})
	}
}

// ---- UpdateExpense tests ----

func TestUpdateExpense(t *testing.T) {
	secret := []byte("test-secret")

	strPtr := func(s string) *string { return &s }

	tests := []struct {
		name           string
		pathID         string
		setAuth        bool
		userID         int64
		body           any
		rawBody        string
		repo           *mockExpenseRepo
		expectedStatus int
	}{
		{
			name:           "unauthorized when no user id",
			pathID:         "1",
			setAuth:        false,
			body:           map[string]any{"title": "Updated"},
			repo:           &mockExpenseRepo{},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "bad request on invalid path id",
			pathID:         "abc",
			setAuth:        true,
			userID:         1,
			body:           map[string]any{"title": "Updated"},
			repo:           &mockExpenseRepo{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:    "bad request on invalid json body",
			pathID:  "1",
			setAuth: true,
			userID:  1,
			rawBody: "{invalid-json",
			repo:    &mockExpenseRepo{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:    "unprocessable entity on empty title",
			pathID:  "1",
			setAuth: true,
			userID:  1,
			body:    map[string]any{"title": "   "},
			repo:    &mockExpenseRepo{},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name:    "unprocessable entity on negative amount",
			pathID:  "1",
			setAuth: true,
			userID:  1,
			body:    map[string]any{"amount": "-10.00"},
			repo:    &mockExpenseRepo{},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name:    "unprocessable entity on invalid date",
			pathID:  "1",
			setAuth: true,
			userID:  1,
			body:    map[string]any{"date": "not-a-date"},
			repo:    &mockExpenseRepo{},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name:    "not found when expense does not exist",
			pathID:  "99",
			setAuth: true,
			userID:  1,
			body:    map[string]any{"title": "New Title"},
			repo: &mockExpenseRepo{
				updateExpenseFunc: func(ctx context.Context, id, userID int64, title *string, amount *decimal.Decimal, category *models.Category, date *time.Time, description *string) (*models.Expense, error) {
					return nil, repository.ErrNotFound
				},
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:    "internal server error on repo failure",
			pathID:  "1",
			setAuth: true,
			userID:  1,
			body:    map[string]any{"title": "New Title"},
			repo: &mockExpenseRepo{
				updateExpenseFunc: func(ctx context.Context, id, userID int64, title *string, amount *decimal.Decimal, category *models.Category, date *time.Time, description *string) (*models.Expense, error) {
					return nil, errors.New("db error")
				},
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:    "success updates title only",
			pathID:  "3",
			setAuth: true,
			userID:  1,
			body:    map[string]any{"title": "Renamed"},
			repo: &mockExpenseRepo{
				updateExpenseFunc: func(ctx context.Context, id, userID int64, title *string, amount *decimal.Decimal, category *models.Category, date *time.Time, description *string) (*models.Expense, error) {
					if id != 3 || title == nil || *title != "Renamed" {
						return nil, errors.New("unexpected params")
					}
					return &models.Expense{ID: id, UserID: userID, Title: *title, Amount: decimal.NewFromFloat(10), Category: models.CategoryOthers}, nil
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:    "success updates all fields",
			pathID:  "4",
			setAuth: true,
			userID:  1,
			body: map[string]any{
				"title":       "Full Update",
				"amount":      "99.99",
				"category":    "Health",
				"date":        "2026-03-01",
				"description": "Updated desc",
			},
			repo: &mockExpenseRepo{
				updateExpenseFunc: func(ctx context.Context, id, userID int64, title *string, amount *decimal.Decimal, category *models.Category, date *time.Time, description *string) (*models.Expense, error) {
					_ = strPtr("unused") // ensure strPtr is used
					return &models.Expense{
						ID:       id,
						UserID:   userID,
						Title:    *title,
						Amount:   *amount,
						Category: *category,
						Date:     *date,
					}, nil
				},
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewWithRepos(nil, tt.repo, newMockRefreshStore(), secret)

			var req *http.Request
			if tt.rawBody != "" {
				req = httptest.NewRequest(http.MethodPut, "/expenses/"+tt.pathID, bytes.NewBufferString(tt.rawBody))
			} else {
				b, _ := json.Marshal(tt.body)
				req = httptest.NewRequest(http.MethodPut, "/expenses/"+tt.pathID, bytes.NewReader(b))
			}
			req.SetPathValue("id", tt.pathID)

			if tt.setAuth {
				req = auth.SetUserID(req, tt.userID)
			}

			rec := httptest.NewRecorder()
			h.UpdateExpense(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Fatalf("expected status %d, got %d. Body: %s", tt.expectedStatus, rec.Code, rec.Body.String())
			}
		})
	}
}

// ---- DeleteExpense tests ----

func TestDeleteExpense(t *testing.T) {
	secret := []byte("test-secret")

	tests := []struct {
		name           string
		pathID         string
		setAuth        bool
		userID         int64
		repo           *mockExpenseRepo
		expectedStatus int
	}{
		{
			name:           "unauthorized when no user id",
			pathID:         "1",
			setAuth:        false,
			repo:           &mockExpenseRepo{},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "bad request on invalid path id",
			pathID:         "abc",
			setAuth:        true,
			userID:         1,
			repo:           &mockExpenseRepo{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:    "not found when expense does not exist",
			pathID:  "99",
			setAuth: true,
			userID:  1,
			repo: &mockExpenseRepo{
				deleteExpenseFunc: func(ctx context.Context, id, userID int64) error {
					return repository.ErrNotFound
				},
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:    "not found when expense belongs to another user",
			pathID:  "5",
			setAuth: true,
			userID:  1,
			repo: &mockExpenseRepo{
				deleteExpenseFunc: func(ctx context.Context, id, userID int64) error {
					return repository.ErrNotFound
				},
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:    "internal server error on repo failure",
			pathID:  "1",
			setAuth: true,
			userID:  1,
			repo: &mockExpenseRepo{
				deleteExpenseFunc: func(ctx context.Context, id, userID int64) error {
					return errors.New("db error")
				},
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:    "success deletes expense",
			pathID:  "10",
			setAuth: true,
			userID:  1,
			repo: &mockExpenseRepo{
				deleteExpenseFunc: func(ctx context.Context, id, userID int64) error {
					if id != 10 || userID != 1 {
						return errors.New("unexpected params")
					}
					return nil
				},
			},
			expectedStatus: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewWithRepos(nil, tt.repo, newMockRefreshStore(), secret)

			req := httptest.NewRequest(http.MethodDelete, "/expenses/"+tt.pathID, nil)
			req.SetPathValue("id", tt.pathID)

			if tt.setAuth {
				req = auth.SetUserID(req, tt.userID)
			}

			rec := httptest.NewRecorder()
			h.DeleteExpense(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Fatalf("expected status %d, got %d. Body: %s", tt.expectedStatus, rec.Code, rec.Body.String())
			}
		})
	}
}
