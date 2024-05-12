package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
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
		return 0, HandleError(err)
	}
	return uint(ShopIDToUint), nil
}

func MarshalJSONData(data interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, HandleError(err)
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

func StringToFloat(price string) (float64, error) {
	result, err := strconv.ParseFloat(price, 64)
	if err != nil {
		return float64(0), HandleError(err)
	}
	return result, nil
}

func ReplaceSign(sentence, oldSign, newSign string) string {
	result := strings.Replace(sentence, oldSign, newSign, -1)
	return result
}
