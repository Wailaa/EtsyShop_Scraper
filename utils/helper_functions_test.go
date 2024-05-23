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

func TestMarshalJSONData(t *testing.T) {
	tests := []struct {
		name          string
		data          interface{}
		expected      string
		expectedError error
	}{
		{
			name:          "normal struct with string field",
			data:          struct{ Name string }{Name: "Example"},
			expected:      `{"Name":"Example"}`,
			expectedError: nil,
		},

		{
			name:          "normal struct with uint fields",
			data:          struct{ ID uint }{ID: 19090},
			expected:      `{"ID":19090}`,
			expectedError: nil,
		},

		{
			name:          "struct with unexported field",
			data:          struct{ name string }{name: "Example"},
			expected:      "{}",
			expectedError: errors.New("json: error calling MarshalJSON for type struct { name string }"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := utils.MarshalJSONData(tc.data)

			if err != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, string(actual))
			}

		})
	}

}

func TestHandleError(t *testing.T) {

	tests := []struct {
		name          string
		message       string
		errorCase     error
		expectedError error
	}{
		{
			name: "Error should be nil",

			errorCase:     nil,
			expectedError: nil,
		},
		{
			name: "error with no additional message",

			errorCase:     errors.New("just anotehr error"),
			expectedError: errors.New("error: just anotehr error"),
		},
		{
			name:          "error with  additional message",
			message:       "another additional message",
			errorCase:     errors.New("just anotehr error"),
			expectedError: errors.New("another additional message: just anotehr error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			if len(tc.message) > 0 {
				err = utils.HandleError(tc.errorCase, tc.message)

			} else {
				err = utils.HandleError(tc.errorCase)
			}
			if tc.errorCase == nil {
				assert.NoError(t, err, "Error should be nil")
			} else {
				assert.EqualError(t, err, tc.expectedError.Error(), "Error message should match")
			}
		})
	}
}

func TestStringToFloat(t *testing.T) {
	tests := []struct {
		Price  string
		result float64
		err    error
	}{
		{
			Price:  "19.7",
			result: 19.7,
		},
		{
			Price:  "1.8",
			result: 1.8,
		},
	}

	for _, tc := range tests {
		t.Run(tc.Price, func(t *testing.T) {
			actual, _ := utils.StringToFloat(tc.Price)
			if actual != tc.result {
				t.Errorf("Expected StringToFloat to be %v, but got %v", tc.result, actual)
			}
		})
	}
}

func TestReplaceSign(t *testing.T) {
	tests := []struct {
		Price    string
		oldSign  string
		newSign  string
		expected string
	}{
		{
			Price:    "1,232$",
			oldSign:  ",",
			newSign:  "",
			expected: "1232$",
		},
		{
			Price:    "1,232$",
			oldSign:  ",",
			newSign:  "",
			expected: "1232$",
		},
		{
			Price:    "1232$",
			oldSign:  ".",
			newSign:  "",
			expected: "1232$",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Price, func(t *testing.T) {
			actual := utils.ReplaceSign(tc.Price, tc.oldSign, tc.newSign)
			if actual != tc.expected {
				t.Errorf("Expected StringToFloat to be %v, but got %v", tc.expected, actual)
			}
		})
	}
}

func TestTruncateDate(t *testing.T) {
	tests := []struct {
		input    time.Time
		expected time.Time
	}{
		{
			input:    time.Date(2024, 5, 13, 10, 30, 0, 0, time.UTC),
			expected: time.Date(2024, 5, 13, 0, 0, 0, 0, time.UTC),
		},
		{
			input:    time.Date(2021, 12, 31, 23, 59, 59, 999, time.UTC),
			expected: time.Date(2021, 12, 31, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, test := range tests {
		result := utils.TruncateDate(test.input)
		if result != test.expected {
			t.Errorf("TruncateDate(%v) = %v; want %v", test.input, result, test.expected)
		}
	}

}

func TestStringContains(t *testing.T) {
	tests := []struct {
		name     string
		str      string
		subStr   string
		expected bool
	}{
		{
			name:     "one word string",
			str:      "this is a test",
			subStr:   "test",
			expected: true,
		},
		{
			name:     "multiple words string",
			str:      "this is a test for multiple words",
			subStr:   "multiple words",
			expected: true,
		},
		{
			name:     "no match at all",
			str:      "nothing to see here",
			subStr:   "again",
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := utils.StringContains(tc.str, tc.subStr)
			if result != tc.expected {
				t.Errorf("Expected StringContains to have %v but got %v", tc.expected, result)
			}
		})
	}
}
