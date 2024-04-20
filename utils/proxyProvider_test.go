package utils_test

import (
	"EtsyScraper/utils"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPickProxyProviderReturnsProxySettingObject(t *testing.T) {
	Utils := utils.Utils{}
	proxySetting := Utils.PickProxyProvider()
	assert.NotNil(t, proxySetting)
	assert.IsType(t, utils.ProxySetting{}, proxySetting)
}
