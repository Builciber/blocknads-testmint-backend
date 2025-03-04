package auth

import (
	"fmt"

	jwt "github.com/golang-jwt/jwt/v5"
)

func ValidateJWT(sessionTokenString, sessionSecret string) (string, error) {
	tokenClaims := jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(
		sessionTokenString,
		&tokenClaims,
		func(t *jwt.Token) (any, error) { return []byte(sessionSecret), nil })
	if err != nil {
		return "", fmt.Errorf("session is invalid or expired")
	}
	subject, err := token.Claims.GetSubject()
	if err != nil {
		return "", err
	}
	return subject, nil
}
