package utils

import (
	"math/rand"
	"strconv"
	"time"
)

func SetSleep(maxSeconds int) {
	if maxSeconds < 10 {
		return
	}
	randTimeSet := time.Duration(rand.Intn(maxSeconds-10) + 10)
	time.Sleep(randTimeSet * time.Second)

}

func StringToUint(text string) (uint, error) {
	ShopIDToUint, err := strconv.ParseUint(text, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(ShopIDToUint), nil
}
