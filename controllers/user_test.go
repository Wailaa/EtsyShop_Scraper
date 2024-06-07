package controllers_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"

	"EtsyScraper/controllers"
	"EtsyScraper/models"
	setupMockServer "EtsyScraper/setupTests"
	"EtsyScraper/utils"
)

type mockUtils struct{ mock.Mock }

func (m *mockUtils) CreateVerificationString() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *mockUtils) SendVerificationEmail(account *models.Account) error {
	return nil
}

func (m *mockUtils) SendResetPassEmail(account *models.Account) error {

	return nil
}

func (m *mockUtils) HashPass(pass string) (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *mockUtils) IsPassVerified(pass string, hashedPass string) bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *mockUtils) CreateJwtToken(exp time.Duration, userUUID uuid.UUID) (*models.Token, error) {

	args := m.Called()
	return args.Get(0).(*models.Token), args.Error(1)
}

func (m *mockUtils) ValidateJWT(JWTToken *models.Token) (*models.CustomClaims, error) {
	args := m.Called()
	return args.Get(0).(*models.CustomClaims), args.Error(1)
}

func (m *mockUtils) RefreshAccToken(token *models.Token) (*models.Token, error) {
	args := m.Called()
	return args.Get(0).(*models.Token), args.Error(1)
}

func (m *mockUtils) BlacklistJWT(token *models.Token) error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockUtils) IsJWTBlackListed(token *models.Token) (bool, error) {
	args := m.Called()
	return args.Bool(0), args.Error(1)
}
func (m *mockUtils) PickProxyProvider() utils.ProxySetting {

	return utils.ProxySetting{}
}
func (m *mockUtils) GetRandomUserAgent() string {

	return ""
}
func (m *mockUtils) GetTokens(ctx *gin.Context) (map[string]*models.Token, error) {

	args := m.Called()
	return args.Get(0).(map[string]*models.Token), args.Error(1)
}

type MockedUserRepository struct {
	mock.Mock
}

func (mr *MockedUserRepository) GetAccountByID(ID uuid.UUID) (account *models.Account, err error) {

	args := mr.Called()
	UserRepo := args.Get(0)
	var Account *models.Account
	if UserRepo != nil {
		Account = UserRepo.(*models.Account)
	}
	return Account, args.Error(1)

}

func (mr *MockedUserRepository) GetAccountByEmail(email string) *models.Account {
	args := mr.Called()
	return args.Get(0).(*models.Account)
}

func (mr *MockedUserRepository) UpdateLastTimeLoggedIn(Account *models.Account) error {
	args := mr.Called()
	return args.Error(0)
}
func (mr *MockedUserRepository) JoinShopFollowing(Account *models.Account) error {
	args := mr.Called()
	return args.Error(0)
}
func (mr *MockedUserRepository) UpdateLastTimeLoggedOut(UserID uuid.UUID) error {
	args := mr.Called()
	return args.Error(0)
}
func (mr *MockedUserRepository) UpdateAccountAfterVerify(Account *models.Account) error {
	args := mr.Called()
	return args.Error(0)
}
func (mr *MockedUserRepository) UpdateAccountNewPass(Account *models.Account, passwardHashed string) error {
	args := mr.Called()
	return args.Error(0)
}
func (mr *MockedUserRepository) UpdateAccountAfterResetPass(Account *models.Account, newPasswardHashed string) error {
	args := mr.Called()
	return args.Error(0)
}
func (mr *MockedUserRepository) SaveAccount(Account *models.Account) error {
	args := mr.Called()
	return args.Error(0)
}

func TestRegisterUserInvalidJson(t *testing.T) {
	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := setupMockServer.SetGinTestMode()

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedDataBase, MockedUtils, nil)

	router.POST("/register", User.RegisterUser)

	body := []byte(`{"invalid-json": "facek-data"}`)
	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, w.Code, http.StatusBadRequest)

}

func TestRegisterUserPassNoMatch(t *testing.T) {
	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := setupMockServer.SetGinTestMode()

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedDataBase, MockedUtils, nil)

	router.POST("/register", User.RegisterUser)

	body := []byte(`{
		"first_name":"Testing",
		"last_name" : "User",
		"email":"test@test.com",
		"password": "1111qqqq",
		"password_confirm":"2222wwww"
		"subscription_type":"free"
		}`)
	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, w.Code, http.StatusBadRequest)
}

func TestRegisterUserPassShort(t *testing.T) {
	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := setupMockServer.SetGinTestMode()

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedDataBase, MockedUtils, nil)

	router.POST("/register", User.RegisterUser)

	body := []byte(`{
		"first_name":"Testing",
		"last_name" : "User",
		"email":"test@test.com",
		"password": "111",
		"password_confirm":"222"
		"subscription_type":"free"
		}`)
	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, w.Code, http.StatusBadRequest)
}

func TestRegisterUserHashPassErr(t *testing.T) {
	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := setupMockServer.SetGinTestMode()

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedDataBase, MockedUtils, nil)

	expectedError := errors.New("expected error")

	MockedUtils.On("HashPass").Return("", expectedError)
	router.POST("/register", User.RegisterUser)

	body := []byte(`{
		"first_name":"Testing",
		"last_name":"User",
		"email":"test@test12421.com",
		"password":"1111wwww",
		"password_confirm":"1111wwww",
		"subscription_type":"free"
		}`)

	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, w.Code, http.StatusConflict)
}

func TestRegisterUserVerificationStringErr(t *testing.T) {
	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := setupMockServer.SetGinTestMode()

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedDataBase, MockedUtils, nil)

	expectedError := errors.New("expected error")

	MockedUtils.On("HashPass").Return("", nil)
	MockedUtils.On("CreateVerificationString").Return("", expectedError)

	router.POST("/register", User.RegisterUser)

	body := []byte(`{
		"first_name":"Testing",
		"last_name":"User",
		"email":"test@test12421.com",
		"password":"1111wwwww",
		"password_confirm":"1111wwwww",
		"subscription_type":"free"
		}`)

	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, w.Code, http.StatusConflict)
}

func TestRegisterUserDataBaseError(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, _ := setupMockServer.SetGinTestMode()

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedDataBase, MockedUtils, nil)

	ErrorCases := []string{"email is in use", "some other error"}

	router.POST("/register", User.RegisterUser)
	for _, Error := range ErrorCases {
		sqlMock.ExpectBegin()
		sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "accounts" ("id","created_at","updated_at","deleted_at","first_name","last_name","email","password_hashed","subscription_type","email_verified","email_verification_token","request_change_pass","account_pass_reset_token","last_time_logged_in","last_time_logged_out") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15) RETURNING "id"`)).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), "Testing", "User", "test@test1242q21.com", "", "free", false, "", false, "", sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnError(errors.New(Error))
		sqlMock.ExpectRollback()

		MockedUtils.On("HashPass").Return("", nil)
		MockedUtils.On("CreateVerificationString").Return("", nil)

		body := []byte(`{
		"first_name":"Testing",
		"last_name":"User",
		"email":"test@test1242q21.com",
		"password":"1111wwwww",
		"password_confirm":"1111wwwww",
		"subscription_type":"free"
		}`)

		req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
		assert.NoError(t, sqlMock.ExpectationsWereMet())
	}
}

func TestRegisterUserExpectCreateAccount(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := setupMockServer.SetGinTestMode()

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedDataBase, MockedUtils, nil)

	MockedUtils.On("HashPass").Return("", nil)
	MockedUtils.On("CreateVerificationString").Return("", nil)

	sqlMock.ExpectBegin()
	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "accounts" ("id","created_at","updated_at","deleted_at","first_name","last_name","email","password_hashed","subscription_type","email_verified","email_verification_token","request_change_pass","account_pass_reset_token","last_time_logged_in","last_time_logged_out") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15) RETURNING "id"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), "Testing", "User", "test@test1242q21.com", "", "free", false, "", false, "", sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"1", "15"}))
	sqlMock.ExpectCommit()

	router.POST("/register", User.RegisterUser)

	body := []byte(`{
		"first_name":"Testing",
		"last_name":"User",
		"email":"test@test1242q21.com",
		"password":"1111wwwww",
		"password_confirm":"1111wwwww",
		"subscription_type":"free"
		}`)

	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, sqlMock.ExpectationsWereMet())

}

func TestLoginAccountInvalidJson(t *testing.T) {
	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := setupMockServer.SetGinTestMode()

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedDataBase, MockedUtils, nil)

	router.POST("/login", User.LoginAccount)

	body := []byte(`{"invalid-json": "facek-data"}`)
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, w.Code, http.StatusNotFound)

}

func TestLoginAccountInvalidEmailEmpty(t *testing.T) {
	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := setupMockServer.SetGinTestMode()

	MockedUtils := &mockUtils{}
	UserRepo := &MockedUserRepository{}
	emptyAccount := &models.Account{}

	User := controllers.NewUserController(MockedDataBase, MockedUtils, UserRepo)

	UserRepo.On("GetAccountByEmail").Return(emptyAccount, errors.New("Account was not found"))

	router.POST("/login", User.LoginAccount)

	body := []byte(`{
		"email": "Test@Test.com",
		"password": "1234qwer"
	  }`)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

}

func TestLoginAccountPassVerifiedfail(t *testing.T) {
	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := setupMockServer.SetGinTestMode()

	MockedUtils := &mockUtils{}
	UserRepo := &MockedUserRepository{}

	email := "Test@Test.com"
	user := uuid.New()

	User := controllers.NewUserController(MockedDataBase, MockedUtils, UserRepo)

	UserRepo.On("GetAccountByEmail").Return(&models.Account{ID: user, Email: email}, nil)
	MockedUtils.On("IsPassVerified").Return(false)

	router.POST("/login", User.LoginAccount)

	body := []byte(`{
		"email": "Test@Test.com",
		"password": "111122222"
	  }`)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, w.Code, http.StatusNotFound)

}

func TestLoginAccountAccessToken(t *testing.T) {
	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := setupMockServer.SetGinTestMode()

	expectedError := errors.New("Error while creating Token")

	MockedUtils := &mockUtils{}
	UserRepo := &MockedUserRepository{}
	User := controllers.NewUserController(MockedDataBase, MockedUtils, UserRepo)

	email := "Test@Test.com"
	user := uuid.New()

	UserRepo.On("GetAccountByEmail").Return(&models.Account{ID: user, Email: email}, nil)

	MockedUtils.On("IsPassVerified").Return(true)
	MockedUtils.On("CreateJwtToken").Return(models.NewToken(""), expectedError)

	router.POST("/login", User.LoginAccount)

	body := []byte(`{
		"email": "Test@Test.com",
		"password": "111122222"
	  }`)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, w.Code, http.StatusBadRequest)

}

func TestLoginAccountAccessTokenfailed(t *testing.T) {
	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := setupMockServer.SetGinTestMode()

	expectedError := errors.New("Error while creating Token")

	MockedUtils := &mockUtils{}
	UserRepo := &MockedUserRepository{}
	email := "Test@Test.com"
	user := uuid.New()

	User := controllers.NewUserController(MockedDataBase, MockedUtils, UserRepo)

	UserRepo.On("GetAccountByEmail").Return(&models.Account{ID: user, Email: email}, nil)
	MockedUtils.On("IsPassVerified").Return(true)
	MockedUtils.On("CreateJwtToken").Return(models.NewToken(""), expectedError)

	router.POST("/login", User.LoginAccount)

	body := []byte(`{
		"email": "Test@Test.com",
		"password": "111122222"
	  }`)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

}

func TestLoginAccountRefreshTokenFailed(t *testing.T) {
	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := setupMockServer.SetGinTestMode()

	expectedError := errors.New("Error while creating Token")

	MockedUtils := &mockUtils{}
	UserRepo := &MockedUserRepository{}
	email := "Test@Test.com"
	user := uuid.New()

	User := controllers.NewUserController(MockedDataBase, MockedUtils, UserRepo)

	TokenExp := 12 * time.Second
	UserRepo.On("GetAccountByEmail").Return(&models.Account{ID: user, Email: email}, nil)
	MockedUtils.On("IsPassVerified").Return(true)
	MockedUtils.On("CreateJwtToken", TokenExp, mock.Anything).Return(models.NewToken("Token"), nil)
	MockedUtils.On("CreateJwtToken").Return(models.NewToken(""), expectedError)

	router.POST("/login", User.LoginAccount)

	body := []byte(`{
		"email": "Test@Test.com",
		"password": "111122222"
	  }`)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

}

func TestLoginAccountSuccess(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := setupMockServer.SetGinTestMode()

	MockedUtils := &mockUtils{}
	UserRepo := &MockedUserRepository{}
	email := "Test@Test.com"
	user := uuid.New()

	User := controllers.NewUserController(MockedDataBase, MockedUtils, UserRepo)

	UserRepo.On("GetAccountByEmail").Return(&models.Account{ID: user, Email: email}, nil)
	MockedUtils.On("IsPassVerified").Return(true)
	MockedUtils.On("CreateJwtToken").Return(models.NewToken("Token"), nil)
	MockedUtils.On("CreateJwtToken").Return(models.NewToken("Token"), nil)

	ShopFollowing := sqlmock.NewRows([]string{"account_id", "shop_id"}).AddRow(user.String(), 1)

	shops := sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "Shop1").AddRow(2, "Shop2")

	sqlMock.ExpectBegin()
	sqlMock.ExpectExec(regexp.QuoteMeta(`UPDATE "accounts" SET "last_time_logged_in"=$1,"updated_at"=$2 WHERE id = $3 AND "accounts"."deleted_at" IS NULL AND "id" = $4`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), user, user).WillReturnResult(sqlmock.NewResult(1, 1))
	sqlMock.ExpectCommit()

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE "accounts"."id" = $1 AND "accounts"."deleted_at" IS NULL AND "accounts"."id" = $2 ORDER BY "accounts"."id" LIMIT $3`)).
		WithArgs(user, user, 1).WillReturnRows(ShopFollowing)

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "account_shop_following" WHERE "account_shop_following"."account_id" = $1`)).
		WithArgs(user.String()).WillReturnRows(shops)

	router.POST("/login", User.LoginAccount)

	body := []byte(`{
		"email": "Test@Test.com",
		"password": "111122222"
	  }`)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	cookies := w.Result().Cookies()
	var accessToken, refreshToken *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "access_token" {
			accessToken = cookie
		}
		if cookie.Name == "refresh_token" {
			refreshToken = cookie
		}
	}

	if accessToken == nil || refreshToken == nil {
		t.Errorf("Expected access_token and refresh_token cookies to be set, but one or both are missing")
	}

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, sqlMock.ExpectationsWereMet())

}

func TestLogOutAccountSuccessNoCookie(t *testing.T) {
	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := setupMockServer.SetGinTestMode()

	MockedUtils := &mockUtils{}
	MockedUtils.On("GetTokens").Return(map[string]*models.Token{}, errors.New("no tokens"))
	User := controllers.NewUserController(MockedDataBase, MockedUtils, nil)

	router.GET("/logout", User.LogOutAccount)

	req, _ := http.NewRequest("GET", "/logout", bytes.NewBuffer([]byte{}))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

}

func TestLogOutAccountSuccessWithCookies(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	ctx, router, w := setupMockServer.SetGinTestMode()

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedDataBase, MockedUtils, nil)
	token1 := models.Token("access_token_value")
	token2 := models.Token("refresh_token_value")
	account := uuid.New()

	MockedUtils.On("GetTokens").Return(map[string]*models.Token{"access_token": &token1, "refresh_token": &token2}, nil)
	MockedUtils.On("ValidateJWT").Return(&models.CustomClaims{
		CreatedAt: 12344533,
		ExpiresAt: 12344533,
		UserUUID:  account,
	}, nil)

	MockedUtils.On("BlacklistJWT").Return(nil)

	sqlMock.ExpectBegin()
	sqlMock.ExpectExec(regexp.QuoteMeta(`UPDATE "accounts" SET "last_time_logged_out"=$1,"updated_at"=$2 WHERE id = $3 AND "accounts"."deleted_at" IS NULL`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), account).WillReturnResult(sqlmock.NewResult(1, 1))
	sqlMock.ExpectCommit()

	router.GET("/logout", User.LogOutAccount)

	req, _ := http.NewRequest("GET", "/logout", bytes.NewBuffer([]byte{}))
	req.AddCookie(&http.Cookie{Name: "access_token", Value: "access_token_value"})
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "refresh_token_value"})
	ctx.Request = req

	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	cookies := w.Result().Cookies()
	var accessToken, refreshToken *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "access_token" {
			accessToken = cookie
		}
		if cookie.Name == "refresh_token" {
			refreshToken = cookie
		}
	}

	if accessToken.Value != "" || refreshToken.Value != "" {
		t.Errorf("Expected access_token and refresh_token cookies to be set, but one or both are missing")
	}

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
	MockedUtils.AssertCalled(t, "BlacklistJWT")
}

func TestLogOutAccountUserNotFound(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	ctx, router, w := setupMockServer.SetGinTestMode()

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedDataBase, MockedUtils, nil)

	account := uuid.New()
	token1 := models.Token("access_token_value")
	token2 := models.Token("refresh_token_value")

	MockedUtils.On("GetTokens").Return(map[string]*models.Token{"access_token": &token1, "refresh_token": &token2}, nil)
	MockedUtils.On("ValidateJWT").Return(&models.CustomClaims{
		CreatedAt: 12344533,
		ExpiresAt: 12344533,
		UserUUID:  account,
	}, nil)

	sqlMock.ExpectBegin()
	sqlMock.ExpectExec(regexp.QuoteMeta(`UPDATE "accounts" SET "last_time_logged_out"=$1,"updated_at"=$2 WHERE id = $3 AND "accounts"."deleted_at" IS NULL`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), account).WillReturnError(errors.New("User Not Found"))
	sqlMock.ExpectRollback()

	router.GET("/logout", User.LogOutAccount)

	req, _ := http.NewRequest("GET", "/logout", bytes.NewBuffer([]byte{}))
	req.AddCookie(&http.Cookie{Name: "access_token", Value: "access_token_value"})
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "refresh_token_value"})
	ctx.Request = req

	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestLogOutAccount(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	ctx, router, w := setupMockServer.SetGinTestMode()

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedDataBase, MockedUtils, nil)

	account := uuid.New()
	token1 := models.Token("access_token_value")
	token2 := models.Token("refresh_token_value")

	MockedUtils.On("GetTokens").Return(map[string]*models.Token{"access_token": &token1, "refresh_token": &token2}, nil)
	MockedUtils.On("ValidateJWT").Return(&models.CustomClaims{
		CreatedAt: 12344533,
		ExpiresAt: 12344533,
		UserUUID:  account,
	}, nil)
	MockedUtils.On("BlacklistJWT").Return(nil)
	sqlMock.ExpectBegin()
	sqlMock.ExpectExec(regexp.QuoteMeta(`UPDATE "accounts" SET "last_time_logged_out"=$1,"updated_at"=$2 WHERE id = $3 AND "accounts"."deleted_at" IS NULL`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), account).WillReturnResult(sqlmock.NewResult(1, 1))
	sqlMock.ExpectCommit()

	router.GET("/logout", User.LogOutAccount)

	req, _ := http.NewRequest("GET", "/logout", bytes.NewBuffer([]byte{}))
	req.AddCookie(&http.Cookie{Name: "access_token", Value: "access_token_value"})
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "refresh_token_value"})
	ctx.Request = req

	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestVerifyAccountNoTranID(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := setupMockServer.SetGinTestMode()

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedDataBase, MockedUtils, nil)

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE email_verification_token = $1 AND "accounts"."deleted_at" IS NULL`)).WillReturnError(errors.New("TransID is not found"))

	router.GET("/verifyaccount", User.VerifyAccount)

	req, _ := http.NewRequest("GET", "/verifyaccount", bytes.NewBuffer([]byte{}))

	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.NoError(t, sqlMock.ExpectationsWereMet())

}
func TestVerifyAccountEmptyAccount(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := setupMockServer.SetGinTestMode()

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedDataBase, MockedUtils, nil)

	emptyAccount := &models.Account{}

	userID := uuid.Nil
	Account := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "first_name", "last_name", "email", "password_hashed", "subscription_type", "email_verified", "email_verification_token", "request_change_pass", "account_pass_reset_token", "last_time_logged_in", "last_time_logged_out"}).
		AddRow(userID.String(), emptyAccount.CreatedAt, emptyAccount.UpdatedAt, emptyAccount.FirstName, emptyAccount.LastName, emptyAccount.Email, emptyAccount.PasswordHashed, emptyAccount.SubscriptionType, emptyAccount.EmailVerified, emptyAccount.EmailVerificationToken, emptyAccount.RequestChangePass, emptyAccount.AccountPassResetToken, emptyAccount.LastTimeLoggedIn, emptyAccount.LastTimeLoggedOut)

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE email_verification_token = $1 AND "accounts"."deleted_at" IS NULL`)).WillReturnRows(Account)

	router.GET("/verifyaccount", User.VerifyAccount)

	req, _ := http.NewRequest("GET", "/verifyaccount", bytes.NewBuffer([]byte{}))

	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.NoError(t, sqlMock.ExpectationsWereMet())

}

func TestVerifyAccountAlreadyVerified(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := setupMockServer.SetGinTestMode()

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedDataBase, MockedUtils, nil)

	emptyAccount := &models.Account{}

	userID := uuid.Nil
	Account := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "first_name", "last_name", "email", "password_hashed", "subscription_type", "email_verified", "email_verification_token", "request_change_pass", "account_pass_reset_token", "last_time_logged_in", "last_time_logged_out"}).
		AddRow(userID.String(), emptyAccount.CreatedAt, emptyAccount.UpdatedAt, emptyAccount.FirstName, emptyAccount.LastName, emptyAccount.Email, emptyAccount.PasswordHashed, emptyAccount.SubscriptionType, true, emptyAccount.EmailVerificationToken, emptyAccount.RequestChangePass, emptyAccount.AccountPassResetToken, emptyAccount.LastTimeLoggedIn, emptyAccount.LastTimeLoggedOut)

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE email_verification_token = $1 AND "accounts"."deleted_at" IS NULL`)).WillReturnRows(Account)

	router.GET("/verifyaccount", User.VerifyAccount)

	req, _ := http.NewRequest("GET", "/verifyaccount", bytes.NewBuffer([]byte{}))

	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.NoError(t, sqlMock.ExpectationsWereMet())

}
func TestVerifyAccountSuccess(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := setupMockServer.SetGinTestMode()

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedDataBase, MockedUtils, nil)

	emptyAccount := &models.Account{}

	userID := uuid.New()
	time := time.Now()
	Account := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "first_name", "last_name", "email", "password_hashed", "subscription_type", "email_verified", "email_verification_token", "request_change_pass", "account_pass_reset_token", "last_time_logged_in", "last_time_logged_out"}).
		AddRow(userID.String(), time, time, nil, "test", "Example", emptyAccount.PasswordHashed, emptyAccount.SubscriptionType, false, "asdfsdgsdgdsfsafads", emptyAccount.RequestChangePass, emptyAccount.AccountPassResetToken, emptyAccount.LastTimeLoggedIn, emptyAccount.LastTimeLoggedOut)

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE email_verification_token = $1 AND "accounts"."deleted_at" IS NULL`)).WillReturnRows(Account)

	sqlMock.ExpectBegin()
	sqlMock.ExpectExec(regexp.QuoteMeta(`UPDATE "accounts" SET "email_verification_token"=$1,"email_verified"=$2,"updated_at"=$3 WHERE "accounts"."deleted_at" IS NULL AND "id" = $4`)).WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 3))
	sqlMock.ExpectCommit()

	router.GET("/verifyaccount", User.VerifyAccount)

	req, _ := http.NewRequest("GET", "/verifyaccount", bytes.NewBuffer([]byte{}))

	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, sqlMock.ExpectationsWereMet())

}

func TestChangePassFailedBindJson(t *testing.T) {

	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	c, router, w := setupMockServer.SetGinTestMode()

	currentUserUUID := uuid.New()
	UserRepo := &MockedUserRepository{}

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedDataBase, MockedUtils, UserRepo)
	UserRepo.On("GetAccountByID").Return(&models.Account{}, nil)

	router.POST("/changepassword", func(ctx *gin.Context) {
		ctx.Set("currentUserUUID", currentUserUUID)

	}, User.ChangePass)

	c.Request, _ = http.NewRequest("POST", "/changepassword", bytes.NewBuffer([]byte{}))

	router.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestChangePassUserNotFound(t *testing.T) {

	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	c, router, w := setupMockServer.SetGinTestMode()

	currentUserUUID := uuid.New()
	UserRepo := &MockedUserRepository{}
	MockedUtils := &mockUtils{}

	User := controllers.NewUserController(MockedDataBase, MockedUtils, UserRepo)

	UserRepo.On("GetAccountByID").Return(nil, errors.New("record not found"))
	router.POST("/changepassword", func(ctx *gin.Context) {
		ctx.Set("currentUserUUID", currentUserUUID)

	}, User.ChangePass)

	c.Request, _ = http.NewRequest("POST", "/changepassword", bytes.NewBuffer([]byte(`{
		"current_password":"qqqq1111",
		"new_password":"1111qqqq",
		"confirm_password":"1111qqqq"
	}`)))

	router.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
func TestChangePassEmptyAccount(t *testing.T) {

	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	c, router, w := setupMockServer.SetGinTestMode()

	currentUserUUID := uuid.New()

	MockedUtils := &mockUtils{}
	UserRepo := &MockedUserRepository{}

	UserRepo.On("GetAccountByID").Return(nil, errors.New("record not found"))

	User := controllers.NewUserController(MockedDataBase, MockedUtils, UserRepo)

	router.POST("/changepassword", func(ctx *gin.Context) {
		ctx.Set("currentUserUUID", currentUserUUID)

	}, User.ChangePass)

	c.Request, _ = http.NewRequest("POST", "/changepassword", bytes.NewBuffer([]byte(`{
		"current_password":"qqqq1111",
		"new_password":"1111qqqq",
		"confirm_password":"1111qqqq"
	}`)))

	router.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusNotFound, w.Code)

}
func TestChangePassPassNoMatch(t *testing.T) {

	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	c, router, w := setupMockServer.SetGinTestMode()

	currentUserUUID := uuid.New()

	MockedUtils := &mockUtils{}
	UserRepo := &MockedUserRepository{}

	UserRepo.On("GetAccountByID").Return(&models.Account{}, nil)
	User := controllers.NewUserController(MockedDataBase, MockedUtils, UserRepo)

	MockedUtils.On("IsPassVerified").Return(false)

	router.POST("/changepassword", func(ctx *gin.Context) {
		ctx.Set("currentUserUUID", currentUserUUID)
	}, User.ChangePass)

	c.Request, _ = http.NewRequest("POST", "/changepassword", bytes.NewBuffer([]byte(`{
		"current_password":"qqqq1111",
		"new_password":"1111qqq1",
		"confirm_password":"1111qqqq"
	}`)))

	router.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusNotFound, w.Code)

}
func TestChangePassHashFail(t *testing.T) {

	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	c, router, w := setupMockServer.SetGinTestMode()

	currentUserUUID := uuid.New()

	MockedUtils := &mockUtils{}
	UserRepo := &MockedUserRepository{}

	User := controllers.NewUserController(MockedDataBase, MockedUtils, UserRepo)

	UserRepo.On("GetAccountByID").Return(&models.Account{ID: currentUserUUID}, nil)
	MockedUtils.On("IsPassVerified").Return(true)
	MockedUtils.On("HashPass").Return("", errors.New("Error while hashing pass"))

	router.POST("/changepassword", func(ctx *gin.Context) {
		ctx.Set("currentUserUUID", currentUserUUID)

	}, User.ChangePass)

	c.Request, _ = http.NewRequest("POST", "/changepassword", bytes.NewBuffer([]byte(`{
		"current_password":"qqqq1111",
		"new_password":"1111qqqq",
		"confirm_password":"1111qqqq"
	}`)))

	router.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusConflict, w.Code)

}

func TestChangePassPassNotConfirmed(t *testing.T) {

	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	c, router, w := setupMockServer.SetGinTestMode()

	currentUserUUID := uuid.New()

	MockedUtils := &mockUtils{}
	UserRepo := &MockedUserRepository{}

	UserRepo.On("GetAccountByID").Return(&models.Account{ID: currentUserUUID}, nil)
	User := controllers.NewUserController(MockedDataBase, MockedUtils, UserRepo)

	MockedUtils.On("IsPassVerified").Return(true)

	router.POST("/changepassword", func(ctx *gin.Context) {
		ctx.Set("currentUserUUID", currentUserUUID)

	}, User.ChangePass)

	c.Request, _ = http.NewRequest("POST", "/changepassword", bytes.NewBuffer([]byte(`{
		"current_password":"qqqq1111",
		"new_password":"1111qqq1",
		"confirm_password":"1111qqqq"
	}`)))

	router.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

}
func TestChangePassSuccess(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	c, router, w := setupMockServer.SetGinTestMode()

	currentUserUUID := uuid.New()

	MockedUtils := &mockUtils{}
	UserRepo := &MockedUserRepository{}

	User := controllers.NewUserController(MockedDataBase, MockedUtils, UserRepo)

	UserRepo.On("GetAccountByID").Return(&models.Account{ID: currentUserUUID}, nil)
	MockedUtils.On("IsPassVerified").Return(true)
	MockedUtils.On("ValidateJWT").Return(&models.CustomClaims{}, nil)
	MockedUtils.On("GetTokens").Return(map[string]*models.Token{}, nil)
	MockedUtils.On("HashPass").Return("SomePass", nil)

	router.POST("/changepassword", func(ctx *gin.Context) {
		ctx.Set("currentUserUUID", currentUserUUID)

	}, User.ChangePass)

	sqlMock.ExpectBegin()
	sqlMock.ExpectExec(regexp.QuoteMeta(`UPDATE "accounts" SET "password_hashed"=$1,"updated_at"=$2 WHERE "accounts"."deleted_at" IS NULL AND "id" = $3`)).
		WithArgs("SomePass", sqlmock.AnyArg(), currentUserUUID.String()).WillReturnResult(sqlmock.NewResult(1, 2))
	sqlMock.ExpectCommit()

	c.Request, _ = http.NewRequest("POST", "/changepassword", bytes.NewBuffer([]byte(`{
		"current_password":"qqqq1111",
		"new_password":"1111qqqq",
		"confirm_password":"1111qqqq"
	}`)))

	router.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestForgotPassReqFailedBindJson(t *testing.T) {

	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	c, router, w := setupMockServer.SetGinTestMode()

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedDataBase, MockedUtils, nil)

	router.POST("/forgotpassword", User.ForgotPassReq)

	c.Request, _ = http.NewRequest("POST", "/forgotpassword", bytes.NewBuffer([]byte{}))

	router.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusNotFound, w.Code)

}
func TestForgotPassReqUserNotFound(t *testing.T) {

	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	c, router, w := setupMockServer.SetGinTestMode()

	MockedUtils := &mockUtils{}
	UserRepo := &MockedUserRepository{}

	User := controllers.NewUserController(MockedDataBase, MockedUtils, UserRepo)

	emptyAccount := &models.Account{}

	UserRepo.On("GetAccountByEmail").Return(emptyAccount, nil)

	router.POST("/forgotpassword", User.ForgotPassReq)

	c.Request, _ = http.NewRequest("POST", "/forgotpassword", bytes.NewBuffer([]byte(`{"email_account":"Some@Test.com"}`)))

	router.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestForgotPassReqVerificationToken(t *testing.T) {

	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	c, router, w := setupMockServer.SetGinTestMode()

	MockedUtils := &mockUtils{}
	UserRepo := &MockedUserRepository{}

	User := controllers.NewUserController(MockedDataBase, MockedUtils, UserRepo)

	email := "Some@Test.com"

	UserRepo.On("GetAccountByEmail").Return(&models.Account{Email: email}, nil)
	MockedUtils.On("CreateVerificationString").Return("", errors.New("failed to create verification token"))

	router.POST("/forgotpassword", User.ForgotPassReq)

	c.Request, _ = http.NewRequest("POST", "/forgotpassword", bytes.NewBuffer([]byte(`{"email_account":"Some@Test.com"}`)))

	router.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestForgotPassReqSuccess(t *testing.T) {

	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	c, router, w := setupMockServer.SetGinTestMode()

	MockedUtils := &mockUtils{}
	UserRepo := &MockedUserRepository{}

	User := controllers.NewUserController(MockedDataBase, MockedUtils, UserRepo)

	email := "Some@Test.com"

	UserRepo.On("GetAccountByEmail").Return(&models.Account{Email: email}, nil)
	MockedUtils.On("CreateVerificationString").Return("SomeToken", nil)
	MockedUtils.On("SendResetPassEmail")
	UserRepo.On("SaveAccount").Return(nil)

	router.POST("/forgotpassword", User.ForgotPassReq)

	c.Request, _ = http.NewRequest("POST", "/forgotpassword", bytes.NewBuffer([]byte(`{"email_account":"Some@Test.com"}`)))

	router.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusOK, w.Code)

}
func TestResetPass(t *testing.T) {

	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := setupMockServer.SetGinTestMode()

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedDataBase, MockedUtils, nil)

	router.POST("/resetpassword", User.ResetPass)

	body := []byte{}
	req, _ := http.NewRequest("POST", "/resetpassword", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, w.Code, http.StatusNotFound)
}

func TestResetPassPassNoMachs(t *testing.T) {

	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()
	_, router, w := setupMockServer.SetGinTestMode()

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedDataBase, MockedUtils, nil)

	router.POST("/resetpassword", User.ResetPass)

	body := []byte(`{"rcp":"SomeToken","new_password":"1234qwer","confirm_password":"1234qwe"}`)
	req, _ := http.NewRequest("POST", "/resetpassword", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, w.Code, http.StatusForbidden)
}

func TestResetPassUserNotFound(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := setupMockServer.SetGinTestMode()

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedDataBase, MockedUtils, nil)

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE account_pass_reset_token = $1 AND "accounts"."deleted_at" IS NULL`)).
		WithArgs("SomeToken").WillReturnError(errors.New("user not found"))

	router.POST("/resetpassword", User.ResetPass)

	body := []byte(`{"rcp":"SomeToken","new_password":"1234qwer","confirm_password":"1234qwer"}`)
	req, _ := http.NewRequest("POST", "/resetpassword", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, w.Code, http.StatusForbidden)
}
func TestResetPassEmptyAccount(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := setupMockServer.SetGinTestMode()

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedDataBase, MockedUtils, nil)

	emptyAccount := models.Account{}

	Account := sqlmock.NewRows([]string{"created_at", "updated_at", "first_name", "last_name", "email", "password_hashed", "subscription_type", "email_verified", "email_verification_token", "request_change_pass", "account_pass_reset_token", "last_time_logged_in", "last_time_logged_out"}).
		AddRow(emptyAccount.CreatedAt, emptyAccount.UpdatedAt, emptyAccount.FirstName, emptyAccount.LastName, emptyAccount.Email, emptyAccount.PasswordHashed, emptyAccount.SubscriptionType, emptyAccount.EmailVerified, emptyAccount.EmailVerificationToken, emptyAccount.RequestChangePass, emptyAccount.AccountPassResetToken, emptyAccount.LastTimeLoggedIn, emptyAccount.LastTimeLoggedOut)

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE account_pass_reset_token = $1 AND "accounts"."deleted_at" IS NULL`)).
		WithArgs("SomeToken").WillReturnRows(Account)

	router.POST("/resetpassword", User.ResetPass)

	body := []byte(`{"rcp":"SomeToken","new_password":"1234qwer","confirm_password":"1234qwer"}`)
	req, _ := http.NewRequest("POST", "/resetpassword", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, w.Code, http.StatusForbidden)
}
func TestResetPassRCPTokenNoMatch(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := setupMockServer.SetGinTestMode()

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedDataBase, MockedUtils, nil)

	emptyAccount := models.Account{}
	user := uuid.New()

	Account := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "first_name", "last_name", "email", "password_hashed", "subscription_type", "email_verified", "email_verification_token", "request_change_pass", "account_pass_reset_token", "last_time_logged_in", "last_time_logged_out"}).
		AddRow(user.String(), emptyAccount.CreatedAt, emptyAccount.UpdatedAt, emptyAccount.FirstName, emptyAccount.LastName, emptyAccount.Email, emptyAccount.PasswordHashed, emptyAccount.SubscriptionType, emptyAccount.EmailVerified, emptyAccount.EmailVerificationToken, emptyAccount.RequestChangePass, "", emptyAccount.LastTimeLoggedIn, emptyAccount.LastTimeLoggedOut)

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE account_pass_reset_token = $1 AND "accounts"."deleted_at" IS NULL`)).
		WithArgs("").WillReturnRows(Account)

	router.POST("/resetpassword", User.ResetPass)

	body := []byte(`{"rcp":"","new_password":"1234qwer","confirm_password":"1234qwer"}`)
	req, _ := http.NewRequest("POST", "/resetpassword", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, w.Code, http.StatusForbidden)
}
func TestResetPassSuccess(t *testing.T) { //delete this

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := setupMockServer.SetGinTestMode()

	MockedUtils := &mockUtils{}
	MockedUtils.On("HashPass").Return("NewHasshedPass", nil)
	User := controllers.NewUserController(MockedDataBase, MockedUtils, nil)

	emptyAccount := models.Account{}
	user := uuid.New()

	Account := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "first_name", "last_name", "email", "password_hashed", "subscription_type", "email_verified", "email_verification_token", "request_change_pass", "account_pass_reset_token", "last_time_logged_in", "last_time_logged_out"}).
		AddRow(user.String(), emptyAccount.CreatedAt, emptyAccount.UpdatedAt, emptyAccount.FirstName, emptyAccount.LastName, emptyAccount.Email, emptyAccount.PasswordHashed, emptyAccount.SubscriptionType, emptyAccount.EmailVerified, emptyAccount.EmailVerificationToken, emptyAccount.RequestChangePass, "SomeToken", emptyAccount.LastTimeLoggedIn, emptyAccount.LastTimeLoggedOut)

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE account_pass_reset_token = $1 AND "accounts"."deleted_at" IS NULL`)).
		WithArgs("SomeToken").WillReturnRows(Account)

	sqlMock.ExpectBegin()
	sqlMock.ExpectExec(regexp.QuoteMeta(`UPDATE "accounts" SET "account_pass_reset_token"=$1,"password_hashed"=$2,"request_change_pass"=$3,"updated_at"=$4 WHERE "accounts"."deleted_at" IS NULL AND "id" = $5`)).
		WithArgs("", "NewHasshedPass", false, sqlmock.AnyArg(), user.String()).WillReturnResult(sqlmock.NewResult(1, 5))
	sqlMock.ExpectCommit()

	router.POST("/resetpassword", User.ResetPass)

	body := []byte(`{"rcp":"SomeToken","new_password":"1234qwer","confirm_password":"1234qwer"}`)
	req, _ := http.NewRequest("POST", "/resetpassword", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
func TestUpdateLastTimeLoggedIn(t *testing.T) { //delete this
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	MockedUtils := &mockUtils{}

	User := controllers.NewUserController(MockedDataBase, MockedUtils, nil)

	Account := models.Account{}
	Account.ID = uuid.New()

	sqlMock.ExpectBegin()
	sqlMock.ExpectExec(regexp.QuoteMeta(`UPDATE "accounts" SET "last_time_logged_in"=$1,"updated_at"=$2 WHERE id = $3 AND "accounts"."deleted_at" IS NULL AND "id" = $4`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), Account.ID, Account.ID).WillReturnResult(sqlmock.NewResult(1, 2))
	sqlMock.ExpectCommit()

	User.UpdateLastTimeLoggedIn(&Account)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestUpdateLastTimeLoggedInFailed(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	MockedUtils := &mockUtils{}

	User := controllers.NewUserController(MockedDataBase, MockedUtils, nil)

	Account := models.Account{}

	sqlMock.ExpectBegin()
	sqlMock.ExpectExec(regexp.QuoteMeta(`UPDATE "accounts" SET "last_time_logged_in"=$1,"updated_at"=$2 WHERE id = $3 AND "accounts"."deleted_at" IS NULL`)).WillReturnError(errors.New(("user not found")))
	sqlMock.ExpectRollback()

	err := User.UpdateLastTimeLoggedIn(&Account)
	assert.Contains(t, err.Error(), "user not found")
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestJoinShopFollowing(t *testing.T) { //delete this
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	MockedUtils := &mockUtils{}

	User := controllers.NewUserController(MockedDataBase, MockedUtils, nil)

	Account := models.Account{}
	Account.ID = uuid.New()
	AccountIdtoString := Account.ID.String()

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE "accounts"."id" = $1 AND "accounts"."deleted_at" IS NULL AND "accounts"."id" = $2 ORDER BY "accounts"."id" LIMIT $3`)).
		WithArgs(AccountIdtoString, AccountIdtoString, 1).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(AccountIdtoString))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "account_shop_following" WHERE "account_shop_following"."account_id" = $1`)).
		WithArgs(AccountIdtoString).WillReturnRows(sqlmock.NewRows([]string{"account_id", "shop_id"}).AddRow(AccountIdtoString, 1))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "shops" WHERE "shops"."id" = $1 AND "shops"."deleted_at" IS NULL`)).
		WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "shops" WHERE "shops"."deleted_at" IS NULL AND "shops"."id" = $1 ORDER BY "shops"."id" LIMIT $2`)).
		WithArgs(1, 1).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "shop_members" WHERE "shop_members"."shop_id" = $1 AND "shop_members"."deleted_at" IS NULL`)).
		WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"shop_members_id", "shop_id"}).AddRow(1, 1))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "reviews" WHERE "reviews"."shop_id" = $1 AND "reviews"."deleted_at" IS NULL`)).
		WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"shop_id"}).AddRow(1))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "shop_menus" WHERE "shop_menus"."shop_id" = $1 AND "shop_menus"."deleted_at" IS NULL`)).
		WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"shop_id"}).AddRow(1))

	err := User.JoinShopFollowing(&Account)

	assert.NoError(t, err)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}
func TestJoinShopFollowingFAIL(t *testing.T) { //delete this
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	MockedUtils := &mockUtils{}

	User := controllers.NewUserController(MockedDataBase, MockedUtils, nil)

	Account := models.Account{}

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE "accounts"."id" = $1 AND "accounts"."deleted_at" IS NULL ORDER BY "accounts"."id" LIMIT $2`)).
		WithArgs(Account.ID, 1).WillReturnError(errors.New("No User Found"))

	err := User.JoinShopFollowing(&Account)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "No User Found")
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestGenerateLoginResponse(t *testing.T) {

	user := controllers.User{}
	Account := models.Account{FirstName: "John", Email: "test@Test.com", ShopsFollowing: []models.Shop{{Name: "ExampleShopName"}, {Name: "ExampleShop2"}}}
	AccessToken := models.Token("Example Token")
	RefreshToken := models.Token("Example Token")

	loginResponse := user.GenerateLoginResponse(&Account, &AccessToken, &RefreshToken)

	assert.Equal(t, &AccessToken, loginResponse.AccessToken)
	assert.Equal(t, &RefreshToken, loginResponse.RefreshToken)
	assert.Equal(t, len(Account.ShopsFollowing), len(loginResponse.User.Shops))

	for i := 0; i < len(Account.ShopsFollowing); i++ {
		assert.Equal(t, Account.ShopsFollowing[i].Name, loginResponse.User.Shops[i].Name, "Shop name are match")
	}

}

func TestUpdateLastTimeLoggedOutSuccess(t *testing.T) { //delete this
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	MockedUtils := &mockUtils{}

	User := controllers.NewUserController(MockedDataBase, MockedUtils, nil)

	Account := models.Account{}
	Account.ID = uuid.New()

	sqlMock.ExpectBegin()
	sqlMock.ExpectExec(regexp.QuoteMeta(`UPDATE "accounts" SET "last_time_logged_out"=$1,"updated_at"=$2 WHERE id = $3 AND "accounts"."deleted_at" IS NULL`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), Account.ID).WillReturnResult(sqlmock.NewResult(1, 2))
	sqlMock.ExpectCommit()

	User.UpdateLastTimeLoggedOut(Account.ID)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestUpdateLastTimeLoggedOutFailed(t *testing.T) { //delete this
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	MockedUtils := &mockUtils{}

	User := controllers.NewUserController(MockedDataBase, MockedUtils, nil)

	Account := models.Account{}

	sqlMock.ExpectBegin()
	sqlMock.ExpectExec(regexp.QuoteMeta(`UPDATE "accounts" SET "last_time_logged_out"=$1,"updated_at"=$2 WHERE id = $3 AND "accounts"."deleted_at" IS NULL`)).WillReturnError(errors.New(("user not found")))
	sqlMock.ExpectRollback()

	err := User.UpdateLastTimeLoggedOut(Account.ID)

	assert.Contains(t, err.Error(), "user not found")
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestUpdateAccountAfterVerify(t *testing.T) { //delete this
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	MockedUtils := &mockUtils{}

	User := controllers.NewUserController(MockedDataBase, MockedUtils, nil)

	Account := &models.Account{}
	Account.ID = uuid.New()

	sqlMock.ExpectBegin()
	sqlMock.ExpectExec(regexp.QuoteMeta(`UPDATE "accounts" SET "email_verification_token"=$1,"email_verified"=$2,"updated_at"=$3 WHERE "accounts"."deleted_at" IS NULL AND "id" = $4`)).
		WithArgs("", true, sqlmock.AnyArg(), Account.ID).WillReturnResult(sqlmock.NewResult(1, 3))
	sqlMock.ExpectCommit()

	err := User.UpdateAccountAfterVerify(Account)
	assert.NoError(t, err)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestUpdateAccountAfterVerifyFail(t *testing.T) { //delete this
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	MockedUtils := &mockUtils{}

	User := controllers.NewUserController(MockedDataBase, MockedUtils, nil)

	Account := &models.Account{}
	Account.ID = uuid.New()

	sqlMock.ExpectBegin()
	sqlMock.ExpectExec(regexp.QuoteMeta(`UPDATE "accounts" SET "email_verification_token"=$1,"email_verified"=$2,"updated_at"=$3 WHERE "accounts"."deleted_at" IS NULL AND "id" = $4`)).
		WithArgs("", true, sqlmock.AnyArg(), Account.ID).WillReturnError(errors.New("error while changing data"))
	sqlMock.ExpectRollback()

	err := User.UpdateAccountAfterVerify(Account)

	assert.Contains(t, err.Error(), "error while changing data")
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestUpdateAccountNewPass(t *testing.T) { //delete this
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	MockedUtils := &mockUtils{}

	User := controllers.NewUserController(MockedDataBase, MockedUtils, nil)

	HashedPass := "SomeHashedPass"
	Account := &models.Account{}
	Account.ID = uuid.New()

	sqlMock.ExpectBegin()
	sqlMock.ExpectExec(regexp.QuoteMeta(`UPDATE "accounts" SET "password_hashed"=$1,"updated_at"=$2 WHERE "accounts"."deleted_at" IS NULL AND "id" = $3`)).
		WithArgs(HashedPass, sqlmock.AnyArg(), Account.ID).WillReturnResult(sqlmock.NewResult(1, 2))
	sqlMock.ExpectCommit()

	err := User.UpdateAccountNewPass(Account, HashedPass)

	assert.NoError(t, err)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestUpdateAccountNewPassFail(t *testing.T) { //delete this
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	MockedUtils := &mockUtils{}

	User := controllers.NewUserController(MockedDataBase, MockedUtils, nil)

	HashedPass := "SomeHashedPass"
	Account := &models.Account{}
	Account.ID = uuid.New()

	sqlMock.ExpectBegin()
	sqlMock.ExpectExec(regexp.QuoteMeta(`UPDATE "accounts" SET "password_hashed"=$1,"updated_at"=$2 WHERE "accounts"."deleted_at" IS NULL AND "id" = $3`)).
		WithArgs(HashedPass, sqlmock.AnyArg(), Account.ID).WillReturnError(errors.New("error while changing data"))
	sqlMock.ExpectRollback()

	err := User.UpdateAccountNewPass(Account, HashedPass)

	assert.Contains(t, err.Error(), "error while changing data")
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestUpdateAccountAfterResetPass(t *testing.T) { //delete this

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	MockedUtils := &mockUtils{}

	User := controllers.NewUserController(MockedDataBase, MockedUtils, nil)

	HashedPass := "SomeHashedPass"
	Account := &models.Account{}
	Account.ID = uuid.New()

	sqlMock.ExpectBegin()
	sqlMock.ExpectExec(regexp.QuoteMeta(`UPDATE "accounts" SET "account_pass_reset_token"=$1,"password_hashed"=$2,"request_change_pass"=$3,"updated_at"=$4 WHERE "accounts"."deleted_at" IS NULL AND "id" = $5`)).
		WithArgs("", HashedPass, false, sqlmock.AnyArg(), Account.ID).WillReturnResult(sqlmock.NewResult(1, 2))
	sqlMock.ExpectCommit()

	err := User.UpdateAccountAfterResetPass(Account, HashedPass)

	assert.NoError(t, err)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestUpdateAccountAfterFail(t *testing.T) { //delete this

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	MockedUtils := &mockUtils{}

	User := controllers.NewUserController(MockedDataBase, MockedUtils, nil)

	HashedPass := "SomeHashedPass"
	Account := &models.Account{}
	Account.ID = uuid.New()

	sqlMock.ExpectBegin()
	sqlMock.ExpectExec(regexp.QuoteMeta(`UPDATE "accounts" SET "account_pass_reset_token"=$1,"password_hashed"=$2,"request_change_pass"=$3,"updated_at"=$4 WHERE "accounts"."deleted_at" IS NULL AND "id" = $5`)).
		WithArgs("", HashedPass, false, sqlmock.AnyArg(), Account.ID).WillReturnError(errors.New("error while saving record"))
	sqlMock.ExpectRollback()

	err := User.UpdateAccountAfterResetPass(Account, HashedPass)

	assert.Contains(t, err.Error(), "error while saving record")
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestCreateNewAccountRecordSuccess(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	MockedUtils := &mockUtils{}

	User := controllers.NewUserController(MockedDataBase, MockedUtils, nil)

	AccountRequestInfo := &controllers.RegisterAccount{

		FirstName:        "Testing",
		LastName:         "User",
		Email:            "test11@testing.com",
		Password:         "1111qqqq",
		PasswordConfirm:  "2222wwww",
		SubscriptionType: "free",
	}

	HashedPass := "SomeHashedPass"
	EmailVerificationToken := "SomeTokenString"

	newAccount := &models.Account{

		FirstName:              AccountRequestInfo.FirstName,
		LastName:               AccountRequestInfo.LastName,
		Email:                  AccountRequestInfo.Email,
		PasswordHashed:         HashedPass,
		SubscriptionType:       AccountRequestInfo.SubscriptionType,
		EmailVerificationToken: EmailVerificationToken,
	}

	sqlMock.ExpectBegin()
	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "accounts" ("id","created_at","updated_at","deleted_at","first_name","last_name","email","password_hashed","subscription_type","email_verified","email_verification_token","request_change_pass","account_pass_reset_token","last_time_logged_in","last_time_logged_out") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15) RETURNING "id"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), nil, AccountRequestInfo.FirstName, AccountRequestInfo.LastName, AccountRequestInfo.Email, HashedPass, newAccount.SubscriptionType, false, EmailVerificationToken, false, "", sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(newAccount.ID.String()))
	sqlMock.ExpectCommit()

	_, err := User.CreateNewAccountRecord(AccountRequestInfo, HashedPass, EmailVerificationToken)

	assert.NoError(t, err)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestCreateNewAccountRecordFail(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	MockedUtils := &mockUtils{}

	User := controllers.NewUserController(MockedDataBase, MockedUtils, nil)

	AccountRequestInfo := &controllers.RegisterAccount{

		FirstName:        "Testing",
		LastName:         "User",
		Email:            "test11@testing.com",
		Password:         "1111qqqq",
		PasswordConfirm:  "2222wwww",
		SubscriptionType: "free",
	}

	HashedPass := "SomeHashedPass"
	EmailVerificationToken := "SomeTokenString"

	TestCases := []string{"this email is already in use", "error while handling DataBase operations"}
	for _, TestCase := range TestCases {
		sqlMock.ExpectBegin()
		sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "accounts" ("id","created_at","updated_at","deleted_at","first_name","last_name","email","password_hashed","subscription_type","email_verified","email_verification_token","request_change_pass","account_pass_reset_token","last_time_logged_in","last_time_logged_out") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15) RETURNING "id"`)).WillReturnError(errors.New(TestCase))
		sqlMock.ExpectRollback()

		_, err := User.CreateNewAccountRecord(AccountRequestInfo, HashedPass, EmailVerificationToken)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), TestCase)
	}
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}
