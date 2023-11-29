package utils

import (
	initializer "EtsyScraper/init"
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt"
)

func CreateJwtToken(exp int64) (string, error) {
	now := time.Now().UTC()

	config, err := initializer.LoadProjConfig(".")
	if err != nil {
		log.Fatal("Could not load environment variables", err)

	}
	JWTSecret := config.JwtSecret

	claims := &jwt.MapClaims{
		"iat":       now,
		"expiresAt": now.Add(time.Hour * time.Duration(exp)),
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	jwtTokenString, err := jwtToken.SignedString([]byte(JWTSecret))
	if err != nil {
		return "", fmt.Errorf("failed to generate JWT Token with the following code %w", err)
	}

	return jwtTokenString, nil

}
