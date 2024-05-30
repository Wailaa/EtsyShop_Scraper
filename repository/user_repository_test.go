package repository_test

import (
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
