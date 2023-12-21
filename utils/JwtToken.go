package utils

import (
	initializer "EtsyScraper/init"
	"EtsyScraper/models"
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

var TokenBlacklistPrefix = "token:blacklist:"

func CreateJwtToken(exp time.Duration, userUUID uuid.UUID) (*models.Token, error) {

	now := time.Now().UTC()

	config := initializer.LoadProjConfig(".")

	JWTSecret := config.JwtSecret

	claims := jwt.MapClaims{
		"iat":      now.Unix(),
		"exp":      now.Add(exp).Unix(),
		"userUUID": userUUID,
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	jwtTokenString, err := jwtToken.SignedString([]byte(JWTSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to generate JWT Token with the following code %w", err)
	}

	token := models.NewToken(jwtTokenString)
	return token, nil

}

func ValidateJWT(JWTToken *models.Token) (*models.CustomClaims, error) {

	config := initializer.LoadProjConfig(".")
	token := fmt.Sprint(*JWTToken)
	parcedtoken, err := jwt.Parse(token, func(Token *jwt.Token) (interface{}, error) {
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

func BlacklistJWT(token *models.Token) error {

	if token == models.NewToken("") {
		return fmt.Errorf("token is missing")
	}

	context := context.TODO()

	checkBlackList, err := IsJWTBlackListed(token)
	if checkBlackList {
		return fmt.Errorf("token is alraedy Blacklisted")
	}
	if err != nil {
		return err
	}

	BlacklistedJWT, err := ValidateJWT(token)
	if err != nil {
		log.Println("error while blacklisting token", err)
		return err
	}

	expiredToken := TokenBlacklistPrefix + fmt.Sprint(token)

	Now := time.Now().UTC()
	EX := time.Unix(BlacklistedJWT.ExpiresAt, 0)
	tokenExpire := EX.Sub(Now)

	errToken := initializer.RedisClient.Set(context, expiredToken, "revokedToken", tokenExpire).Err()
	if errToken != nil {
		return fmt.Errorf(errToken.Error())
	}

	return nil
}

func IsJWTBlackListed(token *models.Token) (bool, error) {
	context := context.TODO()
	blacklistedToken := TokenBlacklistPrefix + fmt.Sprint(token)

	checkInBlackList := initializer.RedisClient.Exists(context, blacklistedToken)
	if checkInBlackList.Err() != nil {
		return false, fmt.Errorf("error while checking blacklist: %v", checkInBlackList.Err())
	}

	if checkInBlackList.Val() > 0 {
		return true, nil
	}

	return false, nil
}
