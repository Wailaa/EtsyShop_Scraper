package initializer

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func DataBaseConnect(config *Config) {

	var err error
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai", config.DataBaseHost, config.DataBaseUserName, config.DatabaseUserPassword, config.DataBaseName, config.DataBasePort)
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("a problem accured while trying to connect to the Database")
	}
	fmt.Println("Successfully connected to the database")

}
