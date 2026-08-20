package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateToken(t *testing.T) {
	userID := int64(123)
	secret := []byte("test-secret")

	token, err := GenerateToken(userID, secret)
	if err != nil {
		t.Fatalf("GenerateToken() returned error: %v", err)
	}

	if token == "" {
		t.Fatal("GenerateToken() returned an empty token")
	}
}

func TestGenerateTokenClaims(t *testing.T) {
	userID := int64(123)
	secret := []byte("test-secret")

	tokenString, err := GenerateToken(userID, secret)
	if err != nil {
		t.Fatalf("GenerateToken() returned error: %v", err)
	}

	token, err := jwt.ParseWithClaims(
		tokenString,
		&jwt.RegisteredClaims{},
		func(token *jwt.Token) (any, error) {
			return secret, nil
		},
	)
	if err != nil {
		t.Fatalf("failed to parse generated token: %v", err)
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		t.Fatal("failed to get registered claims")
	}

	if claims.Subject != "123" {
		t.Errorf("Subject = %q, want %q", claims.Subject, "123")
	}

	if claims.IssuedAt == nil {
		t.Fatal("IssuedAt is nil")
	}

	if claims.ExpiresAt == nil {
		t.Fatal("ExpiresAt is nil")
	}

	if claims.ExpiresAt.Before(time.Now()) {
		t.Error("generated token is already expired")
	}

	if !claims.ExpiresAt.After(claims.IssuedAt.Time) {
		t.Error("ExpiresAt should be after IssuedAt")
	}
}

func TestParseToken(t *testing.T) {
	userID := int64(123)
	secret := []byte("test-secret")

	token, err := GenerateToken(userID, secret)
	if err != nil {
		t.Fatalf("GenerateToken() returned error: %v", err)
	}

	gotUserID, err := ParseToken(token, secret)
	if err != nil {
		t.Fatalf("ParseToken() returned error: %v", err)
	}

	if gotUserID != userID {
		t.Errorf("ParseToken() = %d, want %d", gotUserID, userID)
	}
}

func TestParseTokenWrongSecret(t *testing.T) {
	secret := []byte("test-secret")
	wrongSecret := []byte("wrong-secret")

	token, err := GenerateToken(123, secret)
	if err != nil {
		t.Fatalf("GenerateToken() returned error: %v", err)
	}

	_, err = ParseToken(token, wrongSecret)
	if err == nil {
		t.Fatal("ParseToken() expected error with wrong secret, got nil")
	}
}

func TestParseTokenInvalidToken(t *testing.T) {
	secret := []byte("test-secret")

	_, err := ParseToken("invalid-token", secret)
	if err == nil {
		t.Fatal("ParseToken() expected error for invalid token, got nil")
	}
}

func TestParseTokenExpiredToken(t *testing.T) {
	secret := []byte("test-secret")

	claims := jwt.RegisteredClaims{
		Subject:   "123",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(secret)
	if err != nil {
		t.Fatalf("failed to sign expired token: %v", err)
	}

	_, err = ParseToken(tokenString, secret)
	if err == nil {
		t.Fatal("ParseToken() expected error for expired token, got nil")
	}
}

func TestParseTokenInvalidSubject(t *testing.T) {
	secret := []byte("test-secret")

	claims := jwt.RegisteredClaims{
		Subject:   "not-a-number",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(secret)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	_, err = ParseToken(tokenString, secret)
	if err == nil {
		t.Fatal("ParseToken() expected error for invalid subject, got nil")
	}
}

// ---- Refresh token tests ----

func TestGenerateRefreshToken(t *testing.T) {
	userID := int64(42)
	secret := []byte("test-secret")

	token, err := GenerateRefreshToken(userID, secret)
	if err != nil {
		t.Fatalf("GenerateRefreshToken() returned error: %v", err)
	}
	if token == "" {
		t.Fatal("GenerateRefreshToken() returned empty token")
	}
}

func TestGenerateRefreshTokenClaims(t *testing.T) {
	userID := int64(42)
	secret := []byte("test-secret")

	tokenString, err := GenerateRefreshToken(userID, secret)
	if err != nil {
		t.Fatalf("GenerateRefreshToken() returned error: %v", err)
	}

	var claims refreshClaims
	_, err = jwt.ParseWithClaims(tokenString, &claims, func(t *jwt.Token) (any, error) {
		return secret, nil
	})
	if err != nil {
		t.Fatalf("failed to parse refresh token: %v", err)
	}

	if claims.Subject != "42" {
		t.Errorf("Subject = %q, want %q", claims.Subject, "42")
	}
	if claims.TokenType != tokenTypeRefresh {
		t.Errorf("TokenType = %q, want %q", claims.TokenType, tokenTypeRefresh)
	}
	if claims.ExpiresAt == nil || claims.ExpiresAt.Before(time.Now()) {
		t.Error("refresh token is already expired")
	}
	// Refresh token should live significantly longer than an access token.
	minExpiry := time.Now().Add(RefreshTokenTTL - time.Minute)
	if claims.ExpiresAt.Before(minExpiry) {
		t.Errorf("refresh token TTL too short: expires %v", claims.ExpiresAt)
	}
}

func TestParseRefreshToken(t *testing.T) {
	userID := int64(42)
	secret := []byte("test-secret")

	tokenString, err := GenerateRefreshToken(userID, secret)
	if err != nil {
		t.Fatalf("GenerateRefreshToken() returned error: %v", err)
	}

	gotUserID, err := ParseRefreshToken(tokenString, secret)
	if err != nil {
		t.Fatalf("ParseRefreshToken() returned error: %v", err)
	}
	if gotUserID != userID {
		t.Errorf("ParseRefreshToken() = %d, want %d", gotUserID, userID)
	}
}

func TestParseRefreshTokenRejectsAccessToken(t *testing.T) {
	secret := []byte("test-secret")

	// An access token must not be accepted by ParseRefreshToken.
	accessToken, err := GenerateToken(42, secret)
	if err != nil {
		t.Fatalf("GenerateToken() returned error: %v", err)
	}

	_, err = ParseRefreshToken(accessToken, secret)
	if err == nil {
		t.Fatal("ParseRefreshToken() expected error when given an access token, got nil")
	}
}

func TestParseRefreshTokenRejectsWrongSecret(t *testing.T) {
	secret := []byte("test-secret")
	wrongSecret := []byte("wrong-secret")

	tokenString, err := GenerateRefreshToken(42, secret)
	if err != nil {
		t.Fatalf("GenerateRefreshToken() returned error: %v", err)
	}

	_, err = ParseRefreshToken(tokenString, wrongSecret)
	if err == nil {
		t.Fatal("ParseRefreshToken() expected error with wrong secret, got nil")
	}
}

func TestParseRefreshTokenExpired(t *testing.T) {
	secret := []byte("test-secret")

	claims := refreshClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "42",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
		TokenType: tokenTypeRefresh,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secret)
	if err != nil {
		t.Fatalf("failed to sign expired refresh token: %v", err)
	}

	_, err = ParseRefreshToken(tokenString, secret)
	if err == nil {
		t.Fatal("ParseRefreshToken() expected error for expired token, got nil")
	}
}
