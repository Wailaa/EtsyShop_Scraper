package utils

import (
	"fmt"
	"math/rand"
	"strings"
)

// var GetAllEnvProy = []string{config.ProxyHostURL1, config.ProxyHostURL2, config.ProxyHostURL3}
var GetAllEnvProy = []string{Config.ProxyHostURL1, Config.ProxyHostURL2}
var Countries = []string{"UK", "FR", "DE", "US", "IR", "IT", "SP"}

type ProxySetting struct {
	Provider string
	Url      string
}

func (ut *Utils) PickProxyProvider() ProxySetting {
	ProxySettings := ProxySetting{}

	SelectProxyProvider := rand.Intn(len(GetAllEnvProy))

	CountryList := strings.Split(GetAllEnvProy[SelectProxyProvider], ";")
	Country := rand.Intn(len(CountryList))

	ProxySettings.Provider = fmt.Sprint("Provider ", SelectProxyProvider+1, " Country :", Countries[Country])
	ProxySettings.Url = CountryList[Country]

	return ProxySettings
}
