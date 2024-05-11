package utils

import (
	"encoding/json"
	"fmt"
	"log"
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

func MarshalJSONData(data interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return jsonData, nil
}

func HandleError(err error, message ...string) error {
	if err != nil {
		log.Println(err)
		if len(message) > 0 {
			return fmt.Errorf("%s: %w", message[0], err)
		}
		return fmt.Errorf("error: %w", err)
	}
	return nil
}
