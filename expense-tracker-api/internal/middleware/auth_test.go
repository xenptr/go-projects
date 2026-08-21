package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/xenptr/go-projects/expense-tracker-api/internal/auth"
)

func TestAuth(t *testing.T) {
	secret := []byte("test-jwt-secret-key-32-bytes-long!")
	wrongSecret := []byte("wrong-jwt-secret-key-32-bytes-long!")

	validToken, err := auth.GenerateToken(42, secret)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	wrongSecretToken, err := auth.GenerateToken(42, wrongSecret)
	if err != nil {
		t.Fatalf("failed to generate token with wrong secret: %v", err)
	}

	expiredClaims := jwt.RegisteredClaims{
		Subject:   "42",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
	}
	expiredJwt := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims)
	expiredToken, err := expiredJwt.SignedString(secret)
	if err != nil {
		t.Fatalf("failed to generate expired token: %v", err)
	}

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectNext     bool
		expectedUserID int64
	}{
		{
			name:           "missing authorization header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectNext:     false,
		},
		{
			name:           "invalid scheme basic",
			authHeader:     "Basic dXNlcjpwYXNz",
			expectedStatus: http.StatusUnauthorized,
			expectNext:     false,
		},
		{
			name:           "bearer prefix without space",
			authHeader:     "BearerToken",
			expectedStatus: http.StatusUnauthorized,
			expectNext:     false,
		},
		{
			name:           "invalid jwt token string",
			authHeader:     "Bearer invalid.token.payload",
			expectedStatus: http.StatusUnauthorized,
			expectNext:     false,
		},
		{
			name:           "token signed with different secret",
			authHeader:     "Bearer " + wrongSecretToken,
			expectedStatus: http.StatusUnauthorized,
			expectNext:     false,
		},
		{
			name:           "expired token",
			authHeader:     "Bearer " + expiredToken,
			expectedStatus: http.StatusUnauthorized,
			expectNext:     false,
		},
		{
			name:           "valid token",
			authHeader:     "Bearer " + validToken,
			expectedStatus: http.StatusOK,
			expectNext:     true,
			expectedUserID: 42,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calledNext := false
			var receivedUserID int64

			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calledNext = true
				if uid, ok := auth.GetUserID(r); ok {
					receivedUserID = uid
				}
				w.WriteHeader(http.StatusOK)
			})

			mw := Auth(secret)(nextHandler)

			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			rec := httptest.NewRecorder()
			mw.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Fatalf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			if calledNext != tt.expectNext {
				t.Fatalf("calledNext = %v, want %v", calledNext, tt.expectNext)
			}

			if tt.expectNext && receivedUserID != tt.expectedUserID {
				t.Fatalf("receivedUserID = %d, want %d", receivedUserID, tt.expectedUserID)
			}
		})
	}
}
