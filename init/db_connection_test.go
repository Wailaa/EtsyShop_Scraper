package initializer_test

import (
	initializer "EtsyScraper/init"

	"testing"
)

func TestValidConfigurationConnectionSuccess(t *testing.T) {

	config := initializer.LoadProjConfig("../")

	initializer.DataBaseConnect(&config)

	if initializer.DB == nil {
		t.Error("DB is nil after successful connection")
	}

	sqlDB, err := initializer.DB.DB()
	if err != nil {
		t.Fatalf("failed to get database instance: %v", err)
	}
	sqlDB.Close()
}
