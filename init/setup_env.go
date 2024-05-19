package initializer

import (
	"log"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	DataBaseHost         string `mapstructure:"POSTGRES_HOST"`
	DataBaseUserName     string `mapstructure:"POSTGRES_USERNAME"`
	DatabaseUserPassword string `mapstructure:"POSTGRES_PASSWORD"`
	DataBaseName         string `mapstructure:"POSTGRES_DB"`
	DataBasePort         string `mapstructure:"POSTGRES_PORT"`
	ServerPort           string `mapstructure:"PORT"`
	ClientOrigin         string `mapstructure:"CLIENT_ORIGIN"`

	JwtSecret   string        `mapstructure:"JWT_SECRET"`
	AccTokenExp time.Duration `mapstructure:"ACCESS_TOKEN_DUARATION"`
	RefTokenExp time.Duration `mapstructure:"REFRESH_TOKEN_DUARATION"`

	EmailAddress string `mapstructure:"PROJECT_EMAIL_ADDRESS"`
	SMTPHost     string `mapstructure:"MAILTRAP_SMTP_HOST"`
	SMTPPass     string `mapstructure:"MAILTRAP_SMTP_PASS"`
	SMTPPort     int    `mapstructure:"MAILTRAP_SMTP_PORT"`
	SMTPUser     string `mapstructure:"MAILTRAP_SMTP_USER"`

	RedisURL string `mapstructure:"REDISURL"`

	ScrapShopURL string `mapstructure:"SCRAP_SHOP_URL"`
	MaxPageLimit int    `mapstructure:"SCRAP_MAX_PAGE_LIMIT"`

	ProxyHostURL1 string `mapstructure:"PROXY_HOST_URL1"`
	ProxyHostURL2 string `mapstructure:"PROXY_HOST_URL2"`
	ProxyHostURL3 string `mapstructure:"PROXY_HOST_URL3"`
}

func LoadProjConfig(path string) (config Config) {
	viper.AddConfigPath(path)
	viper.SetConfigType("env")
	viper.SetConfigName("project")
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		return
	}
	config = Config{}
	err = viper.Unmarshal(&config)
	if err != nil {
		log.Fatal("Unmarshalling of configuration failed.")
	}
	return config
}
