package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/xenptr/go-projects/expense-tracker-api/internal/auth"
	"github.com/xenptr/go-projects/expense-tracker-api/internal/dto"
	"github.com/xenptr/go-projects/expense-tracker-api/internal/models"
)

// ---- mock user repository ----

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

// ---- in-memory refresh token store ----

type mockRefreshStore struct {
	mu     sync.Mutex
	tokens map[string]struct{}

	saveErr   error
	existsErr error
	revokeErr error
}

func newMockRefreshStore() *mockRefreshStore {
	return &mockRefreshStore{tokens: make(map[string]struct{})}
}

func (s *mockRefreshStore) key(userID int64, token string) string {
	return fmt.Sprintf("%d:%s", userID, token)
}

func (s *mockRefreshStore) Save(_ context.Context, userID int64, token string, _ time.Duration) error {
	if s.saveErr != nil {
		return s.saveErr
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tokens[s.key(userID, token)] = struct{}{}
	return nil
}

func (s *mockRefreshStore) Exists(_ context.Context, userID int64, token string) (bool, error) {
	if s.existsErr != nil {
		return false, s.existsErr
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.tokens[s.key(userID, token)]
	return ok, nil
}

func (s *mockRefreshStore) Revoke(_ context.Context, userID int64, token string) error {
	if s.revokeErr != nil {
		return s.revokeErr
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.tokens, s.key(userID, token))
	return nil
}

// ---- helper ----

func decodeAuthResponse(t *testing.T, body *bytes.Buffer) dto.AuthResponse {
	t.Helper()
	var res dto.AuthResponse
	if err := json.NewDecoder(body).Decode(&res); err != nil {
		t.Fatalf("failed to decode auth response: %v", err)
	}
	return res
}

// ---- CreateUser tests ----

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
			name: "success creates user and returns token pair",
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
			h := NewWithRepos(tt.repo, nil, newMockRefreshStore(), secret)

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
				res := decodeAuthResponse(t, rec.Body)
				if res.Token == "" {
					t.Error("expected non-empty token")
				}
				if res.RefreshToken == "" {
					t.Error("expected non-empty refresh_token")
				}
			}
		})
	}
}

// ---- Login tests ----

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
			name: "success returns token pair on valid credentials",
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
			h := NewWithRepos(tt.repo, nil, newMockRefreshStore(), secret)

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
				res := decodeAuthResponse(t, rec.Body)
				if res.Token == "" {
					t.Error("expected non-empty token")
				}
				if res.RefreshToken == "" {
					t.Error("expected non-empty refresh_token")
				}
			}
		})
	}
}

// ---- RefreshToken tests ----

func TestRefreshToken(t *testing.T) {
	secret := []byte("test-jwt-secret-key-32-bytes-long!")

	// Helper: issue a real refresh token and pre-populate the store.
	makeTokenAndStore := func(t *testing.T, userID int64) (string, *mockRefreshStore) {
		t.Helper()
		rt, err := auth.GenerateRefreshToken(userID, secret)
		if err != nil {
			t.Fatalf("GenerateRefreshToken() error: %v", err)
		}
		store := newMockRefreshStore()
		_ = store.Save(context.Background(), userID, rt, auth.RefreshTokenTTL)
		return rt, store
	}

	t.Run("bad request on invalid json", func(t *testing.T) {
		h := NewWithRepos(&mockUserRepo{}, nil, newMockRefreshStore(), secret)
		req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBufferString("{bad"))
		rec := httptest.NewRecorder()
		h.RefreshToken(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("bad request when refresh_token field is missing", func(t *testing.T) {
		h := NewWithRepos(&mockUserRepo{}, nil, newMockRefreshStore(), secret)
		b, _ := json.Marshal(map[string]string{})
		req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewReader(b))
		rec := httptest.NewRecorder()
		h.RefreshToken(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("unauthorized on invalid jwt", func(t *testing.T) {
		h := NewWithRepos(&mockUserRepo{}, nil, newMockRefreshStore(), secret)
		b, _ := json.Marshal(dto.RefreshRequest{RefreshToken: "not.a.jwt"})
		req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewReader(b))
		rec := httptest.NewRecorder()
		h.RefreshToken(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rec.Code)
		}
	})

	t.Run("unauthorized when access token provided instead of refresh token", func(t *testing.T) {
		accessToken, _ := auth.GenerateToken(1, secret)
		h := NewWithRepos(&mockUserRepo{}, nil, newMockRefreshStore(), secret)
		b, _ := json.Marshal(dto.RefreshRequest{RefreshToken: accessToken})
		req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewReader(b))
		rec := httptest.NewRecorder()
		h.RefreshToken(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d. Body: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("unauthorized when token not in store (revoked)", func(t *testing.T) {
		rt, _ := auth.GenerateRefreshToken(5, secret)
		// Do NOT save to store — simulates revoked/missing token.
		h := NewWithRepos(&mockUserRepo{}, nil, newMockRefreshStore(), secret)
		b, _ := json.Marshal(dto.RefreshRequest{RefreshToken: rt})
		req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewReader(b))
		rec := httptest.NewRecorder()
		h.RefreshToken(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rec.Code)
		}
	})

	t.Run("unauthorized when user no longer exists", func(t *testing.T) {
		rt, store := makeTokenAndStore(t, 99)
		repo := &mockUserRepo{
			getUserByIDFunc: func(_ context.Context, id int64) (models.User, error) {
				return models.User{}, errors.New("not found")
			},
		}
		h := NewWithRepos(repo, nil, store, secret)
		b, _ := json.Marshal(dto.RefreshRequest{RefreshToken: rt})
		req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewReader(b))
		rec := httptest.NewRecorder()
		h.RefreshToken(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rec.Code)
		}
	})

	t.Run("internal server error when store exists fails", func(t *testing.T) {
		rt, _ := auth.GenerateRefreshToken(7, secret)
		store := newMockRefreshStore()
		store.existsErr = errors.New("redis down")
		h := NewWithRepos(&mockUserRepo{}, nil, store, secret)
		b, _ := json.Marshal(dto.RefreshRequest{RefreshToken: rt})
		req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewReader(b))
		rec := httptest.NewRecorder()
		h.RefreshToken(rec, req)
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", rec.Code)
		}
	})

	t.Run("success issues new token pair and rotates refresh token", func(t *testing.T) {
		const userID = int64(7)
		rt, store := makeTokenAndStore(t, userID)

		repo := &mockUserRepo{
			getUserByIDFunc: func(_ context.Context, id int64) (models.User, error) {
				return models.User{ID: id, Name: "Alice", Email: "alice@example.com"}, nil
			},
		}
		h := NewWithRepos(repo, nil, store, secret)
		b, _ := json.Marshal(dto.RefreshRequest{RefreshToken: rt})
		req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewReader(b))
		rec := httptest.NewRecorder()
		h.RefreshToken(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d. Body: %s", rec.Code, rec.Body.String())
		}

		res := decodeAuthResponse(t, rec.Body)
		if res.Token == "" {
			t.Error("expected non-empty token in response")
		}
		if res.RefreshToken == "" {
			t.Error("expected non-empty refresh_token in response")
		}

		// The old refresh token must have been revoked (rotation).
		exists, _ := store.Exists(context.Background(), userID, rt)
		if exists {
			t.Error("old refresh token should have been revoked after rotation")
		}

		// The new refresh token must be in the store.
		exists, _ = store.Exists(context.Background(), userID, res.RefreshToken)
		if !exists {
			t.Error("new refresh token should be saved in the store")
		}

		// The old and new refresh tokens must be different.
		if res.RefreshToken == rt {
			t.Error("new refresh token should differ from the consumed one")
		}
	})

	t.Run("reuse of consumed refresh token is rejected", func(t *testing.T) {
		const userID = int64(8)
		rt, store := makeTokenAndStore(t, userID)

		repo := &mockUserRepo{
			getUserByIDFunc: func(_ context.Context, id int64) (models.User, error) {
				return models.User{ID: id, Name: "Bob", Email: "bob@example.com"}, nil
			},
		}
		h := NewWithRepos(repo, nil, store, secret)

		// First use — should succeed.
		b, _ := json.Marshal(dto.RefreshRequest{RefreshToken: rt})
		req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewReader(b))
		rec := httptest.NewRecorder()
		h.RefreshToken(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("first refresh expected 200, got %d", rec.Code)
		}

		// Second use of the same (now revoked) token — must be rejected.
		b, _ = json.Marshal(dto.RefreshRequest{RefreshToken: rt})
		req = httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewReader(b))
		rec = httptest.NewRecorder()
		h.RefreshToken(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("second refresh expected 401, got %d", rec.Code)
		}
	})
}
