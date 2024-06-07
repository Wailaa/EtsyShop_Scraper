package repository_test

import (
	"EtsyScraper/models"
	"EtsyScraper/repository"
	setupMockServer "EtsyScraper/setupTests"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestGetAccountByIDSuccess(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	User := repository.DataBase{DB: MockedDataBase}

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

func TestGetAccountByIDFail(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	User := repository.DataBase{DB: MockedDataBase}

	expectedError := errors.New("no Account Found")
	user := uuid.New()

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE ID = $1 AND "accounts"."deleted_at" IS NULL ORDER BY "accounts"."id" LIMIT $2`)).
		WithArgs(user, 1).WillReturnError(expectedError)

	User.GetAccountByID(user)

	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestGetAccountByEmailSuccess(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	User := repository.DataBase{DB: MockedDataBase}
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

func TestUpdateLastTimeLoggedIn(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	User := repository.DataBase{DB: MockedDataBase}

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

	User := repository.DataBase{DB: MockedDataBase}

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

	User := repository.DataBase{DB: MockedDataBase}

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
func TestJoinShopFollowingFAIL(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	User := repository.DataBase{DB: MockedDataBase}

	Account := models.Account{}

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE "accounts"."id" = $1 AND "accounts"."deleted_at" IS NULL ORDER BY "accounts"."id" LIMIT $2`)).
		WithArgs(Account.ID, 1).WillReturnError(errors.New("No User Found"))

	err := User.JoinShopFollowing(&Account)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "No User Found")
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestUpdateLastTimeLoggedOutSuccess(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	User := repository.DataBase{DB: MockedDataBase}

	Account := models.Account{}
	Account.ID = uuid.New()

	sqlMock.ExpectBegin()
	sqlMock.ExpectExec(regexp.QuoteMeta(`UPDATE "accounts" SET "last_time_logged_out"=$1,"updated_at"=$2 WHERE id = $3 AND "accounts"."deleted_at" IS NULL`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), Account.ID).WillReturnResult(sqlmock.NewResult(1, 2))
	sqlMock.ExpectCommit()

	User.UpdateLastTimeLoggedOut(Account.ID)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestUpdateLastTimeLoggedOutFailed(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	User := repository.DataBase{DB: MockedDataBase}

	Account := models.Account{}

	sqlMock.ExpectBegin()
	sqlMock.ExpectExec(regexp.QuoteMeta(`UPDATE "accounts" SET "last_time_logged_out"=$1,"updated_at"=$2 WHERE id = $3 AND "accounts"."deleted_at" IS NULL`)).WillReturnError(errors.New(("user not found")))
	sqlMock.ExpectRollback()

	err := User.UpdateLastTimeLoggedOut(Account.ID)

	assert.Contains(t, err.Error(), "user not found")
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestUpdateAccountAfterVerify(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	User := repository.DataBase{DB: MockedDataBase}

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

func TestUpdateAccountAfterVerifyFail(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	User := repository.DataBase{DB: MockedDataBase}

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

func TestUpdateAccountNewPass(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	User := repository.DataBase{DB: MockedDataBase}

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

func TestUpdateAccountNewPassFail(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	User := repository.DataBase{DB: MockedDataBase}

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

func TestUpdateAccountAfterResetPass(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	User := repository.DataBase{DB: MockedDataBase}

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

func TestUpdateAccountAfterResetFail(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	User := repository.DataBase{DB: MockedDataBase}

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

func TestSaveAccount(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	Account := &models.Account{}
	Account.ID = uuid.New()

	User := repository.DataBase{DB: MockedDataBase}

	sqlMock.ExpectBegin()
	sqlMock.ExpectExec(regexp.QuoteMeta(`UPDATE "accounts" SET "created_at"=$1,"updated_at"=$2,"deleted_at"=$3,"first_name"=$4,"last_name"=$5,"email"=$6,"password_hashed"=$7,"subscription_type"=$8,"email_verified"=$9,"email_verification_token"=$10,"request_change_pass"=$11,"account_pass_reset_token"=$12,"last_time_logged_in"=$13,"last_time_logged_out"=$14 WHERE "accounts"."deleted_at" IS NULL AND "id" = $15`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	sqlMock.ExpectCommit()

	err := User.SaveAccount(Account)

	assert.NoError(t, err)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}
func TestSaveAccountFail(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	Account := &models.Account{}
	Account.ID = uuid.New()

	User := repository.DataBase{DB: MockedDataBase}

	sqlMock.ExpectBegin()
	sqlMock.ExpectExec(regexp.QuoteMeta(`UPDATE "accounts" SET "created_at"=$1,"updated_at"=$2,"deleted_at"=$3,"first_name"=$4,"last_name"=$5,"email"=$6,"password_hashed"=$7,"subscription_type"=$8,"email_verified"=$9,"email_verification_token"=$10,"request_change_pass"=$11,"account_pass_reset_token"=$12,"last_time_logged_in"=$13,"last_time_logged_out"=$14 WHERE "accounts"."deleted_at" IS NULL AND "id" = $15`)).
		WillReturnError(errors.New("error while saving to database"))
	sqlMock.ExpectRollback()

	err := User.SaveAccount(Account)

	assert.Error(t, err)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestCreateAccount(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	newUUID := uuid.New()

	newAccount := &models.Account{
		ID:                     newUUID,
		FirstName:              "Example",
		LastName:               "Test",
		Email:                  "Example@Exampleemail.com",
		PasswordHashed:         "asdasdasd",
		SubscriptionType:       "free",
		EmailVerificationToken: "JustAnotherToken",
	}

	User := repository.DataBase{DB: MockedDataBase}

	sqlMock.ExpectBegin()
	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "accounts" ("id","created_at","updated_at","deleted_at","first_name","last_name","email","password_hashed","subscription_type","email_verified","email_verification_token","request_change_pass","account_pass_reset_token","last_time_logged_in","last_time_logged_out") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15) RETURNING "id"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), "Example", "Test", "Example@Exampleemail.com", "asdasdasd", "free", false, "JustAnotherToken", false, "", sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"1", "15"}))
	sqlMock.ExpectCommit()

	_, err := User.CreateAccount(newAccount)

	assert.NoError(t, err)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}
func TestCreateAccountFail(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	newUUID := uuid.New()

	newAccount := &models.Account{
		ID:                     newUUID,
		FirstName:              "Example",
		LastName:               "Test",
		Email:                  "Example@Exampleemail.com",
		PasswordHashed:         "asdasdasd",
		SubscriptionType:       "free",
		EmailVerificationToken: "JustAnotherToken",
	}

	User := repository.DataBase{DB: MockedDataBase}

	sqlMock.ExpectBegin()
	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "accounts" ("id","created_at","updated_at","deleted_at","first_name","last_name","email","password_hashed","subscription_type","email_verified","email_verification_token","request_change_pass","account_pass_reset_token","last_time_logged_in","last_time_logged_out") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15) RETURNING "id"`)).
		WillReturnError(errors.New("error while creating account"))
	sqlMock.ExpectRollback()

	_, err := User.CreateAccount(newAccount)

	assert.Error(t, err)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestInsertTokenForAccountEmptyAccount(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	User := repository.DataBase{DB: MockedDataBase}

	emptyAccount := &models.Account{ID: uuid.Nil}

	Account := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "first_name", "last_name", "email", "password_hashed", "subscription_type", "email_verified", "email_verification_token", "request_change_pass", "account_pass_reset_token", "last_time_logged_in", "last_time_logged_out"}).
		AddRow(emptyAccount.ID, emptyAccount.CreatedAt, emptyAccount.UpdatedAt, emptyAccount.FirstName, emptyAccount.LastName, emptyAccount.Email, emptyAccount.PasswordHashed, emptyAccount.SubscriptionType, emptyAccount.EmailVerified, emptyAccount.EmailVerificationToken, emptyAccount.RequestChangePass, emptyAccount.AccountPassResetToken, emptyAccount.LastTimeLoggedIn, emptyAccount.LastTimeLoggedOut)

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE "email_verification_token" = $1 AND "accounts"."deleted_at" IS NULL`)).WillReturnRows(Account)

	User.InsertTokenForAccount("email_verification_token", "", emptyAccount)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}
func TestInsertTokenForAccount(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	User := repository.DataBase{DB: MockedDataBase}

	emptyAccount := &models.Account{ID: uuid.New(), EmailVerificationToken: "SomeToken"}

	Account := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "first_name", "last_name", "email", "password_hashed", "subscription_type", "email_verified", "email_verification_token", "request_change_pass", "account_pass_reset_token", "last_time_logged_in", "last_time_logged_out"}).
		AddRow(emptyAccount.ID.String(), emptyAccount.CreatedAt, emptyAccount.UpdatedAt, emptyAccount.FirstName, emptyAccount.LastName, emptyAccount.Email, emptyAccount.PasswordHashed, emptyAccount.SubscriptionType, emptyAccount.EmailVerified, emptyAccount.EmailVerificationToken, emptyAccount.RequestChangePass, emptyAccount.AccountPassResetToken, emptyAccount.LastTimeLoggedIn, emptyAccount.LastTimeLoggedOut)

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE "email_verification_token" = $1 AND "accounts"."deleted_at" IS NULL`)).WillReturnRows(Account)

	User.InsertTokenForAccount("email_verification_token", "SomeToken", emptyAccount)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}
