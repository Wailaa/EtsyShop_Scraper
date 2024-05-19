package setupMockServer

import (
	"database/sql"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func StartMockedDataBase() (sqlmock.Sqlmock, *sql.DB, *gorm.DB) {

	testDB, mock, err := sqlmock.New()
	if err != nil {
		panic("sqlmock.New() occurs an error")
	}

	MockedDataBase, err := gorm.Open(postgres.New(postgres.Config{
		Conn: testDB,
	}), &gorm.Config{})
	if err != nil {
		panic("Cannot open stub database")
	}
	return mock, testDB, MockedDataBase
}
