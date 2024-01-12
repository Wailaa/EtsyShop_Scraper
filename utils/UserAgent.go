package utils

import (
	"math/rand"

	browser "github.com/EDDYCJY/fake-useragent"
)

func GetRandomUserAgent() string {
	userAgents := []string{
		browser.Chrome(),
		browser.Firefox(),
		browser.Safari(),
		browser.Android(),
		browser.MacOSX(),
	}
	NewUserAgent := userAgents[rand.Intn(len(userAgents))]

	return NewUserAgent
}
