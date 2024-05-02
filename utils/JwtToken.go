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

func (ut *Utils) CreateJwtToken(exp time.Duration, userUUID uuid.UUID) (*models.Token, error) {

	now := time.Now().UTC()

	JWTSecret := Config.JwtSecret
	if JWTSecret == "" {
		return nil, fmt.Errorf("failed to generate JWT Token with the current short key")
	}

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

func (ut *Utils) ValidateJWT(JWTToken *models.Token) (*models.CustomClaims, error) {

	token := fmt.Sprint(*JWTToken)
	parcedtoken, err := jwt.Parse(token, func(Token *jwt.Token) (interface{}, error) {
		if _, ok := Token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected method: %s", Token.Header["alg"])
		}
		return []byte(Config.JwtSecret), nil
	})
	if err != nil {
		if !strings.Contains(err.Error(), "used before issued") {
			return nil, fmt.Errorf("invalidate token: %w", err)
		}
	}

	ClaimsData, ok := parcedtoken.Claims.(jwt.MapClaims)
	if err := ClaimsData.Valid(); err != nil || !ok {
		return nil, fmt.Errorf("invalid claims %v", ok)
	}

	Claims := models.CreateClaims(ClaimsData)

	return Claims, nil
}

func (ut *Utils) RefreshAccToken(token *models.Token) (*models.Token, error) {

	refreshTokenClaims, err := ut.ValidateJWT(token)
	if err != nil {
		return nil, err
	}

	newAccessToken, err := ut.CreateJwtToken(Config.AccTokenExp, refreshTokenClaims.UserUUID)
	if err != nil {
		return nil, err
	}

	return newAccessToken, nil
}

func (ut *Utils) BlacklistJWT(token *models.Token) error {

	if token == models.NewToken("") {
		return fmt.Errorf("token is missing")
	}

	context := context.TODO()

	isBlackListed, err := ut.IsJWTBlackListed(token)
	if isBlackListed {
		return fmt.Errorf("token is alraedy Blacklisted")
	}
	if err != nil {
		return err
	}

	Claims, err := ut.ValidateJWT(token)
	if err != nil {
		log.Println("error while blacklisting token", err)
		return err
	}

	expiredToken := TokenBlacklistPrefix + fmt.Sprint(*token)

	Now := time.Now().UTC()
	EX := time.Unix(Claims.ExpiresAt, 0)
	tokenExpire := EX.Sub(Now)

	errToken := initializer.RedisClient.Set(context, expiredToken, "revokedToken", tokenExpire).Err()
	if errToken != nil {
		return fmt.Errorf(errToken.Error())
	}

	return nil
}

func (ut *Utils) IsJWTBlackListed(token *models.Token) (bool, error) {
	context := context.TODO()
	blacklistedToken := TokenBlacklistPrefix + fmt.Sprint(*token)

	checkInBlackList := initializer.RedisClient.Exists(context, blacklistedToken)
	if checkInBlackList.Err() != nil {
		return false, fmt.Errorf("error while checking blacklist: %v", checkInBlackList.Err())
	}

	if checkInBlackList.Val() > 0 {
		return true, nil
	}

	return false, nil
}
