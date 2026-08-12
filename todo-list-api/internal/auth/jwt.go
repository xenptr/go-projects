package auth

import (
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateToken(userID int64, secret []byte) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   strconv.FormatInt(userID, 10),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
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
