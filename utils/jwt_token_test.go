package utils_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	initializer "EtsyScraper/init"
	"EtsyScraper/models"
	setupMockServer "EtsyScraper/setupTests"
	"EtsyScraper/utils"
)

func InitiateUtilTest() (jwt *utils.Utils, exp time.Duration, userUUID uuid.UUID) {
	jwt = &utils.Utils{}
	exp = time.Hour
	userUUID = uuid.New()
	utils.Config = initializer.LoadProjConfig("../")
	return
}

func TestCreateJwtToken_ValidClaims_ReturnsTokenObject(t *testing.T) {
	jwt, exp, userUUID := InitiateUtilTest()

	token, err := jwt.CreateJwtToken(exp, userUUID)

	assert.NoError(t, err)
	assert.NotNil(t, token)
}

func TestCreateJwtToken_FailedToGenerateToken_ReturnsError(t *testing.T) {
	jwt, exp, userUUID := InitiateUtilTest()

	utils.Config.JwtSecret = ""
	token, err := jwt.CreateJwtToken(exp, userUUID)

	assert.Error(t, err)
	assert.Nil(t, token)
}

func TestValidateJWT_ValidToken(t *testing.T) {
	jwt, exp, userUUID := InitiateUtilTest()

	token, _ := jwt.CreateJwtToken(exp, userUUID)

	claims, err := jwt.ValidateJWT(token)

	assert.NoError(t, err)
	assert.NotNil(t, claims)
}

func TestValidateJWT_InvalidSignature(t *testing.T) {
	jwt, _, _ := InitiateUtilTest()

	token := "invalid_token"
	jwtToken := models.Token(token)

	claims, err := jwt.ValidateJWT(&jwtToken)

	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestRefreshAccToken_ValidRefreshToken(t *testing.T) {
	jwt, exp, userUUID := InitiateUtilTest()
	token, _ := jwt.CreateJwtToken(exp, userUUID)

	newAccessToken, err := jwt.RefreshAccToken(token)

	assert.NoError(t, err)
	assert.NotNil(t, newAccessToken)
}

func TestRefreshAccToken_InvalidToken(t *testing.T) {
	jwt := utils.Utils{}
	refreshToken := models.NewToken("invalid_refresh_token")

	newAccessToken, err := jwt.RefreshAccToken(refreshToken)

	assert.Error(t, err)
	assert.Nil(t, newAccessToken)
}

func TestIsJWTBlackListed_NotBlacklistedToken(t *testing.T) {
	jwt, exp, userUUID := InitiateUtilTest()

	mockToken, err := jwt.CreateJwtToken(exp, userUUID)
	if err != nil {
		t.Fatalf("Failed to create JWT token: %v", err)
	}
	initializer.RedisDBConnect(&utils.Config)

	blacklisted, err := jwt.IsJWTBlackListed(mockToken)

	assert.False(t, blacklisted)
	assert.NoError(t, err)
}

func TestIsJWTBlackListed_Exists(t *testing.T) {
	jwt, exp, userUUID := InitiateUtilTest()

	mockToken, err := jwt.CreateJwtToken(exp, userUUID)
	if err != nil {
		t.Fatalf("Failed to create JWT token: %v", err)
	}
	initializer.RedisDBConnect(&utils.Config)

	jwt.BlacklistJWT(mockToken)
	if err != nil {
		t.Fatalf("Failed to create JWT token: %v", err)
	}

	blacklisted, _ := jwt.IsJWTBlackListed(mockToken)

	assert.True(t, blacklisted)
}

func TestBlacklistJWT_ValidToken(t *testing.T) {
	jwt, exp, userUUID := InitiateUtilTest()

	mockToken, _ := jwt.CreateJwtToken(exp, userUUID)

	initializer.RedisDBConnect(&utils.Config)

	err := jwt.BlacklistJWT(mockToken)

	assert.NoError(t, err)

}

func TestBlacklistJWT_EmptyToken(t *testing.T) {
	jwt, _, _ := InitiateUtilTest()

	newToken := models.Token("")

	initializer.RedisDBConnect(&utils.Config)

	err := jwt.BlacklistJWT(&newToken)

	assert.Error(t, err)

}
func TestBlacklistJWT_TokenIsBlackListed(t *testing.T) {
	jwt, exp, userUUID := InitiateUtilTest()

	mockToken, _ := jwt.CreateJwtToken(exp, userUUID)

	initializer.RedisDBConnect(&utils.Config)

	_ = jwt.BlacklistJWT(mockToken)
	err := jwt.BlacklistJWT(mockToken)

	expectedErrorMessage := "error: token is alraedy Blacklisted"

	assert.EqualError(t, err, expectedErrorMessage)

	assert.Error(t, err)

}

func TestBlacklistJWT_TokenNotValid(t *testing.T) {
	jwt, _, _ := InitiateUtilTest()

	newToken := models.Token("test")

	initializer.RedisDBConnect(&utils.Config)

	err := jwt.BlacklistJWT(&newToken)

	expectedErrorMessage := "error while blacklisting token: error: invalidate token: token contains an invalid number of segments"
	assert.EqualError(t, err, expectedErrorMessage)
	assert.Error(t, err)

}
