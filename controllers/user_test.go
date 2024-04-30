package controllers_test

import (
	"EtsyScraper/controllers"
	"EtsyScraper/models"
	setupMockServer "EtsyScraper/setupTests"
	"EtsyScraper/utils"
	"bytes"
	"errors"
	"log"
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

func SetGinTestMode() (*gin.Context, *gin.Engine, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	ctx, router := gin.CreateTestContext(w)
	return ctx, router, w
}

func TestRegisterUser_InvalidJson(t *testing.T) {
	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := SetGinTestMode()

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedDataBase, MockedUtils)

	router.POST("/register", User.RegisterUser)

	body := []byte(`{"invalid-json": "facek-data"}`)
	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, w.Code, http.StatusBadRequest)

}

func TestRegisterUser_PassNoMatch(t *testing.T) {
	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := SetGinTestMode()

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedDataBase, MockedUtils)

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

func TestRegisterUser_PassShort(t *testing.T) {
	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := SetGinTestMode()

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedDataBase, MockedUtils)

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

func TestRegisterUser_HashPassErr(t *testing.T) {
	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := SetGinTestMode()

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedDataBase, MockedUtils)

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

func TestRegisterUser_VerificationStringErr(t *testing.T) {
	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := SetGinTestMode()

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedDataBase, MockedUtils)

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

func TestRegisterUser_DataBaseError(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, _ := SetGinTestMode()

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedDataBase, MockedUtils)

	ErrorCases := []string{"email is in use", "some other error"}

	router.POST("/register", User.RegisterUser)
	for _, Error := range ErrorCases {

		sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "accounts" ("id","created_at","updated_at","deleted_at","first_name","last_name","email","password_hashed","subscription_type","email_verified","email_verification_token","request_change_pass","account_pass_reset_token","last_time_logged_in","last_time_logged_out") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15) RETURNING "id"`)).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), "Testing", "User", "test@test1242q21.com", "", "free", false, "", false, "", sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnError(errors.New(Error))

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
		assert.Error(t, sqlMock.ExpectationsWereMet())
	}
}

func TestRegisterUser_ExpectCreateAccount(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := SetGinTestMode()

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedDataBase, MockedUtils)

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

func TestGetAccountByEmail_Success(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedDataBase, MockedUtils)

	email := "test@test.com"
	user := uuid.New().String()
	time := time.Now()

	Account := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "first_name", "last_name", "email", "password_hashed", "subscription_type", "email_verified", "email_verification_token", "request_change_pass", "account_pass_reset_token", "last_time_logged_in", "last_time_logged_out"}).
		AddRow(user, time, time, time, "Testing", "User", "test@test1242q21.com", "", "free", false, "", false, "", time, time)

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE email = $1 AND "accounts"."deleted_at" IS NULL ORDER BY "accounts"."id" LIMIT $2`)).
		WithArgs(email, 1).WillReturnRows(Account)

	User.GetAccountByEmail(email)

	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestGetAccountByID_Success(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedDataBase, MockedUtils)

	email := "test@test.com"
	user := uuid.New()
	time := time.Now()

	Account := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "first_name", "last_name", "email", "password_hashed", "subscription_type", "email_verified", "email_verification_token", "request_change_pass", "account_pass_reset_token", "last_time_logged_in", "last_time_logged_out"}).
		AddRow(user.String(), time, time, time, "Testing", "User", email, "", "free", false, "", false, "", time, time)

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE ID = $1 AND "accounts"."deleted_at" IS NULL ORDER BY "accounts"."id" LIMIT $2`)).
		WithArgs(user, 1).WillReturnRows(Account)

	User.GetAccountByID(user)

	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestGetAccountByID_Fail(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedDataBase, MockedUtils)

	expectedError := errors.New("no Account Found")
	user := uuid.New()

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE ID = $1 AND "accounts"."deleted_at" IS NULL ORDER BY "accounts"."id" LIMIT $2`)).
		WithArgs(user, 1).WillReturnError(expectedError)

	User.GetAccountByID(user)

	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestLoginAccount_InvalidJson(t *testing.T) {
	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := SetGinTestMode()

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedDataBase, MockedUtils)

	router.POST("/login", User.LoginAccount)

	body := []byte(`{"invalid-json": "facek-data"}`)
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, w.Code, http.StatusNotFound)

}

func TestLoginAccount_InvalidEmailEmpty(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := SetGinTestMode()

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedDataBase, MockedUtils)

	emptyAccount := &models.Account{}

	userID := uuid.Nil
	Account := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "first_name", "last_name", "email", "password_hashed", "subscription_type", "email_verified", "email_verification_token", "request_change_pass", "account_pass_reset_token", "last_time_logged_in", "last_time_logged_out"}).
		AddRow(userID.String(), emptyAccount.CreatedAt, emptyAccount.UpdatedAt, emptyAccount.FirstName, emptyAccount.LastName, emptyAccount.Email, emptyAccount.PasswordHashed, emptyAccount.SubscriptionType, emptyAccount.EmailVerified, emptyAccount.EmailVerificationToken, emptyAccount.RequestChangePass, emptyAccount.AccountPassResetToken, emptyAccount.LastTimeLoggedIn, emptyAccount.LastTimeLoggedOut)

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE email = $1 AND "accounts"."deleted_at" IS NULL ORDER BY "accounts"."id" LIMIT $2`)).WillReturnRows(Account)

	router.POST("/login", User.LoginAccount)

	body := []byte(`{
		"email": "Test@Test.com",
		"password": "1234qwer"
	  }`)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)
	log.Println(w.Body)
	assert.Equal(t, w.Code, http.StatusNotFound)
	assert.NoError(t, sqlMock.ExpectationsWereMet())

}

func TestLoginAccount_PassVerified_fail(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := SetGinTestMode()

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedDataBase, MockedUtils)

	MockedUtils.On("IsPassVerified").Return(false)

	email := "Test@Test.com"
	user := uuid.New().String()
	time := time.Now()

	Account := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "first_name", "last_name", "email", "password_hashed", "subscription_type", "email_verified", "email_verification_token", "request_change_pass", "account_pass_reset_token", "last_time_logged_in", "last_time_logged_out"}).
		AddRow(user, time, time, time, "Testing", "User", "test@test1242q21.com", "111122222", "free", false, "", false, "", time, time)

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE email = $1 AND "accounts"."deleted_at" IS NULL ORDER BY "accounts"."id" LIMIT $2`)).
		WithArgs(email, 1).WillReturnRows(Account)
	router.POST("/login", User.LoginAccount)

	body := []byte(`{
		"email": "Test@Test.com",
		"password": "111122222"
	  }`)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, w.Code, http.StatusNotFound)
	assert.NoError(t, sqlMock.ExpectationsWereMet())

}

func TestLoginAccount_AccessToken(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := SetGinTestMode()

	expectedError := errors.New("Error while creating Token")

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedDataBase, MockedUtils)

	MockedUtils.On("IsPassVerified").Return(true)
	MockedUtils.On("CreateJwtToken").Return(models.NewToken(""), expectedError)

	email := "Test@Test.com"
	user := uuid.New().String()
	time := time.Now()

	Account := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "first_name", "last_name", "email", "password_hashed", "subscription_type", "email_verified", "email_verification_token", "request_change_pass", "account_pass_reset_token", "last_time_logged_in", "last_time_logged_out"}).
		AddRow(user, time, time, time, "Testing", "User", email, "111122222", "free", false, "", false, "", time, time)

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE email = $1 AND "accounts"."deleted_at" IS NULL ORDER BY "accounts"."id" LIMIT $2`)).
		WithArgs(email, 1).WillReturnRows(Account)

	router.POST("/login", User.LoginAccount)

	body := []byte(`{
		"email": "Test@Test.com",
		"password": "111122222"
	  }`)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.NoError(t, sqlMock.ExpectationsWereMet())

}

func TestLoginAccount_AccessToken_failed(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := SetGinTestMode()

	expectedError := errors.New("Error while creating Token")

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedDataBase, MockedUtils)

	email := "Test@Test.com"
	user := uuid.New().String()
	time := time.Now()

	MockedUtils.On("IsPassVerified").Return(true)
	MockedUtils.On("CreateJwtToken").Return(models.NewToken(""), expectedError)

	Account := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "first_name", "last_name", "email", "password_hashed", "subscription_type", "email_verified", "email_verification_token", "request_change_pass", "account_pass_reset_token", "last_time_logged_in", "last_time_logged_out"}).
		AddRow(user, time, time, time, "Testing", "User", email, "111122222", "free", false, "", false, "", time, time)

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE email = $1 AND "accounts"."deleted_at" IS NULL ORDER BY "accounts"."id" LIMIT $2`)).
		WithArgs(email, 1).WillReturnRows(Account)

	router.POST("/login", User.LoginAccount)

	body := []byte(`{
		"email": "Test@Test.com",
		"password": "111122222"
	  }`)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.NoError(t, sqlMock.ExpectationsWereMet())

}

func TestLoginAccount_RefreshTokenFailed(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := SetGinTestMode()

	expectedError := errors.New("Error while creating Token")

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedDataBase, MockedUtils)

	TokenExp := 12 * time.Second
	MockedUtils.On("IsPassVerified").Return(true)
	MockedUtils.On("CreateJwtToken", TokenExp, mock.Anything).Return(models.NewToken("Token"), nil)
	MockedUtils.On("CreateJwtToken").Return(models.NewToken(""), expectedError)

	email := "Test@Test.com"
	user := uuid.New().String()
	time := time.Now()

	Account := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "first_name", "last_name", "email", "password_hashed", "subscription_type", "email_verified", "email_verification_token", "request_change_pass", "account_pass_reset_token", "last_time_logged_in", "last_time_logged_out"}).
		AddRow(user, time, time, time, "Testing", "User", email, "111122222", "free", false, "", false, "", time, time)

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE email = $1 AND "accounts"."deleted_at" IS NULL ORDER BY "accounts"."id" LIMIT $2`)).
		WithArgs(email, 1).WillReturnRows(Account)

	router.POST("/login", User.LoginAccount)

	body := []byte(`{
		"email": "Test@Test.com",
		"password": "111122222"
	  }`)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.NoError(t, sqlMock.ExpectationsWereMet())

}

func TestLoginAccount_Success(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := SetGinTestMode()

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedDataBase, MockedUtils)

	MockedUtils.On("IsPassVerified").Return(true)
	MockedUtils.On("CreateJwtToken").Return(models.NewToken("Token"), nil)
	MockedUtils.On("CreateJwtToken").Return(models.NewToken("Token"), nil)

	email := "Test@Test.com"
	user := uuid.New()
	time := time.Now()

	ShopFollowing := sqlmock.NewRows([]string{"account_id", "shop_id"}).AddRow(user.String(), 1)

	shops := sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "Shop1").AddRow(2, "Shop2")

	Account := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "first_name", "last_name", "email", "password_hashed", "subscription_type", "email_verified", "email_verification_token", "request_change_pass", "account_pass_reset_token", "last_time_logged_in", "last_time_logged_out"}).
		AddRow(user.String(), time, time, time, "Testing", "User", email, "111122222", "free", false, "", false, "", time, time)

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE email = $1 AND "accounts"."deleted_at" IS NULL ORDER BY "accounts"."id" LIMIT $2`)).
		WithArgs(email, 1).WillReturnRows(Account)

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

func TestGetTokens_Success(t *testing.T) {
	ctx, _, _ := SetGinTestMode()

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: "access_token_value"})
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "refresh_token_value"})
	ctx.Request = req

	tokens, _ := controllers.GetTokens(ctx)

	assert.Equal(t, 2, len(tokens))

}

func TestGetTokens_Fail(t *testing.T) {
	ctx, _, _ := SetGinTestMode()

	req := httptest.NewRequest("GET", "/", nil)
	ctx.Request = req

	_, err := controllers.GetTokens(ctx)

	assert.Error(t, err)

}

func TestLogOutAccount_Success_No_Cookie(t *testing.T) {
	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := SetGinTestMode()

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedDataBase, MockedUtils)

	router.GET("/logout", User.LogOutAccount)

	req, _ := http.NewRequest("GET", "/logout", bytes.NewBuffer([]byte{}))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

}

func TestLogOutAccount_SuccessWithCookies(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	ctx, router, w := SetGinTestMode()

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedDataBase, MockedUtils)

	account := uuid.New()
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

func TestLogOutAccount_UserNotFound(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	ctx, router, w := SetGinTestMode()

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedDataBase, MockedUtils)

	account := uuid.New()
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

func TestLogOutAccount_(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	ctx, router, w := SetGinTestMode()

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedDataBase, MockedUtils)

	account := uuid.New()
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

func TestVerifyAccount_NoTranID(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := SetGinTestMode()

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedDataBase, MockedUtils)

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE email_verification_token = $1 AND "accounts"."deleted_at" IS NULL`)).WillReturnError(errors.New("TransID is not found"))

	router.GET("/verifyaccount", User.VerifyAccount)

	req, _ := http.NewRequest("GET", "/verifyaccount", bytes.NewBuffer([]byte{}))

	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.NoError(t, sqlMock.ExpectationsWereMet())

}
func TestVerifyAccount_EmptyAccount(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := SetGinTestMode()

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedDataBase, MockedUtils)

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

func TestVerifyAccount_AlreadyVerified(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := SetGinTestMode()

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedDataBase, MockedUtils)

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
func TestVerifyAccount_Success(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := SetGinTestMode()

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedDataBase, MockedUtils)

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

func TestChangePass_FailedBindJson(t *testing.T) {

	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	c, router, w := SetGinTestMode()

	currentUserUUID := uuid.New()

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedDataBase, MockedUtils)

	router.POST("/changepassword", func(ctx *gin.Context) {
		ctx.Set("currentUserUUID", currentUserUUID)

	}, User.ChangePass)

	c.Request, _ = http.NewRequest("POST", "/changepassword", bytes.NewBuffer([]byte{}))

	router.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestChangePass_UserNotFound(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	c, router, w := SetGinTestMode()

	currentUserUUID := uuid.New()

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedDataBase, MockedUtils)

	router.POST("/changepassword", func(ctx *gin.Context) {
		ctx.Set("currentUserUUID", currentUserUUID)

	}, User.ChangePass)
	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE ID = $1 AND "accounts"."deleted_at" IS NULL ORDER BY "accounts"."id" LIMIT $2`)).
		WithArgs(currentUserUUID, 1).WillReturnError(errors.New("record not found"))

	c.Request, _ = http.NewRequest("POST", "/changepassword", bytes.NewBuffer([]byte(`{
		"current_password":"qqqq1111",
		"new_password":"1111qqqq",
		"confirm_password":"1111qqqq"
	}`)))

	router.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}
func TestChangePass_EmptyAccount(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	c, router, w := SetGinTestMode()

	currentUserUUID := uuid.New()
	emptyAccount := models.Account{}

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedDataBase, MockedUtils)

	router.POST("/changepassword", func(ctx *gin.Context) {
		ctx.Set("currentUserUUID", currentUserUUID)

	}, User.ChangePass)

	Account := sqlmock.NewRows([]string{"created_at", "updated_at", "first_name", "last_name", "email", "password_hashed", "subscription_type", "email_verified", "email_verification_token", "request_change_pass", "account_pass_reset_token", "last_time_logged_in", "last_time_logged_out"}).
		AddRow(emptyAccount.CreatedAt, emptyAccount.UpdatedAt, emptyAccount.FirstName, emptyAccount.LastName, emptyAccount.Email, emptyAccount.PasswordHashed, emptyAccount.SubscriptionType, emptyAccount.EmailVerified, emptyAccount.EmailVerificationToken, emptyAccount.RequestChangePass, emptyAccount.AccountPassResetToken, emptyAccount.LastTimeLoggedIn, emptyAccount.LastTimeLoggedOut)

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE ID = $1 AND "accounts"."deleted_at" IS NULL ORDER BY "accounts"."id" LIMIT $2`)).
		WithArgs(currentUserUUID, 1).WillReturnRows(Account)

	c.Request, _ = http.NewRequest("POST", "/changepassword", bytes.NewBuffer([]byte(`{
		"current_password":"qqqq1111",
		"new_password":"1111qqqq",
		"confirm_password":"1111qqqq"
	}`)))

	router.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}
func TestChangePass_PassNoMatch(t *testing.T) {

	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	c, router, w := SetGinTestMode()

	currentUserUUID := uuid.New()

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedDataBase, MockedUtils)

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
func TestChangePass_HashFail(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	c, router, w := SetGinTestMode()

	currentUserUUID := uuid.New()
	emptyAccount := models.Account{}

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedDataBase, MockedUtils)

	MockedUtils.On("IsPassVerified").Return(true)
	MockedUtils.On("HashPass").Return("", errors.New("Error while hashing pass"))

	router.POST("/changepassword", func(ctx *gin.Context) {
		ctx.Set("currentUserUUID", currentUserUUID)

	}, User.ChangePass)

	Account := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "first_name", "last_name", "email", "password_hashed", "subscription_type", "email_verified", "email_verification_token", "request_change_pass", "account_pass_reset_token", "last_time_logged_in", "last_time_logged_out"}).
		AddRow(currentUserUUID.String(), emptyAccount.CreatedAt, emptyAccount.UpdatedAt, emptyAccount.FirstName, emptyAccount.LastName, emptyAccount.Email, emptyAccount.PasswordHashed, emptyAccount.SubscriptionType, emptyAccount.EmailVerified, emptyAccount.EmailVerificationToken, emptyAccount.RequestChangePass, emptyAccount.AccountPassResetToken, emptyAccount.LastTimeLoggedIn, emptyAccount.LastTimeLoggedOut)

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE ID = $1 AND "accounts"."deleted_at" IS NULL ORDER BY "accounts"."id" LIMIT $2`)).
		WithArgs(currentUserUUID, 1).WillReturnRows(Account)

	c.Request, _ = http.NewRequest("POST", "/changepassword", bytes.NewBuffer([]byte(`{
		"current_password":"qqqq1111",
		"new_password":"1111qqqq",
		"confirm_password":"1111qqqq"
	}`)))

	router.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusConflict, w.Code)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestChangePass_PassNotConfirmed(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	c, router, w := SetGinTestMode()

	currentUserUUID := uuid.New()
	emptyAccount := models.Account{}

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedDataBase, MockedUtils)

	MockedUtils.On("IsPassVerified").Return(true)

	router.POST("/changepassword", func(ctx *gin.Context) {
		ctx.Set("currentUserUUID", currentUserUUID)

	}, User.ChangePass)

	Account := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "first_name", "last_name", "email", "password_hashed", "subscription_type", "email_verified", "email_verification_token", "request_change_pass", "account_pass_reset_token", "last_time_logged_in", "last_time_logged_out"}).
		AddRow(currentUserUUID.String(), emptyAccount.CreatedAt, emptyAccount.UpdatedAt, emptyAccount.FirstName, emptyAccount.LastName, emptyAccount.Email, emptyAccount.PasswordHashed, emptyAccount.SubscriptionType, emptyAccount.EmailVerified, emptyAccount.EmailVerificationToken, emptyAccount.RequestChangePass, emptyAccount.AccountPassResetToken, emptyAccount.LastTimeLoggedIn, emptyAccount.LastTimeLoggedOut)

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE ID = $1 AND "accounts"."deleted_at" IS NULL ORDER BY "accounts"."id" LIMIT $2`)).
		WithArgs(currentUserUUID, 1).WillReturnRows(Account)

	c.Request, _ = http.NewRequest("POST", "/changepassword", bytes.NewBuffer([]byte(`{
		"current_password":"qqqq1111",
		"new_password":"1111qqq1",
		"confirm_password":"1111qqqq"
	}`)))

	router.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}
func TestChangePass_Success(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	c, router, w := SetGinTestMode()

	currentUserUUID := uuid.New()
	emptyAccount := models.Account{}

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedDataBase, MockedUtils)

	MockedUtils.On("IsPassVerified").Return(true)
	MockedUtils.On("HashPass").Return("SomePass", nil)

	router.POST("/changepassword", func(ctx *gin.Context) {
		ctx.Set("currentUserUUID", currentUserUUID)

	}, User.ChangePass)

	Account := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "first_name", "last_name", "email", "password_hashed", "subscription_type", "email_verified", "email_verification_token", "request_change_pass", "account_pass_reset_token", "last_time_logged_in", "last_time_logged_out"}).
		AddRow(currentUserUUID.String(), emptyAccount.CreatedAt, emptyAccount.UpdatedAt, emptyAccount.FirstName, emptyAccount.LastName, emptyAccount.Email, emptyAccount.PasswordHashed, emptyAccount.SubscriptionType, emptyAccount.EmailVerified, emptyAccount.EmailVerificationToken, emptyAccount.RequestChangePass, emptyAccount.AccountPassResetToken, emptyAccount.LastTimeLoggedIn, emptyAccount.LastTimeLoggedOut)

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE ID = $1 AND "accounts"."deleted_at" IS NULL ORDER BY "accounts"."id" LIMIT $2`)).
		WithArgs(currentUserUUID, 1).WillReturnRows(Account)

	c.Request, _ = http.NewRequest("POST", "/changepassword", bytes.NewBuffer([]byte(`{
		"current_password":"qqqq1111",
		"new_password":"1111qqqq",
		"confirm_password":"1111qqqq"
	}`)))

	router.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestForgotPassReq_FailedBindJson(t *testing.T) {

	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	c, router, w := SetGinTestMode()

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedDataBase, MockedUtils)

	router.POST("/forgotpassword", User.ForgotPassReq)

	c.Request, _ = http.NewRequest("POST", "/forgotpassword", bytes.NewBuffer([]byte{}))

	router.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusNotFound, w.Code)

}
func TestForgotPassReq_UserNotFound(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	c, router, w := SetGinTestMode()

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedDataBase, MockedUtils)

	email := "Some@Test.com"
	emptyAccount := models.Account{}

	Account := sqlmock.NewRows([]string{"created_at", "updated_at", "first_name", "last_name", "email", "password_hashed", "subscription_type", "email_verified", "email_verification_token", "request_change_pass", "account_pass_reset_token", "last_time_logged_in", "last_time_logged_out"}).
		AddRow(emptyAccount.CreatedAt, emptyAccount.UpdatedAt, emptyAccount.FirstName, emptyAccount.LastName, emptyAccount.Email, emptyAccount.PasswordHashed, emptyAccount.SubscriptionType, emptyAccount.EmailVerified, emptyAccount.EmailVerificationToken, emptyAccount.RequestChangePass, emptyAccount.AccountPassResetToken, emptyAccount.LastTimeLoggedIn, emptyAccount.LastTimeLoggedOut)

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE email = $1 AND "accounts"."deleted_at" IS NULL ORDER BY "accounts"."id" LIMIT $2`)).
		WithArgs(email, 1).WillReturnRows(Account)

	router.POST("/forgotpassword", User.ForgotPassReq)

	c.Request, _ = http.NewRequest("POST", "/forgotpassword", bytes.NewBuffer([]byte(`{"email_account":"Some@Test.com"}`)))

	router.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestForgotPassReq_VerificationToken(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	c, router, w := SetGinTestMode()

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedDataBase, MockedUtils)

	MockedUtils.On("CreateVerificationString").Return("", errors.New("failed to create verification token"))

	email := "Some@Test.com"
	emptyAccount := models.Account{}
	user := uuid.New()

	Account := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "first_name", "last_name", "email", "password_hashed", "subscription_type", "email_verified", "email_verification_token", "request_change_pass", "account_pass_reset_token", "last_time_logged_in", "last_time_logged_out"}).
		AddRow(user.String(), emptyAccount.CreatedAt, emptyAccount.UpdatedAt, emptyAccount.FirstName, emptyAccount.LastName, emptyAccount.Email, emptyAccount.PasswordHashed, emptyAccount.SubscriptionType, emptyAccount.EmailVerified, emptyAccount.EmailVerificationToken, emptyAccount.RequestChangePass, emptyAccount.AccountPassResetToken, emptyAccount.LastTimeLoggedIn, emptyAccount.LastTimeLoggedOut)

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE email = $1 AND "accounts"."deleted_at" IS NULL ORDER BY "accounts"."id" LIMIT $2`)).
		WithArgs(email, 1).WillReturnRows(Account)

	router.POST("/forgotpassword", User.ForgotPassReq)

	c.Request, _ = http.NewRequest("POST", "/forgotpassword", bytes.NewBuffer([]byte(`{"email_account":"Some@Test.com"}`)))

	router.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestForgotPassReq_Success(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	c, router, w := SetGinTestMode()

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedDataBase, MockedUtils)

	MockedUtils.On("CreateVerificationString").Return("SomeToken", nil)
	MockedUtils.On("SendResetPassEmail")

	email := "Some@Test.com"
	emptyAccount := models.Account{}
	user := uuid.New()

	Account := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "first_name", "last_name", "email", "password_hashed", "subscription_type", "email_verified", "email_verification_token", "request_change_pass", "account_pass_reset_token", "last_time_logged_in", "last_time_logged_out"}).
		AddRow(user.String(), emptyAccount.CreatedAt, emptyAccount.UpdatedAt, emptyAccount.FirstName, emptyAccount.LastName, emptyAccount.Email, emptyAccount.PasswordHashed, emptyAccount.SubscriptionType, emptyAccount.EmailVerified, emptyAccount.EmailVerificationToken, emptyAccount.RequestChangePass, emptyAccount.AccountPassResetToken, emptyAccount.LastTimeLoggedIn, emptyAccount.LastTimeLoggedOut)

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE email = $1 AND "accounts"."deleted_at" IS NULL ORDER BY "accounts"."id" LIMIT $2`)).
		WithArgs(email, 1).WillReturnRows(Account)

	sqlMock.ExpectBegin()
	sqlMock.ExpectExec(regexp.QuoteMeta(`UPDATE "accounts" SET "created_at"=$1,"updated_at"=$2,"deleted_at"=$3,"first_name"=$4,"last_name"=$5,"email"=$6,"password_hashed"=$7,"subscription_type"=$8,"email_verified"=$9,"email_verification_token"=$10,"request_change_pass"=$11,"account_pass_reset_token"=$12,"last_time_logged_in"=$13,"last_time_logged_out"=$14 WHERE "accounts"."deleted_at" IS NULL AND "id" = $15`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, "", "", "", "", "", false, "", true, "SomeToken", sqlmock.AnyArg(), sqlmock.AnyArg(), user.String()).WillReturnResult(sqlmock.NewResult(1, 16))
	sqlMock.ExpectCommit()

	router.POST("/forgotpassword", User.ForgotPassReq)

	c.Request, _ = http.NewRequest("POST", "/forgotpassword", bytes.NewBuffer([]byte(`{"email_account":"Some@Test.com"}`)))

	router.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}
func TestResetPass(t *testing.T) {

	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := SetGinTestMode()

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedDataBase, MockedUtils)

	router.POST("/resetpassword", User.ResetPass)

	body := []byte{}
	req, _ := http.NewRequest("POST", "/resetpassword", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, w.Code, http.StatusNotFound)
}

func TestResetPass_PassNoMachs(t *testing.T) {

	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()
	_, router, w := SetGinTestMode()

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedDataBase, MockedUtils)

	router.POST("/resetpassword", User.ResetPass)

	body := []byte(`{"rcp":"SomeToken","new_password":"1234qwer","confirm_password":"1234qwe"}`)
	req, _ := http.NewRequest("POST", "/resetpassword", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, w.Code, http.StatusForbidden)
}

func TestResetPass_UserNotFound(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := SetGinTestMode()

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedDataBase, MockedUtils)

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE account_pass_reset_token = $1 AND "accounts"."deleted_at" IS NULL`)).
		WithArgs("SomeToken").WillReturnError(errors.New("user not found"))

	router.POST("/resetpassword", User.ResetPass)

	body := []byte(`{"rcp":"SomeToken","new_password":"1234qwer","confirm_password":"1234qwer"}`)
	req, _ := http.NewRequest("POST", "/resetpassword", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, w.Code, http.StatusForbidden)
}
func TestResetPass_EmptyAccount(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := SetGinTestMode()

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedDataBase, MockedUtils)

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
func TestResetPass_RCP_Token_no_Match(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := SetGinTestMode()

	MockedUtils := &mockUtils{}
	User := controllers.NewUserController(MockedDataBase, MockedUtils)

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
func TestResetPass_Success(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := SetGinTestMode()

	MockedUtils := &mockUtils{}
	MockedUtils.On("HashPass").Return("NewHasshedPass", nil)
	User := controllers.NewUserController(MockedDataBase, MockedUtils)

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
func TestUpdateLastTimeLoggedIn(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	MockedUtils := &mockUtils{}

	User := controllers.NewUserController(MockedDataBase, MockedUtils)

	Account := models.Account{}
	Account.ID = uuid.New()

	sqlMock.ExpectBegin()
	sqlMock.ExpectExec(regexp.QuoteMeta(`UPDATE "accounts" SET "last_time_logged_in"=$1,"updated_at"=$2 WHERE id = $3 AND "accounts"."deleted_at" IS NULL AND "id" = $4`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), Account.ID, Account.ID).WillReturnResult(sqlmock.NewResult(1, 2))
	sqlMock.ExpectCommit()

	User.UpdateLastTimeLoggedIn(&Account)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestUpdateLastTimeLoggedIn_Failed(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	MockedUtils := &mockUtils{}

	User := controllers.NewUserController(MockedDataBase, MockedUtils)

	Account := models.Account{}

	sqlMock.ExpectBegin()
	sqlMock.ExpectExec(regexp.QuoteMeta(`UPDATE "accounts" SET "last_time_logged_in"=$1,"updated_at"=$2 WHERE id = $3 AND "accounts"."deleted_at" IS NULL`)).WillReturnError(errors.New(("user not found")))
	sqlMock.ExpectRollback()

	err := User.UpdateLastTimeLoggedIn(&Account)
	assert.Contains(t, err.Error(), "user not found")
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestJoinShopFollowing(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	MockedUtils := &mockUtils{}

	User := controllers.NewUserController(MockedDataBase, MockedUtils)

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
func TestJoinShopFollowing_FAIL(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	MockedUtils := &mockUtils{}

	User := controllers.NewUserController(MockedDataBase, MockedUtils)

	Account := models.Account{}

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE "accounts"."id" = $1 AND "accounts"."deleted_at" IS NULL ORDER BY "accounts"."id" LIMIT $2`)).
		WithArgs(Account.ID, 1).WillReturnError(errors.New("No User Found"))

	err := User.JoinShopFollowing(&Account)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "No User Found")
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestGenerateLoginResponce(t *testing.T) {

	user := controllers.User{}
	Account := models.Account{FirstName: "John", Email: "test@Test.com", ShopsFollowing: []models.Shop{{Name: "ExampleShopName"}, {Name: "ExampleShop2"}}}
	AccessToken := models.Token("Example Token")
	RefreshToken := models.Token("Example Token")

	loginResponse := user.GenerateLoginResponce(&Account, &AccessToken, &RefreshToken)

	assert.Equal(t, &AccessToken, loginResponse.AccessToken)
	assert.Equal(t, &RefreshToken, loginResponse.RefreshToken)
	assert.Equal(t, len(Account.ShopsFollowing), len(loginResponse.User.Shops))

	for i := 0; i < len(Account.ShopsFollowing); i++ {
		assert.Equal(t, Account.ShopsFollowing[i].Name, loginResponse.User.Shops[i].Name, "Shop name are match")
	}

}

func TestUpdateLastTimeLoggedOut_Success(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	MockedUtils := &mockUtils{}

	User := controllers.NewUserController(MockedDataBase, MockedUtils)

	Account := models.Account{}
	Account.ID = uuid.New()

	sqlMock.ExpectBegin()
	sqlMock.ExpectExec(regexp.QuoteMeta(`UPDATE "accounts" SET "last_time_logged_out"=$1,"updated_at"=$2 WHERE id = $3 AND "accounts"."deleted_at" IS NULL`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), Account.ID).WillReturnResult(sqlmock.NewResult(1, 2))
	sqlMock.ExpectCommit()

	User.UpdateLastTimeLoggedOut(Account.ID)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestUpdateLastTimeLoggedOut_Failed(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	MockedUtils := &mockUtils{}

	User := controllers.NewUserController(MockedDataBase, MockedUtils)

	Account := models.Account{}

	sqlMock.ExpectBegin()
	sqlMock.ExpectExec(regexp.QuoteMeta(`UPDATE "accounts" SET "last_time_logged_out"=$1,"updated_at"=$2 WHERE id = $3 AND "accounts"."deleted_at" IS NULL`)).WillReturnError(errors.New(("user not found")))
	sqlMock.ExpectRollback()

	err := User.UpdateLastTimeLoggedOut(Account.ID)

	assert.Contains(t, err.Error(), "user not found")
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}
