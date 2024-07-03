package controllers_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"EtsyScraper/controllers"
	initializer "EtsyScraper/init"
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
func (mr *MockedUserRepository) GetAccountWithShops(accountID uuid.UUID) (*models.Account, error) {

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
func (mr *MockedUserRepository) InsertTokenForAccount(column, token string, VerifyUser *models.Account) (*models.Account, error) {
	args := mr.Called()
	UserRepo := args.Get(0)
	var Account *models.Account
	if UserRepo != nil {
		Account = UserRepo.(*models.Account)
	}
	return Account, args.Error(1)
}

func (mr *MockedUserRepository) CreateAccount(account *models.Account) (*models.Account, error) {

	args := mr.Called()
	UserRepo := args.Get(0)
	var Account *models.Account
	if UserRepo != nil {
		Account = UserRepo.(*models.Account)
	}
	return Account, args.Error(1)

}
func (mr *MockedUserRepository) JoinShopFollowing(Account *models.Account) (*models.Account, error) {

	args := mr.Called()
	UserRepo := args.Get(0)
	if UserRepo != nil {
		Account = UserRepo.(*models.Account)
	}
	return Account, args.Error(1)

}

var MockedConfig = initializer.Config{}

func TestRegisterUserInvalidJson(t *testing.T) {

	_, router, w := setupMockServer.SetGinTestMode()

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedUtils, nil, MockedConfig)

	router.POST("/register", User.RegisterUser)

	body := []byte(`{"invalid-json": "facek-data"}`)
	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, w.Code, http.StatusBadRequest)

}

func TestRegisterUserPassNoMatch(t *testing.T) {

	_, router, w := setupMockServer.SetGinTestMode()

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedUtils, nil, MockedConfig)

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

	_, router, w := setupMockServer.SetGinTestMode()

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedUtils, nil, MockedConfig)

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

	_, router, w := setupMockServer.SetGinTestMode()

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedUtils, nil, MockedConfig)

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

	_, router, w := setupMockServer.SetGinTestMode()

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedUtils, nil, MockedConfig)

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

	_, router, _ := setupMockServer.SetGinTestMode()

	MockedUtils := &mockUtils{}
	UserRepo := &MockedUserRepository{}

	User := controllers.NewUserController(MockedUtils, UserRepo, MockedConfig)

	router.POST("/register", User.RegisterUser)

	MockedUtils.On("HashPass").Return(" ", nil)
	MockedUtils.On("CreateVerificationString").Return(" ", nil)
	UserRepo.On("CreateAccount").Return(nil, errors.New("email is in use"))

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

}

func TestRegisterUserExpectCreateAccount(t *testing.T) {

	_, router, w := setupMockServer.SetGinTestMode()

	MockedUtils := &mockUtils{}
	UserRepo := &MockedUserRepository{}
	User := controllers.NewUserController(MockedUtils, UserRepo, MockedConfig)

	MockedUtils.On("HashPass").Return("", nil)
	MockedUtils.On("CreateVerificationString").Return("", nil)
	UserRepo.On("CreateAccount").Return(&models.Account{FirstName: "Testing", LastName: "User", Email: "test@test1242q21.com"}, nil)

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

}

func TestLoginAccountInvalidJson(t *testing.T) {

	_, router, w := setupMockServer.SetGinTestMode()

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedUtils, nil, MockedConfig)

	router.POST("/login", User.LoginAccount)

	body := []byte(`{"invalid-json": "facek-data"}`)
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, w.Code, http.StatusNotFound)

}

func TestLoginAccountInvalidEmailEmpty(t *testing.T) {

	_, router, w := setupMockServer.SetGinTestMode()

	MockedUtils := &mockUtils{}
	UserRepo := &MockedUserRepository{}
	emptyAccount := &models.Account{}

	User := controllers.NewUserController(MockedUtils, UserRepo, MockedConfig)

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

	_, router, w := setupMockServer.SetGinTestMode()

	MockedUtils := &mockUtils{}
	UserRepo := &MockedUserRepository{}

	email := "Test@Test.com"
	user := uuid.New()

	User := controllers.NewUserController(MockedUtils, UserRepo, MockedConfig)

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

	_, router, w := setupMockServer.SetGinTestMode()

	expectedError := errors.New("Error while creating Token")

	MockedUtils := &mockUtils{}
	UserRepo := &MockedUserRepository{}
	User := controllers.NewUserController(MockedUtils, UserRepo, MockedConfig)

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

	_, router, w := setupMockServer.SetGinTestMode()

	expectedError := errors.New("Error while creating Token")

	MockedUtils := &mockUtils{}
	UserRepo := &MockedUserRepository{}
	email := "Test@Test.com"
	user := uuid.New()

	User := controllers.NewUserController(MockedUtils, UserRepo, MockedConfig)

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

	_, router, w := setupMockServer.SetGinTestMode()

	expectedError := errors.New("Error while creating Token")

	MockedUtils := &mockUtils{}
	UserRepo := &MockedUserRepository{}
	email := "Test@Test.com"
	user := uuid.New()

	User := controllers.NewUserController(MockedUtils, UserRepo, MockedConfig)

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

	_, router, w := setupMockServer.SetGinTestMode()

	MockedUtils := &mockUtils{}
	UserRepo := &MockedUserRepository{}
	email := "Test@Test.com"
	user := uuid.New()

	User := controllers.NewUserController(MockedUtils, UserRepo, MockedConfig)

	UserRepo.On("GetAccountByEmail").Return(&models.Account{ID: user, Email: email}, nil)
	UserRepo.On("UpdateLastTimeLoggedIn").Return(nil)
	UserRepo.On("JoinShopFollowing").Return(&models.Account{ID: user, Email: email}, nil)
	MockedUtils.On("IsPassVerified").Return(true)
	MockedUtils.On("CreateJwtToken").Return(models.NewToken("Token"), nil)
	MockedUtils.On("CreateJwtToken").Return(models.NewToken("Token"), nil)

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

}

func TestLogOutAccountSuccessNoCookie(t *testing.T) {

	_, router, w := setupMockServer.SetGinTestMode()

	MockedUtils := &mockUtils{}
	MockedUtils.On("GetTokens").Return(map[string]*models.Token{}, errors.New("no tokens"))
	User := controllers.NewUserController(MockedUtils, nil, MockedConfig)

	router.GET("/logout", User.LogOutAccount)

	req, _ := http.NewRequest("GET", "/logout", bytes.NewBuffer([]byte{}))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

}

func TestLogOutAccountSuccessWithCookies(t *testing.T) {

	ctx, router, w := setupMockServer.SetGinTestMode()

	MockedUtils := &mockUtils{}
	UserRepo := &MockedUserRepository{}

	User := controllers.NewUserController(MockedUtils, UserRepo, MockedConfig)
	token1 := models.Token("access_token_value")
	token2 := models.Token("refresh_token_value")
	account := uuid.New()

	MockedUtils.On("GetTokens").Return(map[string]*models.Token{"access_token": &token1, "refresh_token": &token2}, nil)
	MockedUtils.On("ValidateJWT").Return(&models.CustomClaims{
		CreatedAt: 12344533,
		ExpiresAt: 12344533,
		UserUUID:  account,
	}, nil)
	UserRepo.On("UpdateLastTimeLoggedOut").Return(nil)
	MockedUtils.On("BlacklistJWT").Return(nil)

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
	MockedUtils.AssertCalled(t, "BlacklistJWT")
}

func TestLogOutAccountUserNotFound(t *testing.T) {

	ctx, router, w := setupMockServer.SetGinTestMode()

	MockedUtils := &mockUtils{}
	UserRepo := &MockedUserRepository{}

	User := controllers.NewUserController(MockedUtils, UserRepo, MockedConfig)

	account := uuid.New()
	token1 := models.Token("access_token_value")
	token2 := models.Token("refresh_token_value")

	MockedUtils.On("GetTokens").Return(map[string]*models.Token{"access_token": &token1, "refresh_token": &token2}, nil)
	MockedUtils.On("ValidateJWT").Return(&models.CustomClaims{
		CreatedAt: 12344533,
		ExpiresAt: 12344533,
		UserUUID:  account,
	}, nil)
	UserRepo.On("UpdateLastTimeLoggedOut").Return(errors.New("Account was not found"))

	router.GET("/logout", User.LogOutAccount)

	req, _ := http.NewRequest("GET", "/logout", bytes.NewBuffer([]byte{}))
	req.AddCookie(&http.Cookie{Name: "access_token", Value: "access_token_value"})
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "refresh_token_value"})
	ctx.Request = req

	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestLogOutAccount(t *testing.T) {

	ctx, router, w := setupMockServer.SetGinTestMode()

	MockedUtils := &mockUtils{}
	UserRepo := &MockedUserRepository{}

	User := controllers.NewUserController(MockedUtils, UserRepo, MockedConfig)

	account := uuid.New()
	token1 := models.Token("access_token_value")
	token2 := models.Token("refresh_token_value")

	MockedUtils.On("GetTokens").Return(map[string]*models.Token{"access_token": &token1, "refresh_token": &token2}, nil)
	MockedUtils.On("ValidateJWT").Return(&models.CustomClaims{
		CreatedAt: 12344533,
		ExpiresAt: 12344533,
		UserUUID:  account,
	}, nil)
	UserRepo.On("UpdateLastTimeLoggedOut").Return(nil)

	MockedUtils.On("BlacklistJWT").Return(nil)

	router.GET("/logout", User.LogOutAccount)

	req, _ := http.NewRequest("GET", "/logout", bytes.NewBuffer([]byte{}))
	req.AddCookie(&http.Cookie{Name: "access_token", Value: "access_token_value"})
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "refresh_token_value"})
	ctx.Request = req

	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

}

func TestVerifyAccountNoTranID(t *testing.T) {

	_, router, w := setupMockServer.SetGinTestMode()

	MockedUtils := &mockUtils{}
	UserRepo := &MockedUserRepository{}
	User := controllers.NewUserController(MockedUtils, UserRepo, MockedConfig)

	UserRepo.On("InsertTokenForAccount").Return(nil, errors.New("no TranID"))

	router.GET("/verifyaccount", User.VerifyAccount)

	req, _ := http.NewRequest("GET", "/verifyaccount", bytes.NewBuffer([]byte{}))

	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)

}
func TestVerifyAccountEmptyAccount(t *testing.T) {

	_, router, w := setupMockServer.SetGinTestMode()

	MockedUtils := &mockUtils{}
	UserRepo := &MockedUserRepository{}
	User := controllers.NewUserController(MockedUtils, UserRepo, MockedConfig)

	UserRepo.On("InsertTokenForAccount").Return(&models.Account{}, nil)
	router.GET("/verifyaccount", User.VerifyAccount)

	req, _ := http.NewRequest("GET", "/verifyaccount", bytes.NewBuffer([]byte{}))

	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)

}

func TestVerifyAccountAlreadyVerified(t *testing.T) {

	_, router, w := setupMockServer.SetGinTestMode()

	MockedUtils := &mockUtils{}
	UserRepo := &MockedUserRepository{}
	User := controllers.NewUserController(MockedUtils, UserRepo, MockedConfig)

	emptyAccount := &models.Account{EmailVerified: true}
	UserRepo.On("InsertTokenForAccount").Return(emptyAccount, nil)

	router.GET("/verifyaccount", User.VerifyAccount)

	req, _ := http.NewRequest("GET", "/verifyaccount", bytes.NewBuffer([]byte{}))

	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)

}
func TestVerifyAccountSuccess(t *testing.T) {
	sqlMock, testDB, _ := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := setupMockServer.SetGinTestMode()

	MockedUtils := &mockUtils{}
	UserRepo := &MockedUserRepository{}
	User := controllers.NewUserController(MockedUtils, UserRepo, MockedConfig)

	UserRepo.On("InsertTokenForAccount").Return(&models.Account{ID: uuid.New(), EmailVerified: false}, nil)
	UserRepo.On("UpdateAccountAfterVerify").Return(nil)

	router.GET("/verifyaccount", User.VerifyAccount)

	req, _ := http.NewRequest("GET", "/verifyaccount", bytes.NewBuffer([]byte{}))

	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, sqlMock.ExpectationsWereMet())

}

func TestChangePassFailedBindJson(t *testing.T) {

	c, router, w := setupMockServer.SetGinTestMode()

	currentUserUUID := uuid.New()
	UserRepo := &MockedUserRepository{}

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedUtils, UserRepo, MockedConfig)
	UserRepo.On("GetAccountByID").Return(&models.Account{}, nil)

	router.POST("/changepassword", func(ctx *gin.Context) {
		ctx.Set("currentUserUUID", currentUserUUID)

	}, User.ChangePass)

	c.Request, _ = http.NewRequest("POST", "/changepassword", bytes.NewBuffer([]byte{}))

	router.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestChangePassUserNotFound(t *testing.T) {

	c, router, w := setupMockServer.SetGinTestMode()

	currentUserUUID := uuid.New()
	UserRepo := &MockedUserRepository{}
	MockedUtils := &mockUtils{}

	User := controllers.NewUserController(MockedUtils, UserRepo, MockedConfig)

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

	c, router, w := setupMockServer.SetGinTestMode()

	currentUserUUID := uuid.New()

	MockedUtils := &mockUtils{}
	UserRepo := &MockedUserRepository{}

	UserRepo.On("GetAccountByID").Return(nil, errors.New("record not found"))

	User := controllers.NewUserController(MockedUtils, UserRepo, MockedConfig)

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

	c, router, w := setupMockServer.SetGinTestMode()

	currentUserUUID := uuid.New()

	MockedUtils := &mockUtils{}
	UserRepo := &MockedUserRepository{}

	UserRepo.On("GetAccountByID").Return(&models.Account{}, nil)
	User := controllers.NewUserController(MockedUtils, UserRepo, MockedConfig)

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

	c, router, w := setupMockServer.SetGinTestMode()

	currentUserUUID := uuid.New()

	MockedUtils := &mockUtils{}
	UserRepo := &MockedUserRepository{}

	User := controllers.NewUserController(MockedUtils, UserRepo, MockedConfig)

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

	c, router, w := setupMockServer.SetGinTestMode()

	currentUserUUID := uuid.New()

	MockedUtils := &mockUtils{}
	UserRepo := &MockedUserRepository{}

	UserRepo.On("GetAccountByID").Return(&models.Account{ID: currentUserUUID}, nil)
	User := controllers.NewUserController(MockedUtils, UserRepo, MockedConfig)

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

	c, router, w := setupMockServer.SetGinTestMode()

	currentUserUUID := uuid.New()

	MockedUtils := &mockUtils{}
	UserRepo := &MockedUserRepository{}

	User := controllers.NewUserController(MockedUtils, UserRepo, MockedConfig)

	UserRepo.On("GetAccountByID").Return(&models.Account{ID: currentUserUUID}, nil)
	MockedUtils.On("IsPassVerified").Return(true)
	MockedUtils.On("ValidateJWT").Return(&models.CustomClaims{}, nil)
	MockedUtils.On("GetTokens").Return(map[string]*models.Token{}, nil)
	MockedUtils.On("HashPass").Return("SomePass", nil)
	UserRepo.On("UpdateAccountNewPass").Return(nil)

	router.POST("/changepassword", func(ctx *gin.Context) {
		ctx.Set("currentUserUUID", currentUserUUID)

	}, User.ChangePass)

	c.Request, _ = http.NewRequest("POST", "/changepassword", bytes.NewBuffer([]byte(`{
		"current_password":"qqqq1111",
		"new_password":"1111qqqq",
		"confirm_password":"1111qqqq"
	}`)))

	router.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestForgotPassReqFailedBindJson(t *testing.T) {

	c, router, w := setupMockServer.SetGinTestMode()

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedUtils, nil, MockedConfig)

	router.POST("/forgotpassword", User.ForgotPassReq)

	c.Request, _ = http.NewRequest("POST", "/forgotpassword", bytes.NewBuffer([]byte{}))

	router.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusNotFound, w.Code)

}
func TestForgotPassReqUserNotFound(t *testing.T) {

	c, router, w := setupMockServer.SetGinTestMode()

	MockedUtils := &mockUtils{}
	UserRepo := &MockedUserRepository{}

	User := controllers.NewUserController(MockedUtils, UserRepo, MockedConfig)

	emptyAccount := &models.Account{}

	UserRepo.On("GetAccountByEmail").Return(emptyAccount, nil)

	router.POST("/forgotpassword", User.ForgotPassReq)

	c.Request, _ = http.NewRequest("POST", "/forgotpassword", bytes.NewBuffer([]byte(`{"email_account":"Some@Test.com"}`)))

	router.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestForgotPassReqVerificationToken(t *testing.T) {

	c, router, w := setupMockServer.SetGinTestMode()

	MockedUtils := &mockUtils{}
	UserRepo := &MockedUserRepository{}

	User := controllers.NewUserController(MockedUtils, UserRepo, MockedConfig)

	email := "Some@Test.com"

	UserRepo.On("GetAccountByEmail").Return(&models.Account{Email: email}, nil)
	MockedUtils.On("CreateVerificationString").Return("", errors.New("failed to create verification token"))

	router.POST("/forgotpassword", User.ForgotPassReq)

	c.Request, _ = http.NewRequest("POST", "/forgotpassword", bytes.NewBuffer([]byte(`{"email_account":"Some@Test.com"}`)))

	router.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestForgotPassReqSuccess(t *testing.T) {

	c, router, w := setupMockServer.SetGinTestMode()

	MockedUtils := &mockUtils{}
	UserRepo := &MockedUserRepository{}

	User := controllers.NewUserController(MockedUtils, UserRepo, MockedConfig)

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

	_, router, w := setupMockServer.SetGinTestMode()

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedUtils, nil, MockedConfig)

	router.POST("/resetpassword", User.ResetPass)

	body := []byte{}
	req, _ := http.NewRequest("POST", "/resetpassword", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, w.Code, http.StatusNotFound)
}

func TestResetPassPassNoMachs(t *testing.T) {

	_, router, w := setupMockServer.SetGinTestMode()

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedUtils, nil, MockedConfig)

	router.POST("/resetpassword", User.ResetPass)

	body := []byte(`{"rcp":"SomeToken","new_password":"1234qwer","confirm_password":"1234qwe"}`)
	req, _ := http.NewRequest("POST", "/resetpassword", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, w.Code, http.StatusForbidden)
}

func TestResetPassUserNotFound(t *testing.T) {

	_, router, w := setupMockServer.SetGinTestMode()

	MockedUtils := &mockUtils{}
	UserRepo := &MockedUserRepository{}
	User := controllers.NewUserController(MockedUtils, UserRepo, MockedConfig)

	UserRepo.On("InsertTokenForAccount").Return(nil, errors.New("user not found"))

	router.POST("/resetpassword", User.ResetPass)

	body := []byte(`{"rcp":"SomeToken","new_password":"1234qwer","confirm_password":"1234qwer"}`)
	req, _ := http.NewRequest("POST", "/resetpassword", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, w.Code, http.StatusForbidden)
}
func TestResetPassEmptyAccount(t *testing.T) {

	_, router, w := setupMockServer.SetGinTestMode()

	MockedUtils := &mockUtils{}
	UserRepo := &MockedUserRepository{}
	User := controllers.NewUserController(MockedUtils, UserRepo, MockedConfig)

	emptyAccount := &models.Account{}

	UserRepo.On("InsertTokenForAccount").Return(emptyAccount, nil)

	router.POST("/resetpassword", User.ResetPass)

	body := []byte(`{"rcp":"SomeToken","new_password":"1234qwer","confirm_password":"1234qwer"}`)
	req, _ := http.NewRequest("POST", "/resetpassword", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, w.Code, http.StatusForbidden)
}
func TestResetPassRCPTokenNoMatch(t *testing.T) {

	_, router, w := setupMockServer.SetGinTestMode()

	MockedUtils := &mockUtils{}
	UserRepo := &MockedUserRepository{}
	User := controllers.NewUserController(MockedUtils, UserRepo, MockedConfig)

	emptyAccount := &models.Account{AccountPassResetToken: "NoMAtch"}
	UserRepo.On("InsertTokenForAccount").Return(emptyAccount, nil)

	router.POST("/resetpassword", User.ResetPass)

	body := []byte(`{"rcp":"SomeToken","new_password":"1234qwer","confirm_password":"1234qwer"}`)
	req, _ := http.NewRequest("POST", "/resetpassword", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, w.Code, http.StatusForbidden)
}
func TestResetPassSuccess(t *testing.T) { //delete this

	_, router, w := setupMockServer.SetGinTestMode()

	MockedUtils := &mockUtils{}
	UserRepo := &MockedUserRepository{}
	User := controllers.NewUserController(MockedUtils, UserRepo, MockedConfig)

	emptyAccount := &models.Account{ID: uuid.New(), AccountPassResetToken: "SomeToken"}

	MockedUtils.On("HashPass").Return("NewHasshedPass", nil)
	UserRepo.On("InsertTokenForAccount").Return(emptyAccount, nil)
	UserRepo.On("UpdateAccountAfterResetPass").Return(nil)

	router.POST("/resetpassword", User.ResetPass)

	body := []byte(`{"rcp":"SomeToken","new_password":"1234qwer","confirm_password":"1234qwer"}`)
	req, _ := http.NewRequest("POST", "/resetpassword", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
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

func TestCreateNewAccountRecordSuccess(t *testing.T) {

	MockedUtils := &mockUtils{}
	UserRepo := &MockedUserRepository{}
	User := controllers.NewUserController(MockedUtils, UserRepo, MockedConfig)

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
	UserRepo.On("CreateAccount").Return(newAccount, nil)

	_, err := User.CreateNewAccountRecord(AccountRequestInfo, HashedPass, EmailVerificationToken)

	assert.NoError(t, err)

}

func TestCreateNewAccountRecordFail(t *testing.T) {

	MockedUtils := &mockUtils{}

	UserRepo := &MockedUserRepository{}
	User := controllers.NewUserController(MockedUtils, UserRepo, MockedConfig)

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
	UserRepo.On("CreateAccount").Return(nil, errors.New("error"))
	_, err := User.CreateNewAccountRecord(AccountRequestInfo, HashedPass, EmailVerificationToken)

	assert.Error(t, err)

}
