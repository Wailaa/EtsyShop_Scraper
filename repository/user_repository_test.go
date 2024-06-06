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
