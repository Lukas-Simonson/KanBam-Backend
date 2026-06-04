package jwtutil

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func Sign(id string, secret string) (string, error) {
	signer := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.RegisteredClaims{
			Subject:   id,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(14 * 24 * time.Hour)),
		},
	)
	return signer.SignedString([]byte(secret))
}

func Parse(tokenString string, secret string) (string, error) {
	claims := jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && token.Valid {
		return claims.Subject, nil
	}

	return "", fmt.Errorf("invalid token claims")
}
