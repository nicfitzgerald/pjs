package utils

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	DBPath string `mapstructure:"db_dir"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		log.Println("unable to access config file:", err)
	}
	return
}
