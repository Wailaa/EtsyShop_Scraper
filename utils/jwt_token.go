package utils

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"

	initializer "EtsyScraper/init"
	"EtsyScraper/models"
)

var TokenBlacklistPrefix = "token:blacklist:"

func (ut *Utils) CreateJwtToken(exp time.Duration, userUUID uuid.UUID) (*models.Token, error) {

	now := time.Now().UTC()

	JWTSecret := Config.JwtSecret
	if JWTSecret == "" {
		err := fmt.Errorf("failed to generate JWT Token with the current short key")
		return nil, HandleError(err)
	}

	claims := jwt.MapClaims{
		"iat":      now.Unix(),
		"exp":      now.Add(exp).Unix(),
		"userUUID": userUUID,
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	jwtTokenString, err := jwtToken.SignedString([]byte(JWTSecret))
	if err != nil {
		return nil, HandleError(err, "failed to generate JWT Token with the following code")
	}

	token := models.NewToken(jwtTokenString)
	return token, nil

}

func (ut *Utils) ValidateJWT(JWTToken *models.Token) (*models.CustomClaims, error) {

	token := fmt.Sprint(*JWTToken)
	parcedtoken, err := jwt.Parse(token, func(Token *jwt.Token) (interface{}, error) {
		if _, ok := Token.Method.(*jwt.SigningMethodHMAC); !ok {
			err := fmt.Errorf("unexpected method: %s", Token.Header["alg"])
			return nil, HandleError(err)
		}
		return []byte(Config.JwtSecret), nil
	})
	if err != nil {
		if !strings.Contains(err.Error(), "used before issued") {
			err := fmt.Errorf("invalidate token: %w", err)
			return nil, HandleError(err)
		}
	}

	ClaimsData, ok := parcedtoken.Claims.(jwt.MapClaims)
	if err := ClaimsData.Valid(); err != nil || !ok {
		message := fmt.Sprintf("invalid claims %v", ok)
		return nil, HandleError(err, message)
	}

	Claims := models.CreateClaims(ClaimsData)

	return Claims, nil
}

func (ut *Utils) RefreshAccToken(token *models.Token) (*models.Token, error) {

	refreshTokenClaims, err := ut.ValidateJWT(token)
	if err != nil {
		return nil, HandleError(err)
	}

	newAccessToken, err := ut.CreateJwtToken(Config.AccTokenExp, refreshTokenClaims.UserUUID)
	if err != nil {
		return nil, HandleError(err)
	}

	return newAccessToken, nil
}

func (ut *Utils) BlacklistJWT(token *models.Token) error {

	if token == models.NewToken("") {
		err := fmt.Errorf("token is missing")
		return HandleError(err)
	}

	context := context.TODO()

	isBlackListed, err := ut.IsJWTBlackListed(token)
	if isBlackListed {
		err := fmt.Errorf("token is alraedy Blacklisted")
		return HandleError(err)
	}
	if err != nil {
		return HandleError(err)
	}

	Claims, err := ut.ValidateJWT(token)
	if err != nil {
		return HandleError(err, "error while blacklisting token")
	}

	expiredToken := TokenBlacklistPrefix + fmt.Sprint(*token)

	Now := time.Now().UTC()
	EX := time.Unix(Claims.ExpiresAt, 0)
	tokenExpire := EX.Sub(Now)

	errToken := initializer.RedisClient.Set(context, expiredToken, "revokedToken", tokenExpire).Err()
	if errToken != nil {
		return HandleError(errToken)
	}

	return nil
}

func (ut *Utils) IsJWTBlackListed(token *models.Token) (bool, error) {
	context := context.TODO()
	blacklistedToken := TokenBlacklistPrefix + fmt.Sprint(*token)

	checkInBlackList := initializer.RedisClient.Exists(context, blacklistedToken)
	if checkInBlackList.Err() != nil {
		err := fmt.Errorf("error while checking blacklist: %v", checkInBlackList.Err())
		return false, HandleError(err)
	}

	if checkInBlackList.Val() > 0 {
		return true, nil
	}

	return false, nil
}
func (ut *Utils) GetTokens(ctx *gin.Context) (map[string]*models.Token, error) {

	tokens := make(map[string]*models.Token)
	if accessToken, err := ctx.Cookie("access_token"); err == nil {
		tokens["access_token"] = models.NewToken(accessToken)
	}
	if refreshToken, err := ctx.Cookie("refresh_token"); err == nil {
		tokens["refresh_token"] = models.NewToken(refreshToken)
	}
	if len(tokens) == 0 {
		err := fmt.Errorf("failed to retrieve both tokens ")
		return nil, HandleError(err)
	}
	return tokens, nil
}
