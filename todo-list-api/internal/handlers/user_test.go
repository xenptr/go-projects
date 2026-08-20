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

type mockUserRepo struct {
	createUserFunc     func(ctx context.Context, u models.User) (int64, error)
	getUserByIDFunc    func(ctx context.Context, id int64) (models.User, error)
	getUserByEmailFunc func(ctx context.Context, email string) (models.User, error)
}

func (m *mockUserRepo) CreateUser(ctx context.Context, u models.User) (int64, error) {
	if m.createUserFunc != nil {
		return m.createUserFunc(ctx, u)
	}
	return 1, nil
}

func (m *mockUserRepo) GetUserByID(ctx context.Context, id int64) (models.User, error) {
	if m.getUserByIDFunc != nil {
		return m.getUserByIDFunc(ctx, id)
	}
	return models.User{ID: id, Name: "Test User", Email: "test@example.com"}, nil
}

func (m *mockUserRepo) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	if m.getUserByEmailFunc != nil {
		return m.getUserByEmailFunc(ctx, email)
	}
	return models.User{}, errors.New("user not found")
}

func TestCreateUser(t *testing.T) {
	secret := []byte("test-jwt-secret-key-32-bytes-long!")

	tests := []struct {
		name           string
		body           any
		rawBody        string
		repo           *mockUserRepo
		expectedStatus int
	}{
		{
			name:           "bad request on invalid json body",
			rawBody:        "{invalid-json",
			repo:           &mockUserRepo{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "bad request on validation error",
			body: map[string]string{
				"name":     "",
				"email":    "invalid-email",
				"password": "short",
			},
			repo:           &mockUserRepo{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "conflict when email already in use duplicate",
			body: map[string]string{
				"name":     "Jane Doe",
				"email":    "existing@example.com",
				"password": "strongPassword123!",
			},
			repo: &mockUserRepo{
				createUserFunc: func(ctx context.Context, u models.User) (int64, error) {
					return 0, errors.New("duplicate key error")
				},
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name: "conflict when email already in use unique",
			body: map[string]string{
				"name":     "Jane Doe",
				"email":    "existing@example.com",
				"password": "strongPassword123!",
			},
			repo: &mockUserRepo{
				createUserFunc: func(ctx context.Context, u models.User) (int64, error) {
					return 0, errors.New("violates unique constraint")
				},
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name: "internal server error on unexpected repo failure",
			body: map[string]string{
				"name":     "Jane Doe",
				"email":    "jane@example.com",
				"password": "strongPassword123!",
			},
			repo: &mockUserRepo{
				createUserFunc: func(ctx context.Context, u models.User) (int64, error) {
					return 0, errors.New("database connection down")
				},
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "success creates user and returns token",
			body: map[string]string{
				"name":     "Jane Doe",
				"email":    "jane@example.com",
				"password": "strongPassword123!",
			},
			repo: &mockUserRepo{
				createUserFunc: func(ctx context.Context, u models.User) (int64, error) {
					if u.Email != "jane@example.com" || u.Name != "Jane Doe" {
						return 0, errors.New("unexpected params")
					}
					return 42, nil
				},
			},
			expectedStatus: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewWithRepos(tt.repo, nil, secret)

			var req *http.Request
			if tt.rawBody != "" {
				req = httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(tt.rawBody))
			} else {
				b, _ := json.Marshal(tt.body)
				req = httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(b))
			}

			rec := httptest.NewRecorder()
			h.CreateUser(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Fatalf("expected status %d, got %d. Body: %s", tt.expectedStatus, rec.Code, rec.Body.String())
			}

			if tt.expectedStatus == http.StatusCreated {
				var res map[string]string
				if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if res["token"] == "" {
					t.Error("expected non-empty token")
				}
			}
		})
	}
}

func TestLogin(t *testing.T) {
	secret := []byte("test-jwt-secret-key-32-bytes-long!")
	password := "correctPassword123"
	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	tests := []struct {
		name           string
		body           any
		rawBody        string
		repo           *mockUserRepo
		expectedStatus int
	}{
		{
			name:           "bad request on invalid json body",
			rawBody:        "{invalid-json",
			repo:           &mockUserRepo{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "bad request on validation error",
			body: map[string]string{
				"email":    "not-an-email",
				"password": "",
			},
			repo:           &mockUserRepo{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "unauthorized when user not found",
			body: map[string]string{
				"email":    "unknown@example.com",
				"password": password,
			},
			repo: &mockUserRepo{
				getUserByEmailFunc: func(ctx context.Context, email string) (models.User, error) {
					return models.User{}, errors.New("user not found")
				},
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "unauthorized on incorrect password",
			body: map[string]string{
				"email":    "john@example.com",
				"password": "wrongPassword123",
			},
			repo: &mockUserRepo{
				getUserByEmailFunc: func(ctx context.Context, email string) (models.User, error) {
					return models.User{
						ID:           10,
						Email:        email,
						PasswordHash: hash,
					}, nil
				},
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "success returns token on valid credentials",
			body: map[string]string{
				"email":    "john@example.com",
				"password": password,
			},
			repo: &mockUserRepo{
				getUserByEmailFunc: func(ctx context.Context, email string) (models.User, error) {
					return models.User{
						ID:           10,
						Name:         "John",
						Email:        email,
						PasswordHash: hash,
						CreatedAt:    time.Now(),
					}, nil
				},
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewWithRepos(tt.repo, nil, secret)

			var req *http.Request
			if tt.rawBody != "" {
				req = httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(tt.rawBody))
			} else {
				b, _ := json.Marshal(tt.body)
				req = httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(b))
			}

			rec := httptest.NewRecorder()
			h.Login(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Fatalf("expected status %d, got %d. Body: %s", tt.expectedStatus, rec.Code, rec.Body.String())
			}

			if tt.expectedStatus == http.StatusOK {
				var res map[string]string
				if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if res["token"] == "" {
					t.Error("expected non-empty token")
				}
			}
		})
	}
}
