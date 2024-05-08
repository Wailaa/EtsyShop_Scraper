package utils_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"EtsyScraper/utils"
)

func TestSetSleep(t *testing.T) {
	MaxSeconds := 20
	start := time.Now().Unix()

	utils.SetSleep(MaxSeconds)
	end := time.Now().Unix()

	if (end-start) < 10 || (end-start) > int64(MaxSeconds) {
		t.Error("Incorrect sleep function")
	}

}

func TestStringToUint(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected uint
		err      error
	}{
		{
			name:     "testing a positive number",
			text:     "19",
			expected: 19,
			err:      nil,
		},
		{
			name:     "testing a zero",
			text:     "0",
			expected: 0,
			err:      nil,
		},
		{
			name:     "testing negative number",
			text:     "-1",
			expected: 0,
			err:      errors.New("invalid syntax"),
		},
		{
			name:     "testing text",
			text:     "IsThisPossible",
			expected: 0,
			err:      errors.New("invalid syntax"),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := utils.StringToUint(tc.text)
			if actual != tc.expected {
				t.Errorf("Expected StringToUint(%s) to be %v, but got %v and error: %s", tc.text, tc.expected, actual, err.Error())
			}
			if err != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
