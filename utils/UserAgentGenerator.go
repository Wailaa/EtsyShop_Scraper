package utils

import (
	browser "github.com/EDDYCJY/fake-useragent"
)

func CreateUserAgent() string {
	client := browser.Client{
		MaxPage: 3,
	}
	cache := browser.Cache{}
	b := browser.NewBrowser(client, cache)

	userAgent := b.Random()
	return userAgent
}
