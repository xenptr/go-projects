package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// AccessTokenTTL is the lifetime of a short-lived access token.
const AccessTokenTTL = 15 * time.Minute

// RefreshTokenTTL is the lifetime of a long-lived refresh token.
const RefreshTokenTTL = 7 * 24 * time.Hour

// refreshClaims extends RegisteredClaims with a token-type discriminator
// so refresh tokens can never be accepted where access tokens are expected.
type refreshClaims struct {
	jwt.RegisteredClaims
	TokenType string `json:"type"`
}

const tokenTypeRefresh = "refresh"

func GenerateToken(userID int64, secret []byte) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   strconv.FormatInt(userID, 10),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(AccessTokenTTL)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := t.SignedString(secret)
	if err != nil {
		return "", err
	}
	return s, nil
}

func ParseToken(tokenString string, secret []byte) (int64, error) {
	var claims jwt.RegisteredClaims

	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{"HS256"}),
	)

	keyFunc := func(token *jwt.Token) (any, error) {
		return secret, nil
	}

	_, err := parser.ParseWithClaims(tokenString, &claims, keyFunc)
	if err != nil {
		return 0, err
	}

	userID, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil {
		return 0, err
	}

	return userID, nil
}

// GenerateRefreshToken creates a long-lived JWT that carries a "type":"refresh"
// claim, preventing it from being used as an access token. A random JTI
// (JWT ID) is embedded so every issued token is unique.
func GenerateRefreshToken(userID int64, secret []byte) (string, error) {
	jti, err := generateJTI()
	if err != nil {
		return "", err
	}

	claims := refreshClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			Subject:   strconv.FormatInt(userID, 10),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(RefreshTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		TokenType: tokenTypeRefresh,
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := t.SignedString(secret)
	if err != nil {
		return "", err
	}
	return s, nil
}

// ParseRefreshToken validates and parses a refresh token, returning the userID.
// It rejects tokens that are not explicitly typed as "refresh".
func ParseRefreshToken(tokenString string, secret []byte) (int64, error) {
	var claims refreshClaims

	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{"HS256"}),
	)

	keyFunc := func(token *jwt.Token) (any, error) {
		return secret, nil
	}

	_, err := parser.ParseWithClaims(tokenString, &claims, keyFunc)
	if err != nil {
		return 0, err
	}

	if claims.TokenType != tokenTypeRefresh {
		return 0, errors.New("invalid token type")
	}

	userID, err := strconv.ParseInt(claims.RegisteredClaims.Subject, 10, 64)
	if err != nil {
		return 0, err
	}

	return userID, nil
}

// generateJTI returns a cryptographically random 16-byte hex string.
func generateJTI() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
