package utils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"EtsyScraper/utils"
)

func TestPickProxyProviderReturnsProxySettingObject(t *testing.T) {
	Utils := utils.Utils{}
	proxySetting := Utils.PickProxyProvider()
	assert.NotNil(t, proxySetting)
	assert.IsType(t, utils.ProxySetting{}, proxySetting)
}
