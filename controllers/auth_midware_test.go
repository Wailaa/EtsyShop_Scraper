package controllers_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"

	"EtsyScraper/controllers"
	initializer "EtsyScraper/init"
	"EtsyScraper/models"
	setupMockServer "EtsyScraper/setupTests"
)

func TestAuthMiddleWare_noCookies(t *testing.T) {

	ctx, router, w := setupMockServer.SetGinTestMode()

	MockedUtils := &mockUtils{}
	MockedUtils.On("GetTokens").Return(map[string]*models.Token{}, fmt.Errorf("failed to retrieve both tokens "))
	router.GET("/auth", controllers.AuthMiddleWare(MockedUtils))

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: ""})
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: ""})
	ctx.Request = req

	req = httptest.NewRequest("GET", "/auth", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
func TestAuthMiddleWare_AccessToken_IsBlackListed(t *testing.T) {
	_, router, w := setupMockServer.SetGinTestMode()

	userID := uuid.New()
	MockedUtils := &mockUtils{}
	MockedUtils.On("GetTokens").Return(map[string]*models.Token{}, nil)
	MockedUtils.On("ValidateJWT").Return(&models.CustomClaims{
		CreatedAt: 12344533,
		ExpiresAt: 12344533,
		UserUUID:  userID,
	}, nil)

	MockedUtils.On("IsJWTBlackListed").Return(true, nil)

	router.GET("/auth", controllers.AuthMiddleWare(MockedUtils))

	req := httptest.NewRequest("GET", "/auth", nil)
	req.Header.Set("Cookie", "access_token=SomeToken1")

	router.ServeHTTP(w, req)

	responseBody := w.Body.String()
	expectedFailMessage := "no refreshToken found"

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, responseBody, expectedFailMessage, "response body should contain the expected message")
}

func TestAuthMiddleWare_AccessToken_IsBlackListed_False(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := setupMockServer.SetGinTestMode()

	initializer.DB = MockedDataBase
	currentUserUUID := interface{}(nil)
	Now := time.Now()
	userID := uuid.New()
	MockedUtils := &mockUtils{}

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

	Account := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "first_name", "last_name", "email", "password_hashed", "subscription_type", "email_verified", "email_verification_token", "request_change_pass", "account_pass_reset_token", "last_time_logged_in", "last_time_logged_out"}).
		AddRow(userID.String(), Now, Now, nil, "Testing", "User", "test@test1242q21.com", "", "free", false, "", false, "", Now, Now)

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE id = $1 AND "accounts"."deleted_at" IS NULL ORDER BY "accounts"."id" LIMIT $2`)).
		WithArgs(userID.String(), 1).WillReturnRows(Account)

	router.GET("/auth", controllers.AuthMiddleWare(MockedUtils), func(ctx *gin.Context) {
		currentUserUUID, _ = ctx.Get("currentUserUUID")

	})

	req := httptest.NewRequest("GET", "/auth", nil)
	req.Header.Set("Cookie", "access_token=SomeToken1")

	router.ServeHTTP(w, req)

	assert.Equal(t, userID, currentUserUUID)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestAuthMiddleWare_RefreshToken_Error(t *testing.T) {

	_, router, w := setupMockServer.SetGinTestMode()

	MockedUtils := &mockUtils{}
	MockedUtils.On("GetTokens").Return(map[string]*models.Token{}, nil)
	MockedUtils.On("ValidateJWT").Return(&models.CustomClaims{}, errors.New("not Valid"))
	MockedUtils.On("IsJWTBlackListed").Return(true, nil)

	router.GET("/auth", controllers.AuthMiddleWare(MockedUtils))

	req := httptest.NewRequest("GET", "/auth", nil)
	req.Header.Set("Cookie", "access_token=SomeToken1;")
	router.ServeHTTP(w, req)

	responseBody := w.Body.String()
	expectedFailMessage := "no refreshToken found"

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, responseBody, expectedFailMessage, "response body should contain the expected message")
}

func TestAuthMiddleWare_RefreshAccessToken_fail(t *testing.T) {

	_, router, w := setupMockServer.SetGinTestMode()

	token := models.Token("")
	MockedUtils := &mockUtils{}
	refToken := models.NewToken("SomeToken")

	MockedUtils.On("GetTokens").Return(map[string]*models.Token{"refresh_token": refToken}, nil)
	MockedUtils.On("ValidateJWT").Return(&models.CustomClaims{}, errors.New("not Valid"))
	MockedUtils.On("IsJWTBlackListed").Return(true, nil)
	MockedUtils.On("RefreshAccToken").Return(&token, errors.New("login required"))

	router.GET("/auth", controllers.AuthMiddleWare(MockedUtils))

	req := httptest.NewRequest("GET", "/auth", nil)
	req.Header.Set("Cookie", "access_token=SomeToken1; refresh_token=SomeToken2")
	router.ServeHTTP(w, req)

	responseBody := w.Body.String()
	expectedFailMessage := "login required"

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, responseBody, expectedFailMessage, "response body should contain the expected message")
}

func TestAuthMiddleWare_ValidateNewToken(t *testing.T) {

	_, router, w := setupMockServer.SetGinTestMode()

	token := models.Token("")

	MockedUtils := &mockUtils{}
	refToken := models.NewToken("SomeToken")

	MockedUtils.On("GetTokens").Return(map[string]*models.Token{"refresh_token": refToken}, nil)
	MockedUtils.On("ValidateJWT").Return(&models.CustomClaims{}, errors.New("login required"))
	MockedUtils.On("IsJWTBlackListed").Return(true, nil)

	MockedUtils.On("RefreshAccToken").Return(&token, nil)
	MockedUtils.On("ValidateJWT").Return(&models.CustomClaims{}, errors.New("login required"))

	router.GET("/auth", controllers.AuthMiddleWare(MockedUtils))

	req := httptest.NewRequest("GET", "/auth", nil)
	req.Header.Set("Cookie", "access_token=SomeToken1; refresh_token=SomeToken2")
	router.ServeHTTP(w, req)

	responseBody := w.Body.String()
	expectedFailMessage := "login required"

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, responseBody, expectedFailMessage, "response body should contain the expected message")
}

func TestAuthMiddleWare_NewCookie(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := setupMockServer.SetGinTestMode()

	initializer.DB = MockedDataBase

	token := models.Token("NoewRefreshedToken")
	user := uuid.New()
	Now := time.Now()

	MockedUtils := &mockUtils{}

	MockedUtils.On("GetTokens").Return(map[string]*models.Token{"refresh_token": &token}, nil)
	MockedUtils.On("IsJWTBlackListed").Return(false, nil)
	MockedUtils.On("ValidateJWT").Return(&models.CustomClaims{
		CreatedAt: 12344533,
		ExpiresAt: 12344533,
		UserUUID:  user,
	}, nil)
	MockedUtils.On("RefreshAccToken").Return(&token, nil)

	Account := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "first_name", "last_name", "email", "password_hashed", "subscription_type", "email_verified", "email_verification_token", "request_change_pass", "account_pass_reset_token", "last_time_logged_in", "last_time_logged_out"}).
		AddRow(user.String(), Now, Now, nil, "Testing", "User", "test@test1242q21.com", "", "free", false, "", false, "", Now, Now)

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE id = $1 AND "accounts"."deleted_at" IS NULL ORDER BY "accounts"."id" LIMIT $2`)).
		WithArgs(user.String(), 1).WillReturnRows(Account)

	router.GET("/auth", controllers.AuthMiddleWare(MockedUtils))

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
func TestAuthMiddleWare_Success_Key_Set(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := setupMockServer.SetGinTestMode()

	initializer.DB = MockedDataBase

	token := models.Token("NoewRefreshedToken")
	user := uuid.New()
	Now := time.Now()
	var currentUserUUID uuid.UUID

	MockedUtils := &mockUtils{}
	MockedUtils.On("GetTokens").Return(map[string]*models.Token{"refresh_token": &token}, nil)
	MockedUtils.On("IsJWTBlackListed").Return(false, nil)
	MockedUtils.On("RefreshAccToken").Return(&token, nil)
	MockedUtils.On("ValidateJWT").Return(&models.CustomClaims{
		CreatedAt: 12344533,
		ExpiresAt: 12344533,
		UserUUID:  user,
	}, nil)

	Account := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "first_name", "last_name", "email", "password_hashed", "subscription_type", "email_verified", "email_verification_token", "request_change_pass", "account_pass_reset_token", "last_time_logged_in", "last_time_logged_out"}).
		AddRow(user.String(), Now, Now, nil, "Testing", "User", "test@test1242q21.com", "", "free", false, "", false, "", Now, Now)

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE id = $1 AND "accounts"."deleted_at" IS NULL ORDER BY "accounts"."id" LIMIT $2`)).
		WithArgs(user.String(), 1).WillReturnRows(Account)

	router.GET("/auth", controllers.AuthMiddleWare(MockedUtils), func(ctx *gin.Context) {
		currentUserUUID = ctx.MustGet("currentUserUUID").(uuid.UUID)
	})

	req := httptest.NewRequest("GET", "/auth", nil)
	req.Header.Set("Cookie", "refresh_token=SomeToken2")
	router.ServeHTTP(w, req)

	assert.Equal(t, user, currentUserUUID)
	assert.Equal(t, http.StatusOK, w.Code)

}

func TestAuthorization_NoCurrentUser_Panic(t *testing.T) {
	_, testDB, _ := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	ctx, router, w := setupMockServer.SetGinTestMode()

	router.Use(controllers.Authorization())

	router.GET("/test", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)

	assert.Panics(t, func() {
		ctx.Set("currentUserUUID", nil)
		router.ServeHTTP(w, req)
	})

}
func TestAuthorization_UserNotFound(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	c, router, w := setupMockServer.SetGinTestMode()

	initializer.DB = MockedDataBase

	currentUserUUID := uuid.New()

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE id = $1 AND "accounts"."deleted_at" IS NULL ORDER BY "accounts"."id" LIMIT $2`)).
		WithArgs(currentUserUUID.String(), 1).WillReturnError(errors.New("No record found"))

	router.POST("/", func(ctx *gin.Context) {
		ctx.Set("currentUserUUID", currentUserUUID)

	}, controllers.Authorization())

	c.Request, _ = http.NewRequest("POST", "/", nil)

	router.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
func TestAuthorization_NotVerified(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	c, router, w := setupMockServer.SetGinTestMode()

	initializer.DB = MockedDataBase

	currentUserUUID := uuid.New()
	Now := time.Now()

	Account := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "first_name", "last_name", "email", "password_hashed", "subscription_type", "email_verified", "email_verification_token", "request_change_pass", "account_pass_reset_token", "last_time_logged_in", "last_time_logged_out"}).
		AddRow(currentUserUUID.String(), Now, Now, nil, "Testing", "User", "test@test1242q21.com", "", "free", false, "", false, "", Now, Now)

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE id = $1 AND "accounts"."deleted_at" IS NULL ORDER BY "accounts"."id" LIMIT $2`)).
		WithArgs(currentUserUUID.String(), 1).WillReturnRows(Account)

	router.POST("/", func(ctx *gin.Context) {
		ctx.Set("currentUserUUID", currentUserUUID)

	}, controllers.Authorization())

	c.Request, _ = http.NewRequest("POST", "/", nil)

	router.ServeHTTP(w, c.Request)

	assert.Contains(t, w.Body.String(), "email not verified")
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
func TestAuthorization_Success(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	c, router, w := setupMockServer.SetGinTestMode()

	initializer.DB = MockedDataBase

	IsNextCalled := false
	currentUserUUID := uuid.New()
	Now := time.Now()

	Account := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "first_name", "last_name", "email", "password_hashed", "subscription_type", "email_verified", "email_verification_token", "request_change_pass", "account_pass_reset_token", "last_time_logged_in", "last_time_logged_out"}).
		AddRow(currentUserUUID.String(), Now, Now, nil, "Testing", "User", "test@test1242q21.com", "", "free", true, "", false, "", Now, Now)

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE id = $1 AND "accounts"."deleted_at" IS NULL ORDER BY "accounts"."id" LIMIT $2`)).
		WithArgs(currentUserUUID.String(), 1).WillReturnRows(Account)

	router.POST("/", func(ctx *gin.Context) {
		ctx.Set("currentUserUUID", currentUserUUID)

	}, controllers.Authorization(), func(ctx *gin.Context) {
		IsNextCalled = true
	})

	c.Request, _ = http.NewRequest("POST", "/", nil)

	router.ServeHTTP(w, c.Request)

	assert.True(t, IsNextCalled)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestIsAccountFollowingShop_Success(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	c, router, w := setupMockServer.SetGinTestMode()

	initializer.DB = MockedDataBase

	currentUserUUID := uuid.New()
	ShopID := 1
	IsNextCalled := false

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE id = $1 AND "accounts"."deleted_at" IS NULL ORDER BY "accounts"."id" LIMIT $2`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(currentUserUUID.String()))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "account_shop_following" WHERE "account_shop_following"."account_id" = $1`)).
		WithArgs(currentUserUUID.String()).WillReturnRows(sqlmock.NewRows([]string{"shop_id", "account_id"}).AddRow(ShopID, currentUserUUID.String()))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "shops" WHERE "shops"."id" = $1 AND "shops"."deleted_at" IS NULL`)).
		WithArgs(ShopID).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(ShopID))

	router.POST("/:shopID", func(ctx *gin.Context) {
		ctx.Set("currentUserUUID", currentUserUUID)

	}, controllers.IsAccountFollowingShop(), func(ctx *gin.Context) {
		IsNextCalled = true
	})

	c.Request, _ = http.NewRequest("POST", "/1", nil)

	router.ServeHTTP(w, c.Request)
	assert.True(t, IsNextCalled)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestIsAccountFollowingShop_NoShopID(t *testing.T) {

	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	c, router, w := setupMockServer.SetGinTestMode()

	initializer.DB = MockedDataBase

	currentUserUUID := uuid.New()

	router.POST("/:shopID", func(ctx *gin.Context) {
		ctx.Set("currentUserUUID", currentUserUUID)

	}, controllers.IsAccountFollowingShop())

	c.Request, _ = http.NewRequest("POST", "/", nil)

	router.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestIsAccountFollowingShop_DataBaseError(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	c, router, w := setupMockServer.SetGinTestMode()

	initializer.DB = MockedDataBase

	currentUserUUID := uuid.New()

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE id = $1 AND "accounts"."deleted_at" IS NULL ORDER BY "accounts"."id" LIMIT $2`)).
		WillReturnError(errors.New("internal error"))

	router.POST("/:shopID", func(ctx *gin.Context) {
		ctx.Set("currentUserUUID", currentUserUUID)

	}, controllers.IsAccountFollowingShop())

	c.Request, _ = http.NewRequest("POST", "/1", nil)

	router.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestIsAccountFollowingShop_NotFollowing(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	c, router, w := setupMockServer.SetGinTestMode()

	initializer.DB = MockedDataBase

	currentUserUUID := uuid.New()
	ShopID := 2

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE id = $1 AND "accounts"."deleted_at" IS NULL ORDER BY "accounts"."id" LIMIT $2`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(currentUserUUID.String()))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "account_shop_following" WHERE "account_shop_following"."account_id" = $1`)).
		WithArgs(currentUserUUID.String()).WillReturnRows(sqlmock.NewRows([]string{"shop_id", "account_id"}).AddRow(ShopID, currentUserUUID.String()))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "shops" WHERE "shops"."id" = $1 AND "shops"."deleted_at" IS NULL`)).
		WithArgs(ShopID).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(ShopID))

	router.POST("/:shopID", func(ctx *gin.Context) {
		ctx.Set("currentUserUUID", currentUserUUID)

	}, controllers.IsAccountFollowingShop())

	c.Request, _ = http.NewRequest("POST", "/1", nil)

	router.ServeHTTP(w, c.Request)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestIsAuthorized(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	initializer.DB = MockedDataBase
	c, _, w := setupMockServer.SetGinTestMode()

	user := uuid.New()

	token := models.Token("anotherToken")
	Now := time.Now()
	MockedUtils := &mockUtils{}
	MockedUtils.On("IsJWTBlackListed").Return(false, nil)
	MockedUtils.On("ValidateJWT").Return(&models.CustomClaims{
		CreatedAt: 12344533,
		ExpiresAt: 12344533,
		UserUUID:  user,
	}, nil)

	Account := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "first_name", "last_name", "email", "password_hashed", "subscription_type", "email_verified", "email_verification_token", "request_change_pass", "account_pass_reset_token", "last_time_logged_in", "last_time_logged_out"}).
		AddRow(user.String(), Now, Now, nil, "Testing", "User", "test@test1242q21.com", "", "free", false, "", false, "", Now, Now)

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE id = $1 AND "accounts"."deleted_at" IS NULL ORDER BY "accounts"."id" LIMIT $2`)).
		WithArgs(user.String(), 1).WillReturnRows(Account)

	controllers.IsAuthorized(c, MockedUtils, &token)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestIsAuthorizedFailNoClaims(t *testing.T) {
	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	initializer.DB = MockedDataBase
	c, _, _ := setupMockServer.SetGinTestMode()

	token := models.Token("anotherToken")

	MockedUtils := &mockUtils{}
	MockedUtils.On("ValidateJWT").Return(&models.CustomClaims{}, errors.New("some error"))

	result := controllers.IsAuthorized(c, MockedUtils, &token)

	assert.False(t, result)
}

func TestIsAuthorizedFailedIsBlackListed(t *testing.T) {
	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	initializer.DB = MockedDataBase
	c, _, _ := setupMockServer.SetGinTestMode()

	user := uuid.New()

	token := models.Token("anotherToken")

	MockedUtils := &mockUtils{}
	MockedUtils.On("IsJWTBlackListed").Return(true, nil)
	MockedUtils.On("ValidateJWT").Return(&models.CustomClaims{
		CreatedAt: 12344533,
		ExpiresAt: 12344533,
		UserUUID:  user,
	}, nil)

	result := controllers.IsAuthorized(c, MockedUtils, &token)

	assert.False(t, result)
}
