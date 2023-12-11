package models

import (
	"github.com/golang-jwt/jwt/v4"
)

type Token string

type CustomClaims struct {
	CreatedAt int64
	ExpiresAt int64
}

func NewToken(token string) *Token {
	newToken := Token(token)
	return &newToken
}
func CreateClaims(claims jwt.MapClaims) *CustomClaims {

	return &CustomClaims{
		CreatedAt: int64(claims["createdAt"].(float64)),
		ExpiresAt: int64(claims["expiresAt"].(float64)),
	}
}
