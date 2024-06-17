package controllers_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"EtsyScraper/controllers"
	initializer "EtsyScraper/init"
	"EtsyScraper/models"
	setupMockServer "EtsyScraper/setupTests"
)

func TestAuthMiddleWareNoCookies(t *testing.T) {

	ctx, router, w := setupMockServer.SetGinTestMode()

	MockedUtils := &mockUtils{}
	UserRepo := &MockedUserRepository{}
	MockedUtils.On("GetTokens").Return(map[string]*models.Token{}, fmt.Errorf("failed to retrieve both tokens "))
	router.GET("/auth", controllers.AuthMiddleWare(MockedUtils, UserRepo))

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: ""})
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: ""})
	ctx.Request = req

	req = httptest.NewRequest("GET", "/auth", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
func TestAuthMiddleWareAccessTokenIsBlackListed(t *testing.T) {
	_, router, w := setupMockServer.SetGinTestMode()

	userID := uuid.New()
	MockedUtils := &mockUtils{}
	UserRepo := &MockedUserRepository{}
	MockedUtils.On("GetTokens").Return(map[string]*models.Token{}, nil)
	MockedUtils.On("ValidateJWT").Return(&models.CustomClaims{
		CreatedAt: 12344533,
		ExpiresAt: 12344533,
		UserUUID:  userID,
	}, nil)

	MockedUtils.On("IsJWTBlackListed").Return(true, nil)

	router.GET("/auth", controllers.AuthMiddleWare(MockedUtils, UserRepo))

	req := httptest.NewRequest("GET", "/auth", nil)
	req.Header.Set("Cookie", "access_token=SomeToken1")

	router.ServeHTTP(w, req)

	responseBody := w.Body.String()
	expectedFailMessage := "no refreshToken found"

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, responseBody, expectedFailMessage, "response body should contain the expected message")
}

func TestAuthMiddleWareAccessTokenIsBlackListedFalse(t *testing.T) {

	_, router, w := setupMockServer.SetGinTestMode()

	currentUserUUID := interface{}(nil)

	userID := uuid.New()
	MockedUtils := &mockUtils{}
	UserRepo := &MockedUserRepository{}

	token := models.NewToken("SomeToken")
	MockedUtils.On("GetTokens").Return(map[string]*models.Token{
		"access_token": token,
	}, nil)

	MockedUtils.On("ValidateJWT").Return(&models.CustomClaims{
		CreatedAt: 12344533,
		ExpiresAt: 12344533,
		UserUUID:  userID,
	}, nil)

	MockedUtils.On("IsJWTBlackListed").Return(false, nil)
	UserRepo.On("GetAccountByID").Return(&models.Account{ID: userID}, nil)

	router.GET("/auth", controllers.AuthMiddleWare(MockedUtils, UserRepo), func(ctx *gin.Context) {
		currentUserUUID, _ = ctx.Get("currentUserUUID")

	})

	req := httptest.NewRequest("GET", "/auth", nil)
	req.Header.Set("Cookie", "access_token=SomeToken1")

	router.ServeHTTP(w, req)

	assert.Equal(t, userID, currentUserUUID)
}

func TestAuthMiddleWareRefreshTokenError(t *testing.T) {

	_, router, w := setupMockServer.SetGinTestMode()

	MockedUtils := &mockUtils{}
	UserRepo := &MockedUserRepository{}
	MockedUtils.On("GetTokens").Return(map[string]*models.Token{}, nil)
	MockedUtils.On("ValidateJWT").Return(&models.CustomClaims{}, errors.New("not Valid"))
	MockedUtils.On("IsJWTBlackListed").Return(true, nil)

	router.GET("/auth", controllers.AuthMiddleWare(MockedUtils, UserRepo))

	req := httptest.NewRequest("GET", "/auth", nil)
	req.Header.Set("Cookie", "access_token=SomeToken1;")
	router.ServeHTTP(w, req)

	responseBody := w.Body.String()
	expectedFailMessage := "no refreshToken found"

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, responseBody, expectedFailMessage, "response body should contain the expected message")
}

func TestAuthMiddleWareRefreshAccessTokenFail(t *testing.T) {

	_, router, w := setupMockServer.SetGinTestMode()

	token := models.Token("")
	MockedUtils := &mockUtils{}
	UserRepo := &MockedUserRepository{}

	refToken := models.NewToken("SomeToken")

	MockedUtils.On("GetTokens").Return(map[string]*models.Token{"refresh_token": refToken}, nil)
	MockedUtils.On("ValidateJWT").Return(&models.CustomClaims{}, errors.New("not Valid"))
	MockedUtils.On("IsJWTBlackListed").Return(true, nil)
	MockedUtils.On("RefreshAccToken").Return(&token, errors.New("login required"))

	router.GET("/auth", controllers.AuthMiddleWare(MockedUtils, UserRepo))

	req := httptest.NewRequest("GET", "/auth", nil)
	req.Header.Set("Cookie", "access_token=SomeToken1; refresh_token=SomeToken2")
	router.ServeHTTP(w, req)

	responseBody := w.Body.String()
	expectedFailMessage := "login required"

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, responseBody, expectedFailMessage, "response body should contain the expected message")
}

func TestAuthMiddleWareValidateNewToken(t *testing.T) {

	_, router, w := setupMockServer.SetGinTestMode()

	token := models.Token("")

	MockedUtils := &mockUtils{}
	UserRepo := &MockedUserRepository{}

	refToken := models.NewToken("SomeToken")

	MockedUtils.On("GetTokens").Return(map[string]*models.Token{"refresh_token": refToken}, nil)
	MockedUtils.On("ValidateJWT").Return(&models.CustomClaims{}, errors.New("login required"))
	MockedUtils.On("IsJWTBlackListed").Return(true, nil)

	MockedUtils.On("RefreshAccToken").Return(&token, nil)
	MockedUtils.On("ValidateJWT").Return(&models.CustomClaims{}, errors.New("login required"))

	router.GET("/auth", controllers.AuthMiddleWare(MockedUtils, UserRepo))

	req := httptest.NewRequest("GET", "/auth", nil)
	req.Header.Set("Cookie", "access_token=SomeToken1; refresh_token=SomeToken2")
	router.ServeHTTP(w, req)

	responseBody := w.Body.String()
	expectedFailMessage := "login required"

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, responseBody, expectedFailMessage, "response body should contain the expected message")
}

func TestAuthMiddleWareNewCookie(t *testing.T) {

	_, router, w := setupMockServer.SetGinTestMode()

	token := models.Token("NoewRefreshedToken")
	user := uuid.New()

	MockedUtils := &mockUtils{}
	UserRepo := &MockedUserRepository{}

	MockedUtils.On("GetTokens").Return(map[string]*models.Token{"refresh_token": &token}, nil)
	MockedUtils.On("IsJWTBlackListed").Return(false, nil)
	MockedUtils.On("ValidateJWT").Return(&models.CustomClaims{
		CreatedAt: 12344533,
		ExpiresAt: 12344533,
		UserUUID:  user,
	}, nil)

	Account := &models.Account{
		ID: user,
	}

	MockedUtils.On("RefreshAccToken").Return(&token, nil)
	UserRepo.On("GetAccountByID").Return(Account, nil)
	router.GET("/auth", controllers.AuthMiddleWare(MockedUtils, UserRepo))

	req := httptest.NewRequest("GET", "/auth", nil)
	req.Header.Set("Cookie", "refresh_token=NoewRefreshedToken")
	router.ServeHTTP(w, req)

	cookies := w.Result().Cookies()
	var accessToken *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "access_token" {
			accessToken = cookie
		}
	}

	assert.Equal(t, string(token), accessToken.Value)
	assert.Equal(t, http.StatusOK, w.Code)

}
func TestAuthMiddleWareSuccessKeySet(t *testing.T) {

	_, router, w := setupMockServer.SetGinTestMode()

	token := models.Token("NoewRefreshedToken")
	user := uuid.New()
	var currentUserUUID uuid.UUID

	MockedUtils := &mockUtils{}
	UserRepo := &MockedUserRepository{}
	MockedUtils.On("GetTokens").Return(map[string]*models.Token{"refresh_token": &token}, nil)
	MockedUtils.On("IsJWTBlackListed").Return(false, nil)
	MockedUtils.On("RefreshAccToken").Return(&token, nil)
	MockedUtils.On("ValidateJWT").Return(&models.CustomClaims{
		CreatedAt: 12344533,
		ExpiresAt: 12344533,
		UserUUID:  user,
	}, nil)

	Account := &models.Account{
		ID: user,
	}

	UserRepo.On("GetAccountByID").Return(Account, nil)

	router.GET("/auth", controllers.AuthMiddleWare(MockedUtils, UserRepo), func(ctx *gin.Context) {
		currentUserUUID = ctx.MustGet("currentUserUUID").(uuid.UUID)
	})

	req := httptest.NewRequest("GET", "/auth", nil)
	req.Header.Set("Cookie", "refresh_token=SomeToken2")
	router.ServeHTTP(w, req)

	assert.Equal(t, user, currentUserUUID)
	assert.Equal(t, http.StatusOK, w.Code)

}

func TestAuthorizationNoCurrentUserPanic(t *testing.T) {
	_, testDB, _ := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	ctx, router, w := setupMockServer.SetGinTestMode()

	UserRepo := &MockedUserRepository{}

	router.Use(controllers.Authorization(UserRepo))

	router.GET("/test", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)

	assert.Panics(t, func() {
		ctx.Set("currentUserUUID", nil)
		router.ServeHTTP(w, req)
	})

}
func TestAuthorizationUserNotFound(t *testing.T) {

	c, router, w := setupMockServer.SetGinTestMode()

	UserRepo := &MockedUserRepository{}

	currentUserUUID := uuid.New()

	UserRepo.On("GetAccountByID").Return(nil, errors.New("No record found"))

	router.POST("/", func(ctx *gin.Context) {
		ctx.Set("currentUserUUID", currentUserUUID)

	}, controllers.Authorization(UserRepo))

	c.Request, _ = http.NewRequest("POST", "/", nil)

	router.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
func TestAuthorizationNotVerified(t *testing.T) {

	c, router, w := setupMockServer.SetGinTestMode()

	UserRepo := &MockedUserRepository{}
	currentUserUUID := uuid.New()

	Account := &models.Account{
		EmailVerified: false,
	}

	UserRepo.On("GetAccountByID").Return(Account, nil)

	router.POST("/", func(ctx *gin.Context) {
		ctx.Set("currentUserUUID", currentUserUUID)

	}, controllers.Authorization(UserRepo))

	c.Request, _ = http.NewRequest("POST", "/", nil)

	router.ServeHTTP(w, c.Request)

	assert.Contains(t, w.Body.String(), "email not verified")
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
func TestAuthorizationSuccess(t *testing.T) {

	c, router, w := setupMockServer.SetGinTestMode()

	UserRepo := &MockedUserRepository{}

	IsNextCalled := false
	currentUserUUID := uuid.New()

	Account := &models.Account{
		ID:            currentUserUUID,
		EmailVerified: true,
	}

	UserRepo.On("GetAccountByID").Return(Account, nil)

	router.POST("/", func(ctx *gin.Context) {
		ctx.Set("currentUserUUID", currentUserUUID)

	}, controllers.Authorization(UserRepo), func(ctx *gin.Context) {
		IsNextCalled = true
	})

	c.Request, _ = http.NewRequest("POST", "/", nil)

	router.ServeHTTP(w, c.Request)

	assert.True(t, IsNextCalled)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestIsAccountFollowingShopSuccess(t *testing.T) {

	c, router, w := setupMockServer.SetGinTestMode()

	UserRepo := &MockedUserRepository{}

	currentUserUUID := uuid.New()
	IsNextCalled := false
	ShopExample := models.Shop{Name: "Example"}
	ShopExample.ID = 1
	Account := &models.Account{
		ID:             currentUserUUID,
		ShopsFollowing: []models.Shop{ShopExample},
	}

	UserRepo.On("GetAccountWithShops").Return(Account, nil)

	router.POST("/:shopID", func(ctx *gin.Context) {
		ctx.Set("currentUserUUID", currentUserUUID)

	}, controllers.IsAccountFollowingShop(UserRepo), func(ctx *gin.Context) {
		IsNextCalled = true
	})

	c.Request, _ = http.NewRequest("POST", "/1", nil)

	router.ServeHTTP(w, c.Request)
	assert.True(t, IsNextCalled)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestIsAccountFollowingShopNoShopID(t *testing.T) {

	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	c, router, w := setupMockServer.SetGinTestMode()

	initializer.DB = MockedDataBase
	UserRepo := &MockedUserRepository{}

	currentUserUUID := uuid.New()

	router.POST("/:shopID", func(ctx *gin.Context) {
		ctx.Set("currentUserUUID", currentUserUUID)

	}, controllers.IsAccountFollowingShop(UserRepo))

	c.Request, _ = http.NewRequest("POST", "/", nil)

	router.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestIsAccountFollowingShopDataBaseError(t *testing.T) {

	c, router, w := setupMockServer.SetGinTestMode()

	UserRepo := &MockedUserRepository{}

	currentUserUUID := uuid.New()

	UserRepo.On("GetAccountWithShops").Return(nil, errors.New("internal error"))

	router.POST("/:shopID", func(ctx *gin.Context) {
		ctx.Set("currentUserUUID", currentUserUUID)

	}, controllers.IsAccountFollowingShop(UserRepo))

	c.Request, _ = http.NewRequest("POST", "/1", nil)

	router.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestIsAccountFollowingShopNotFollowing(t *testing.T) {

	c, router, w := setupMockServer.SetGinTestMode()

	UserRepo := &MockedUserRepository{}

	currentUserUUID := uuid.New()

	ShopExample := models.Shop{Name: "Example"}
	ShopExample.ID = uint(2)
	Account := &models.Account{
		ID:             currentUserUUID,
		ShopsFollowing: []models.Shop{ShopExample},
	}

	UserRepo.On("GetAccountWithShops").Return(Account, nil)

	router.POST("/:shopID", func(ctx *gin.Context) {
		ctx.Set("currentUserUUID", currentUserUUID)

	}, controllers.IsAccountFollowingShop(UserRepo))

	c.Request, _ = http.NewRequest("POST", "/1", nil)

	router.ServeHTTP(w, c.Request)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestIsAuthorized(t *testing.T) {

	c, _, w := setupMockServer.SetGinTestMode()

	UserRepo := &MockedUserRepository{}

	userID := uuid.New()
	Account := &models.Account{
		ID: userID,
	}

	UserRepo.On("GetAccountByID").Return(Account, nil)
	token := models.Token("anotherToken")

	MockedUtils := &mockUtils{}
	MockedUtils.On("IsJWTBlackListed").Return(false, nil)
	MockedUtils.On("ValidateJWT").Return(&models.CustomClaims{
		CreatedAt: 12344533,
		ExpiresAt: 12344533,
		UserUUID:  userID,
	}, nil)

	controllers.IsAuthorized(c, MockedUtils, &token, UserRepo)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestIsAuthorizedFailNoClaims(t *testing.T) {

	c, _, _ := setupMockServer.SetGinTestMode()
	UserRepo := &MockedUserRepository{}

	token := models.Token("anotherToken")

	MockedUtils := &mockUtils{}
	MockedUtils.On("ValidateJWT").Return(&models.CustomClaims{}, errors.New("some error"))

	result := controllers.IsAuthorized(c, MockedUtils, &token, UserRepo)

	assert.False(t, result)
}

func TestIsAuthorizedFailedIsBlackListed(t *testing.T) {

	c, _, _ := setupMockServer.SetGinTestMode()
	UserRepo := &MockedUserRepository{}

	user := uuid.New()

	token := models.Token("anotherToken")

	MockedUtils := &mockUtils{}
	MockedUtils.On("IsJWTBlackListed").Return(true, nil)
	MockedUtils.On("ValidateJWT").Return(&models.CustomClaims{
		CreatedAt: 12344533,
		ExpiresAt: 12344533,
		UserUUID:  user,
	}, nil)

	result := controllers.IsAuthorized(c, MockedUtils, &token, UserRepo)

	assert.False(t, result)
}
