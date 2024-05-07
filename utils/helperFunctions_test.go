package utils_test

import (
	"EtsyScraper/utils"
	"testing"
	"time"
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
