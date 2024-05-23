package utils_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"EtsyScraper/utils"
)

func TestValidUserAgent(t *testing.T) {
	Utils := &utils.Utils{}
	userAgent := Utils.GetRandomUserAgent()
	assert.True(t, utils.StringContains(userAgent, "Chrome") || utils.StringContains(userAgent, "Firefox") || utils.StringContains(userAgent, "Safari") || utils.StringContains(userAgent, "Android"))
	assert.False(t, utils.StringContains(strings.ToLower(userAgent), "windows nt") || utils.StringContains(strings.ToLower(userAgent), "iphone") || utils.StringContains(strings.ToLower(userAgent), "ipad"))
}
