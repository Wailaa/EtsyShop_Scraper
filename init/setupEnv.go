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
