package scrap

import (
	initializer "EtsyScraper/init"
	setupMockServer "EtsyScraper/setupTests"

	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckForUpdates_Success(t *testing.T) {

	mockConfig := initializer.Config{
		ProxyHostURL1: "",
		ProxyHostURL2: "",
		ProxyHostURL3: "",
	}
	Config = mockConfig
	setupMockServer.GlobalTestSetupMockServer("../setupTests/testing.html")

	defer setupMockServer.MockServer.Close()

	mockURL := setupMockServer.MockServer.URL
	response, err := CheckForUpdates(mockURL, false)
	if err != nil {
		t.Errorf("CheckForUpdates failed: %v", err)
	}

	assert.Equal(t, 2072, response.TotalSales)
	assert.Equal(t, 694, response.Admirers)
	assert.Equal(t, false, response.OnVacation)

}
