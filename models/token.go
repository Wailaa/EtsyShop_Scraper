package models

import (
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

type Token string

type CustomClaims struct {
	CreatedAt int64
	ExpiresAt int64
	UserUUID  uuid.UUID
}

func NewToken(token string) *Token {
	newToken := Token(token)
	return &newToken
}
func CreateClaims(claims jwt.MapClaims) *CustomClaims {
	userUUID, _ := uuid.Parse(claims["userUUID"].(string))
	return &CustomClaims{
		CreatedAt: int64(claims["iat"].(float64)),
		ExpiresAt: int64(claims["exp"].(float64)),
		UserUUID:  userUUID,
	}
}
