package initializer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfigError(t *testing.T) {

	path := "."

	config := LoadProjConfig(path)

	assert.Empty(t, config)
}
