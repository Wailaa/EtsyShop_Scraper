package utils

import (
	initializer "EtsyScraper/init"
	"math/rand"
	"strings"
)

var config = initializer.LoadProjConfig(".")
var GetAllEnvProy = strings.Split(config.ProxyHostURL, ";")

func PickProxyProvider() string {

	RandomIndex := rand.Intn(len(GetAllEnvProy))
	ProxyProvider := GetAllEnvProy[RandomIndex]

	return ProxyProvider
}
