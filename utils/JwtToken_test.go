package utils_test

import (
	initializer "EtsyScraper/init"
	"EtsyScraper/models"
	"EtsyScraper/utils"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCreateJwtToken_ValidClaims_ReturnsTokenObject(t *testing.T) {

	exp := time.Hour
	userUUID := uuid.New()

	token, err := utils.CreateJwtToken(exp, userUUID)

	assert.NoError(t, err)
	assert.NotNil(t, token)
}

func TestCreateJwtToken_FailedToGenerateToken_ReturnsError(t *testing.T) {

	exp := time.Hour
	userUUID := uuid.New()

	utils.Config.JwtSecret = ""
	token, err := utils.CreateJwtToken(exp, userUUID)

	assert.Error(t, err)
	assert.Nil(t, token)
}

func TestValidateJWT_ValidToken(t *testing.T) {
	exp := time.Hour
	userUUID := uuid.New()
	utils.Config = initializer.LoadProjConfig("../.")
	token, _ := utils.CreateJwtToken(exp, userUUID)

	claims, err := utils.ValidateJWT(token)

	assert.NoError(t, err)
	assert.NotNil(t, claims)
}

func TestValidateJWT_InvalidSignature(t *testing.T) {
	token := "invalid_token"
	jwtToken := models.Token(token)

	claims, err := utils.ValidateJWT(&jwtToken)

	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestRefreshAccToken_ValidRefreshToken(t *testing.T) {
	exp := time.Hour
	userUUID := uuid.New()

	token, _ := utils.CreateJwtToken(exp, userUUID)

	newAccessToken, err := utils.RefreshAccToken(token)

	assert.NoError(t, err)
	assert.NotNil(t, newAccessToken)
}

func TestRefreshAccToken_InvalidToken(t *testing.T) {

	refreshToken := models.NewToken("invalid_refresh_token")

	newAccessToken, err := utils.RefreshAccToken(refreshToken)

	assert.Error(t, err)
	assert.Nil(t, newAccessToken)
}

func TestIsJWTBlackListed_NotBlacklistedToken(t *testing.T) {
	exp := time.Millisecond
	userUUID := uuid.New()
	utils.Config = initializer.LoadProjConfig("../.") // Ensure this path is correct
	mockToken, err := utils.CreateJwtToken(exp, userUUID)
	if err != nil {
		t.Fatalf("Failed to create JWT token: %v", err)
	}
	initializer.RedisDBConnect(&utils.Config)

	blacklisted, err := utils.IsJWTBlackListed(mockToken)

	assert.False(t, blacklisted)
	assert.NoError(t, err)
}

func TestIsJWTBlackListed_Exists(t *testing.T) {

	exp := time.Hour
	userUUID := uuid.New()
	utils.Config = initializer.LoadProjConfig("../.") // Ensure this path is correct
	mockToken, err := utils.CreateJwtToken(exp, userUUID)
	if err != nil {
		t.Fatalf("Failed to create JWT token: %v", err)
	}
	initializer.RedisDBConnect(&utils.Config)

	utils.BlacklistJWT(mockToken)
	if err != nil {
		t.Fatalf("Failed to create JWT token: %v", err)
	}

	blacklisted, _ := utils.IsJWTBlackListed(mockToken)

	assert.True(t, blacklisted)
}

func TestBlacklistJWT_ValidToken(t *testing.T) {

	exp := time.Hour
	userUUID := uuid.New()
	utils.Config = initializer.LoadProjConfig("../.")
	mockToken, _ := utils.CreateJwtToken(exp, userUUID)

	initializer.RedisDBConnect(&utils.Config)

	err := utils.BlacklistJWT(mockToken)

	assert.NoError(t, err)

}

func TestBlacklistJWT_EmptyToken(t *testing.T) {

	utils.Config = initializer.LoadProjConfig("../.")
	newToken := models.Token("")

	initializer.RedisDBConnect(&utils.Config)

	err := utils.BlacklistJWT(&newToken)

	assert.Error(t, err)

}
func TestBlacklistJWT_TokenIsBlackListed(t *testing.T) {

	exp := time.Hour
	userUUID := uuid.New()
	utils.Config = initializer.LoadProjConfig("../.")
	mockToken, _ := utils.CreateJwtToken(exp, userUUID)

	initializer.RedisDBConnect(&utils.Config)

	_ = utils.BlacklistJWT(mockToken)
	err := utils.BlacklistJWT(mockToken)

	expectedErrorMessage := "token is alraedy Blacklisted"

	assert.EqualError(t, err, expectedErrorMessage)

	assert.Error(t, err)

}

func TestBlacklistJWT_TokenNotValid(t *testing.T) {

	utils.Config = initializer.LoadProjConfig("../.")
	newToken := models.Token("test")

	initializer.RedisDBConnect(&utils.Config)

	err := utils.BlacklistJWT(&newToken)

	expectedErrorMessage := "invalidate token: token contains an invalid number of segments"
	assert.EqualError(t, err, expectedErrorMessage)
	assert.Error(t, err)

}
