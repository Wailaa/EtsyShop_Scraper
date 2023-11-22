package initializer

import "github.com/spf13/viper"

type Config struct {
	DataBaseHost         string `mapstructure:"POSTGRES_HOST"`
	DataBaseUserName     string `mapstructure:"POSTGRES_USERNAME"`
	DatabaseUserPassword string `mapstructure:"POSTGRES_PASSWORD"`
	DataBaseName         string `mapstructure:"POSTGRES_DB"`
	DataBasePort         string `mapstructure:"POSTGRES_PORT"`
	ServerPort           string `mapstructure:"PORT"`
	ClientOrigin         string `mapstructure:"CLIENT_ORIGIN"`

	JwtSecret string `mapstructure:"JWT_SECRET"`

	EmailAddress string `mapstructure:"PROJECT_EMAIL_ADDRESS"`
	SMTPHost     string `mapstructure:"MAILTRAP_SMTP_HOST"`
	SMTPPass     string `mapstructure:"MAILTRAP_SMTP_PASS"`
	SMTPPort     int    `mapstructure:"MAILTRAP_SMTP_PORT"`
	SMTPUser     string `mapstructure:"MAILTRAP_SMTP_USER"`
}

func LoadProjConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigType("env")
	viper.SetConfigName("project")
	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}
	err = viper.Unmarshal(&config)
	return
}
