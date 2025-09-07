package utils

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var SECRET_KEY string = os.Getenv("SECRET_KEY")

func GenerateJwt(issuer string) (string, error) {

	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.RegisteredClaims{
		Issuer:    issuer,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 72)), // 3 days
	})

	token, err := claims.SignedString([]byte(SECRET_KEY))

	return token, err
}

func VerifyJwt(cookie string) (string, error) {

	token, err := jwt.ParseWithClaims(cookie, &jwt.RegisteredClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(SECRET_KEY), nil
	})

	if err != nil || !token.Valid {
		return "", err
	}

	claims := token.Claims.(*jwt.RegisteredClaims)

	return claims.Issuer, nil
}
