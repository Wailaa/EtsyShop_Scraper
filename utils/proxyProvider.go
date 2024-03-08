package utils

import (
	initializer "EtsyScraper/init"
	"fmt"
	"math/rand"
	"strings"
)

var config = initializer.LoadProjConfig(".")
var GetAllEnvProy = []string{config.ProxyHostURL1, config.ProxyHostURL2, config.ProxyHostURL3}
var Countries = []string{"UK", "FR", "DE", "US", "IR", "IT", "SP"}

type ProxySetting struct {
	Provider string
	Url      string
}

func PickProxyProvider() ProxySetting {
	ProxySettings := ProxySetting{}

	RandomIndex := rand.Intn(len(GetAllEnvProy))

	CountryList := strings.Split(GetAllEnvProy[RandomIndex], ";")
	Country := rand.Intn(len(CountryList))

	ProxySettings.Provider = fmt.Sprint("Provider ", RandomIndex+1, " Country :", Countries[Country])
	ProxySettings.Url = CountryList[Country]

	return ProxySettings
}
