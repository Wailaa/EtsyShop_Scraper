package utils_test

import (
	"net/http"
	"net/http/httptest"
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

func TestCreateJwtTokenValidClaimsReturnsTokenObject(t *testing.T) {
	jwt, exp, userUUID := InitiateUtilTest()

	token, err := jwt.CreateJwtToken(exp, userUUID)

	assert.NoError(t, err)
	assert.NotNil(t, token)
}

func TestCreateJwtTokenFailedToGenerateTokenReturnsError(t *testing.T) {
	jwt, exp, userUUID := InitiateUtilTest()

	utils.Config.JwtSecret = ""
	token, err := jwt.CreateJwtToken(exp, userUUID)

	assert.Error(t, err)
	assert.Nil(t, token)
}

func TestValidateJWTValidToken(t *testing.T) {
	jwt, exp, userUUID := InitiateUtilTest()

	token, _ := jwt.CreateJwtToken(exp, userUUID)

	claims, err := jwt.ValidateJWT(token)

	assert.NoError(t, err)
	assert.NotNil(t, claims)
}

func TestValidateJWTInvalidSignature(t *testing.T) {
	jwt, _, _ := InitiateUtilTest()

	token := "invalid_token"
	jwtToken := models.Token(token)

	claims, err := jwt.ValidateJWT(&jwtToken)

	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestRefreshAccTokenValidRefreshToken(t *testing.T) {
	jwt, exp, userUUID := InitiateUtilTest()
	token, _ := jwt.CreateJwtToken(exp, userUUID)

	newAccessToken, err := jwt.RefreshAccToken(token)

	assert.NoError(t, err)
	assert.NotNil(t, newAccessToken)
}

func TestRefreshAccTokenInvalidToken(t *testing.T) {
	jwt := utils.Utils{}
	refreshToken := models.NewToken("invalid_refresh_token")

	newAccessToken, err := jwt.RefreshAccToken(refreshToken)

	assert.Error(t, err)
	assert.Nil(t, newAccessToken)
}

func TestIsJWTBlackListedNotBlacklistedToken(t *testing.T) {
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

func TestIsJWTBlackListedExists(t *testing.T) {
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

func TestBlacklistJWTValidToken(t *testing.T) {
	jwt, exp, userUUID := InitiateUtilTest()

	mockToken, _ := jwt.CreateJwtToken(exp, userUUID)

	initializer.RedisDBConnect(&utils.Config)

	err := jwt.BlacklistJWT(mockToken)

	assert.NoError(t, err)

}

func TestBlacklistJWTEmptyToken(t *testing.T) {
	jwt, _, _ := InitiateUtilTest()

	newToken := models.Token("")

	initializer.RedisDBConnect(&utils.Config)

	err := jwt.BlacklistJWT(&newToken)

	assert.Error(t, err)

}
func TestBlacklistJWTTokenIsBlackListed(t *testing.T) {
	jwt, exp, userUUID := InitiateUtilTest()

	mockToken, _ := jwt.CreateJwtToken(exp, userUUID)

	initializer.RedisDBConnect(&utils.Config)

	_ = jwt.BlacklistJWT(mockToken)
	err := jwt.BlacklistJWT(mockToken)

	expectedErrorMessage := "error: token is alraedy Blacklisted"

	assert.EqualError(t, err, expectedErrorMessage)

	assert.Error(t, err)

}

func TestBlacklistJWTTokenNotValid(t *testing.T) {
	jwt, _, _ := InitiateUtilTest()

	newToken := models.Token("test")

	initializer.RedisDBConnect(&utils.Config)

	err := jwt.BlacklistJWT(&newToken)

	expectedErrorMessage := "error while blacklisting token: error: invalidate token: token contains an invalid number of segments"
	assert.EqualError(t, err, expectedErrorMessage)
	assert.Error(t, err)

}

func TestGetTokensSuccess(t *testing.T) {
	ctx, _, _ := setupMockServer.SetGinTestMode()
	jwt := utils.Utils{}
	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: "access_token_value"})
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "refresh_token_value"})
	ctx.Request = req

	tokens, _ := jwt.GetTokens(ctx)

	assert.Equal(t, 2, len(tokens))

}

func TestGetTokensFail(t *testing.T) {
	ctx, _, _ := setupMockServer.SetGinTestMode()
	jwt := utils.Utils{}
	req := httptest.NewRequest("GET", "/", nil)
	ctx.Request = req

	_, err := jwt.GetTokens(ctx)

	assert.Error(t, err)

}
