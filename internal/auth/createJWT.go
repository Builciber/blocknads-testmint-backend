package auth

import (
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

func CreateJWT(discordUserID string, sessionSecret string) (string, error) {
	currentTimeInUTC := time.Now().UTC()
	expirationTimeinUTC := currentTimeInUTC.Add(30 * time.Minute)
	sessionClaims := &jwt.RegisteredClaims{
		Issuer:    "mint-session",
		IssuedAt:  jwt.NewNumericDate(currentTimeInUTC),
		ExpiresAt: jwt.NewNumericDate(expirationTimeinUTC),
		Subject:   discordUserID,
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, sessionClaims)
	signedSessionToken, err := accessToken.SignedString([]byte(sessionSecret))
	if err != nil {
		return "", err
	}
	return signedSessionToken, nil
}
