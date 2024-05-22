package utils

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"EtsyScraper/models"
)

type Utils struct {
}

type UtilsProcess interface {
	CreateVerificationString() (string, error)
	SendVerificationEmail(account *models.Account) error
	SendResetPassEmail(account *models.Account) error
	HashPass(pass string) (string, error)
	IsPassVerified(pass string, hashedPass string) bool
	CreateJwtToken(exp time.Duration, userUUID uuid.UUID) (*models.Token, error)
	ValidateJWT(JWTToken *models.Token) (*models.CustomClaims, error)
	RefreshAccToken(token *models.Token) (*models.Token, error)
	BlacklistJWT(token *models.Token) error
	IsJWTBlackListed(token *models.Token) (bool, error)
	PickProxyProvider() ProxySetting
	GetRandomUserAgent() string
	GetTokens(ctx *gin.Context) (map[string]*models.Token, error)
}
