package utils_test

import (
	"EtsyScraper/utils"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidUserAgent(t *testing.T) {
	Utils := &utils.Utils{}
	userAgent := Utils.GetRandomUserAgent()
	assert.True(t, strings.Contains(userAgent, "Chrome") || strings.Contains(userAgent, "Firefox") || strings.Contains(userAgent, "Safari") || strings.Contains(userAgent, "Android"))
	assert.False(t, strings.Contains(strings.ToLower(userAgent), "windows nt") || strings.Contains(strings.ToLower(userAgent), "iphone") || strings.Contains(strings.ToLower(userAgent), "ipad"))
}
