package utils

import (
	"math/rand"
	"strings"

	browser "github.com/EDDYCJY/fake-useragent"
)

func GetRandomUserAgent() string {
	userAgents := []string{
		browser.Chrome(),
		browser.Firefox(),
		browser.Safari(),
		browser.Android(),
	}

	exclude := []string{
		"windows nt",
		"iphone",
		"ipad",
	}

	for {
		newUserAgent := userAgents[rand.Intn(len(userAgents))]
		newUserAgentToLower := strings.ToLower(newUserAgent)
		matched := false

		for _, sub := range exclude {
			if strings.Contains(newUserAgentToLower, sub) {
				matched = true
				break
			}
		}

		if !matched {
			return newUserAgent
		}
	}
}
