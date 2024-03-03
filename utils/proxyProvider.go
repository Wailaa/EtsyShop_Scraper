package utils

import (
	initializer "EtsyScraper/init"
	"fmt"
	"math/rand"
)

var config = initializer.LoadProjConfig(".")
var GetAllEnvProy = []string{config.ProxyHostURL1, config.ProxyHostURL2, config.ProxyHostURL3}

type ProxySetting struct {
	Provider string
	Url      string
}

func PickProxyProvider() ProxySetting {
	ProxySettings := ProxySetting{}

	RandomIndex := rand.Intn(len(GetAllEnvProy))

	ProxySettings.Provider = fmt.Sprint("Provider ", RandomIndex+1)
	ProxySettings.Url = GetAllEnvProy[RandomIndex]

	return ProxySettings
}
