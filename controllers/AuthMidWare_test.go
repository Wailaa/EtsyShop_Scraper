package controllers_test

import (
	"EtsyScraper/controllers"
	initializer "EtsyScraper/init"
	"EtsyScraper/models"
	setupMockServer "EtsyScraper/setupTests"
	"errors"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestAuthMiddleWare_noCookies(t *testing.T) {

	ctx, router, w := SetGinTestMode()

	MockedUtils := &mockUtils{}
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
	_, router, w := SetGinTestMode()

	userID := uuid.New()
	MockedUtils := &mockUtils{}

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

	_, router, w := SetGinTestMode()

	initializer.DB = MockedDataBase

	currentUserUUID := interface{}(nil)
	Now := time.Now()
	userID := uuid.New()
	MockedUtils := &mockUtils{}

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

	_, router, w := SetGinTestMode()

	MockedUtils := &mockUtils{}
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

	_, router, w := SetGinTestMode()

	token := models.Token("")
	MockedUtils := &mockUtils{}
	MockedUtils.On("ValidateJWT").Return(&models.CustomClaims{}, errors.New("not Valid"))
	MockedUtils.On("IsJWTBlackListed").Return(true, nil)
	MockedUtils.On("RefreshAccToken").Return(&token, errors.New("failed to refresh access token"))

	router.GET("/auth", controllers.AuthMiddleWare(MockedUtils))

	req := httptest.NewRequest("GET", "/auth", nil)
	req.Header.Set("Cookie", "access_token=SomeToken1; refresh_token=SomeToken2")
	router.ServeHTTP(w, req)

	responseBody := w.Body.String()
	expectedFailMessage := "failed to refresh access token"

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, responseBody, expectedFailMessage, "response body should contain the expected message")
}

func TestAuthMiddleWare_ValidateNewToken(t *testing.T) {

	_, router, w := SetGinTestMode()

	token := models.Token("")

	MockedUtils := &mockUtils{}
	MockedUtils.On("ValidateJWT").Return(&models.CustomClaims{}, errors.New("not Valid"))
	MockedUtils.On("IsJWTBlackListed").Return(true, nil)

	MockedUtils.On("RefreshAccToken").Return(&token, nil)
	MockedUtils.On("ValidateJWT").Return(&models.CustomClaims{}, errors.New("not Valid"))

	router.GET("/auth", controllers.AuthMiddleWare(MockedUtils))

	req := httptest.NewRequest("GET", "/auth", nil)
	req.Header.Set("Cookie", "access_token=SomeToken1; refresh_token=SomeToken2")
	router.ServeHTTP(w, req)

	responseBody := w.Body.String()
	expectedFailMessage := "auth failed because of not Valid"

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, responseBody, expectedFailMessage, "response body should contain the expected message")
}

func TestAuthMiddleWare_NewCookie(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := SetGinTestMode()

	initializer.DB = MockedDataBase

	token := models.Token("NoewRefreshedToken")
	user := uuid.New()
	Now := time.Now()

	MockedUtils := &mockUtils{}

	MockedUtils.On("IsJWTBlackListed").Return(true, nil)

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

	router.GET("/auth", controllers.AuthMiddleWare(MockedUtils))

	req := httptest.NewRequest("GET", "/auth", nil)
	req.Header.Set("Cookie", "access_token=SomeToken1; refresh_token=SomeToken2")
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

	_, router, w := SetGinTestMode()

	initializer.DB = MockedDataBase

	token := models.Token("NoewRefreshedToken")
	user := uuid.New()
	Now := time.Now()
	var currentUserUUID uuid.UUID

	MockedUtils := &mockUtils{}

	MockedUtils.On("IsJWTBlackListed").Return(true, nil)

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
	req.Header.Set("Cookie", "access_token=SomeToken1; refresh_token=SomeToken2")
	router.ServeHTTP(w, req)

	assert.Equal(t, user, currentUserUUID)
	assert.Equal(t, http.StatusOK, w.Code)

}

func TestAuthorization_NoCurrentUser_Panic(t *testing.T) {
	_, testDB, _ := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	ctx, router, w := SetGinTestMode()

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

	c, router, w := SetGinTestMode()

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

	c, router, w := SetGinTestMode()

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

	c, router, w := SetGinTestMode()

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
