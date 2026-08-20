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

	"github.com/xenptr/go-projects/todo-list-api/internal/auth"
	"github.com/xenptr/go-projects/todo-list-api/internal/models"
)

type mockTodoRepo struct {
	createTodoFunc      func(ctx context.Context, t models.Todo) (int64, error)
	getTodoByIDFunc     func(ctx context.Context, id int64) (models.Todo, error)
	listTodosByUserFunc func(ctx context.Context, userID int64, page *int64, limit *int64, search *string) ([]models.Todo, int64, error)
	getTodoOwnerFunc    func(ctx context.Context, id int64) (int64, error)
	updateTodoFunc      func(ctx context.Context, id int64, t models.Todo) error
	deleteTodoFunc      func(ctx context.Context, id int64) error
}

func (m *mockTodoRepo) CreateTodo(ctx context.Context, t models.Todo) (int64, error) {
	if m.createTodoFunc != nil {
		return m.createTodoFunc(ctx, t)
	}
	return 1, nil
}

func (m *mockTodoRepo) GetTodoByID(ctx context.Context, id int64) (models.Todo, error) {
	if m.getTodoByIDFunc != nil {
		return m.getTodoByIDFunc(ctx, id)
	}
	return models.Todo{ID: id, UserID: 1, Title: "Sample Todo", Completed: false}, nil
}

func (m *mockTodoRepo) ListTodosByUser(ctx context.Context, userID int64, page *int64, limit *int64, search *string) ([]models.Todo, int64, error) {
	if m.listTodosByUserFunc != nil {
		return m.listTodosByUserFunc(ctx, userID, page, limit, search)
	}
	return []models.Todo{}, 0, nil
}

func (m *mockTodoRepo) GetTodoOwner(ctx context.Context, id int64) (int64, error) {
	if m.getTodoOwnerFunc != nil {
		return m.getTodoOwnerFunc(ctx, id)
	}
	return 1, nil
}

func (m *mockTodoRepo) UpdateTodo(ctx context.Context, id int64, t models.Todo) error {
	if m.updateTodoFunc != nil {
		return m.updateTodoFunc(ctx, id, t)
	}
	return nil
}

func (m *mockTodoRepo) DeleteTodo(ctx context.Context, id int64) error {
	if m.deleteTodoFunc != nil {
		return m.deleteTodoFunc(ctx, id)
	}
	return nil
}

func TestCreateTodo(t *testing.T) {
	secret := []byte("test-secret")

	tests := []struct {
		name           string
		setAuth        bool
		userID         int64
		body           any
		rawBody        string
		repo           *mockTodoRepo
		expectedStatus int
	}{
		{
			name:           "unauthorized when no user id",
			setAuth:        false,
			body:           map[string]string{"title": "Test"},
			repo:           &mockTodoRepo{},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "bad request when invalid json",
			setAuth:        true,
			userID:         1,
			rawBody:        "{invalid-json",
			repo:           &mockTodoRepo{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "bad request when validation fails",
			setAuth:        true,
			userID:         1,
			body:           map[string]string{"title": "   "},
			repo:           &mockTodoRepo{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:    "internal server error on repo CreateTodo error",
			setAuth: true,
			userID:  1,
			body:    map[string]string{"title": "Test"},
			repo: &mockTodoRepo{
				createTodoFunc: func(ctx context.Context, t models.Todo) (int64, error) {
					return 0, errors.New("db error")
				},
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:    "internal server error on repo GetTodoByID error",
			setAuth: true,
			userID:  1,
			body:    map[string]string{"title": "Test"},
			repo: &mockTodoRepo{
				createTodoFunc: func(ctx context.Context, t models.Todo) (int64, error) {
					return 10, nil
				},
				getTodoByIDFunc: func(ctx context.Context, id int64) (models.Todo, error) {
					return models.Todo{}, errors.New("db error")
				},
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:    "success creates todo",
			setAuth: true,
			userID:  1,
			body:    map[string]string{"title": "Test Title", "description": "Test Desc"},
			repo: &mockTodoRepo{
				createTodoFunc: func(ctx context.Context, t models.Todo) (int64, error) {
					if t.Title != "Test Title" || t.Description != "Test Desc" || t.UserID != 1 {
						return 0, errors.New("unexpected params")
					}
					return 42, nil
				},
				getTodoByIDFunc: func(ctx context.Context, id int64) (models.Todo, error) {
					return models.Todo{
						ID:          id,
						UserID:      1,
						Title:       "Test Title",
						Description: "Test Desc",
						Completed:   false,
						CreatedAt:   time.Now(),
					}, nil
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
				req = httptest.NewRequest(http.MethodPost, "/todos", bytes.NewBufferString(tt.rawBody))
			} else {
				b, _ := json.Marshal(tt.body)
				req = httptest.NewRequest(http.MethodPost, "/todos", bytes.NewReader(b))
			}

			if tt.setAuth {
				req = auth.SetUserID(req, tt.userID)
			}

			rec := httptest.NewRecorder()
			h.CreateTodo(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Fatalf("expected status %d, got %d. Body: %s", tt.expectedStatus, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestUpdateTodo(t *testing.T) {
	secret := []byte("test-secret")

	tests := []struct {
		name           string
		pathID         string
		setAuth        bool
		userID         int64
		body           any
		rawBody        string
		repo           *mockTodoRepo
		expectedStatus int
	}{
		{
			name:           "bad request on invalid path id",
			pathID:         "invalid",
			setAuth:        true,
			userID:         1,
			body:           map[string]any{"title": "Valid"},
			repo:           &mockTodoRepo{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "unauthorized when no user id",
			pathID:         "10",
			setAuth:        false,
			body:           map[string]any{"title": "Valid"},
			repo:           &mockTodoRepo{},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:    "not found when todo does not exist",
			pathID:  "10",
			setAuth: true,
			userID:  1,
			body:    map[string]any{"title": "Valid"},
			repo: &mockTodoRepo{
				getTodoOwnerFunc: func(ctx context.Context, id int64) (int64, error) {
					return 0, errors.New("not found")
				},
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:    "forbidden when user is not owner",
			pathID:  "10",
			setAuth: true,
			userID:  1,
			body:    map[string]any{"title": "Valid"},
			repo: &mockTodoRepo{
				getTodoOwnerFunc: func(ctx context.Context, id int64) (int64, error) {
					return 2, nil
				},
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:    "bad request on invalid json body",
			pathID:  "10",
			setAuth: true,
			userID:  1,
			rawBody: "{invalid-json",
			repo: &mockTodoRepo{
				getTodoOwnerFunc: func(ctx context.Context, id int64) (int64, error) {
					return 1, nil
				},
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:    "bad request on validation error",
			pathID:  "10",
			setAuth: true,
			userID:  1,
			body:    map[string]any{"title": "   "},
			repo: &mockTodoRepo{
				getTodoOwnerFunc: func(ctx context.Context, id int64) (int64, error) {
					return 1, nil
				},
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:    "internal server error on repo UpdateTodo error",
			pathID:  "10",
			setAuth: true,
			userID:  1,
			body:    map[string]any{"title": "Updated"},
			repo: &mockTodoRepo{
				getTodoOwnerFunc: func(ctx context.Context, id int64) (int64, error) {
					return 1, nil
				},
				updateTodoFunc: func(ctx context.Context, id int64, t models.Todo) error {
					return errors.New("db error")
				},
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:    "internal server error on repo GetTodoByID error",
			pathID:  "10",
			setAuth: true,
			userID:  1,
			body:    map[string]any{"title": "Updated"},
			repo: &mockTodoRepo{
				getTodoOwnerFunc: func(ctx context.Context, id int64) (int64, error) {
					return 1, nil
				},
				updateTodoFunc: func(ctx context.Context, id int64, t models.Todo) error {
					return nil
				},
				getTodoByIDFunc: func(ctx context.Context, id int64) (models.Todo, error) {
					return models.Todo{}, errors.New("db error")
				},
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:    "success updates todo",
			pathID:  "10",
			setAuth: true,
			userID:  1,
			body:    map[string]any{"title": "Updated Title", "description": "Updated Desc", "completed": true},
			repo: &mockTodoRepo{
				getTodoOwnerFunc: func(ctx context.Context, id int64) (int64, error) {
					return 1, nil
				},
				updateTodoFunc: func(ctx context.Context, id int64, t models.Todo) error {
					if id != 10 || t.Title != "Updated Title" || !t.Completed {
						return errors.New("unexpected params")
					}
					return nil
				},
				getTodoByIDFunc: func(ctx context.Context, id int64) (models.Todo, error) {
					return models.Todo{
						ID:          id,
						UserID:      1,
						Title:       "Updated Title",
						Description: "Updated Desc",
						Completed:   true,
						CreatedAt:   time.Now(),
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
				req = httptest.NewRequest(http.MethodPut, "/todos/"+tt.pathID, bytes.NewBufferString(tt.rawBody))
			} else {
				b, _ := json.Marshal(tt.body)
				req = httptest.NewRequest(http.MethodPut, "/todos/"+tt.pathID, bytes.NewReader(b))
			}
			req.SetPathValue("id", tt.pathID)

			if tt.setAuth {
				req = auth.SetUserID(req, tt.userID)
			}

			rec := httptest.NewRecorder()
			h.UpdateTodo(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Fatalf("expected status %d, got %d. Body: %s", tt.expectedStatus, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestDeleteTodo(t *testing.T) {
	secret := []byte("test-secret")

	tests := []struct {
		name           string
		pathID         string
		setAuth        bool
		userID         int64
		repo           *mockTodoRepo
		expectedStatus int
	}{
		{
			name:           "bad request on invalid path id",
			pathID:         "invalid",
			setAuth:        true,
			userID:         1,
			repo:           &mockTodoRepo{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "unauthorized when no user id",
			pathID:         "10",
			setAuth:        false,
			repo:           &mockTodoRepo{},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:    "not found when todo does not exist",
			pathID:  "10",
			setAuth: true,
			userID:  1,
			repo: &mockTodoRepo{
				getTodoOwnerFunc: func(ctx context.Context, id int64) (int64, error) {
					return 0, errors.New("not found")
				},
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:    "forbidden when user is not owner",
			pathID:  "10",
			setAuth: true,
			userID:  1,
			repo: &mockTodoRepo{
				getTodoOwnerFunc: func(ctx context.Context, id int64) (int64, error) {
					return 2, nil
				},
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:    "internal server error on repo DeleteTodo error",
			pathID:  "10",
			setAuth: true,
			userID:  1,
			repo: &mockTodoRepo{
				getTodoOwnerFunc: func(ctx context.Context, id int64) (int64, error) {
					return 1, nil
				},
				deleteTodoFunc: func(ctx context.Context, id int64) error {
					return errors.New("db error")
				},
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:    "success deletes todo",
			pathID:  "10",
			setAuth: true,
			userID:  1,
			repo: &mockTodoRepo{
				getTodoOwnerFunc: func(ctx context.Context, id int64) (int64, error) {
					return 1, nil
				},
				deleteTodoFunc: func(ctx context.Context, id int64) error {
					return nil
				},
			},
			expectedStatus: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewWithRepos(nil, tt.repo, newMockRefreshStore(), secret)

			req := httptest.NewRequest(http.MethodDelete, "/todos/"+tt.pathID, nil)
			req.SetPathValue("id", tt.pathID)

			if tt.setAuth {
				req = auth.SetUserID(req, tt.userID)
			}

			rec := httptest.NewRecorder()
			h.DeleteTodo(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Fatalf("expected status %d, got %d. Body: %s", tt.expectedStatus, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestListTodos(t *testing.T) {
	secret := []byte("test-secret")

	tests := []struct {
		name           string
		query          string
		setAuth        bool
		userID         int64
		repo           *mockTodoRepo
		expectedStatus int
		verifyBody     func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name:           "unauthorized when no user id",
			query:          "",
			setAuth:        false,
			repo:           &mockTodoRepo{},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "bad request on invalid page",
			query:          "?page=abc",
			setAuth:        true,
			userID:         1,
			repo:           &mockTodoRepo{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "bad request on zero page",
			query:          "?page=0",
			setAuth:        true,
			userID:         1,
			repo:           &mockTodoRepo{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "bad request on invalid limit",
			query:          "?limit=abc",
			setAuth:        true,
			userID:         1,
			repo:           &mockTodoRepo{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "bad request on zero limit",
			query:          "?limit=0",
			setAuth:        true,
			userID:         1,
			repo:           &mockTodoRepo{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "bad request when page provided without limit",
			query:          "?page=1",
			setAuth:        true,
			userID:         1,
			repo:           &mockTodoRepo{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:    "internal server error on repo error",
			query:   "",
			setAuth: true,
			userID:  1,
			repo: &mockTodoRepo{
				listTodosByUserFunc: func(ctx context.Context, userID int64, page *int64, limit *int64, search *string) ([]models.Todo, int64, error) {
					return nil, 0, errors.New("db error")
				},
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:    "success with query search parameter",
			query:   "?search=learn&page=1&limit=5",
			setAuth: true,
			userID:  1,
			repo: &mockTodoRepo{
				listTodosByUserFunc: func(ctx context.Context, userID int64, page *int64, limit *int64, search *string) ([]models.Todo, int64, error) {
					if search == nil || *search != "learn" {
						return nil, 0, errors.New("expected search param")
					}
					if page == nil || *page != 1 || limit == nil || *limit != 5 {
						return nil, 0, errors.New("expected pagination")
					}
					return []models.Todo{{ID: 1, Title: "learn go"}}, 1, nil
				},
			},
			expectedStatus: http.StatusOK,
			verifyBody: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var res map[string]any
				if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
					t.Fatalf("failed to decode json: %v", err)
				}
				if res["page"].(float64) != 1 || res["limit"].(float64) != 5 || res["total"].(float64) != 1 {
					t.Errorf("unexpected response: %+v", res)
				}
			},
		},
		{
			name:    "success with s query parameter",
			query:   "?s=architect",
			setAuth: true,
			userID:  1,
			repo: &mockTodoRepo{
				listTodosByUserFunc: func(ctx context.Context, userID int64, page *int64, limit *int64, search *string) ([]models.Todo, int64, error) {
					if search == nil || *search != "architect" {
						return nil, 0, errors.New("expected s param")
					}
					return []models.Todo{{ID: 2, Title: "architect"}}, 1, nil
				},
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewWithRepos(nil, tt.repo, newMockRefreshStore(), secret)

			req := httptest.NewRequest(http.MethodGet, "/todos"+tt.query, nil)
			if tt.setAuth {
				req = auth.SetUserID(req, tt.userID)
			}

			rec := httptest.NewRecorder()
			h.ListTodos(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Fatalf("expected status %d, got %d. Body: %s", tt.expectedStatus, rec.Code, rec.Body.String())
			}
			if tt.verifyBody != nil {
				tt.verifyBody(t, rec)
			}
		})
	}
}
