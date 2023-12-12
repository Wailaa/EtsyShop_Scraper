package utils

import (
	initializer "EtsyScraper/init"
	"EtsyScraper/models"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var TokenBlacklistPrefix = "token:blacklist:"

func CreateJwtToken(exp time.Duration) (*models.Token, error) {

	now := time.Now().UTC()

	config := initializer.LoadProjConfig(".")

	JWTSecret := config.JwtSecret

	claims := jwt.MapClaims{
		"createdAt": now.Unix(),
		"expiresAt": now.Add(exp).Unix(),
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	jwtTokenString, err := jwtToken.SignedString([]byte(JWTSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to generate JWT Token with the following code %w", err)
	}

	token := models.NewToken(jwtTokenString)
	return token, nil

}

func ValidateJWT(JWTToken string) (*models.CustomClaims, error) {

	config := initializer.LoadProjConfig(".")

	parcedtoken, err := jwt.Parse(JWTToken, func(Token *jwt.Token) (interface{}, error) {
		if _, ok := Token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected method: %s", Token.Header["alg"])
		}
		return []byte(config.JwtSecret), nil
	})
	if err != nil {
		if !strings.Contains(err.Error(), "used before issued") {
			return nil, fmt.Errorf("invalidate token: %w", err)
		}
	}

	getClaimsData, ok := parcedtoken.Claims.(jwt.MapClaims)
	getClaims := models.CreateClaims(getClaimsData)

	if err := getClaimsData.Valid(); err != nil || !ok {
		return nil, fmt.Errorf("invalid claims %v", ok)
	}

	return getClaims, nil
}

func BlacklistJWT(token string) error {

	if token == "" {
		return fmt.Errorf("token is missing")
	}

	context := context.TODO()

	BlacklistedJWT, err := ValidateJWT(token)
	if err != nil {
		return err
	}

	expiredToken := TokenBlacklistPrefix + token

	Now := time.Now().UTC()
	EX := time.Unix(BlacklistedJWT.ExpiresAt, 0)
	tokenExpire := EX.Sub(Now)

	errToken := initializer.RedisClient.Set(context, expiredToken, "revokedToken", tokenExpire).Err()
	if errToken != nil {
		return fmt.Errorf(errToken.Error())
	}

	return nil
}

func IsJWTBlackListed(token string) (bool, error) {
	context := context.TODO()
	blacklistedToken := TokenBlacklistPrefix + token

	checkInBlackList := initializer.RedisClient.Exists(context, blacklistedToken)
	if checkInBlackList.Err() != nil {
		return false, fmt.Errorf("error while checking blacklist: %v", checkInBlackList.Err())
	}

	if checkInBlackList.Val() > 0 {
		return true, nil
	}

	return false, nil
}
