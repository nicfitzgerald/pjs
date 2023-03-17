package utils

import (
	"log"

	"github.com/adrg/xdg"
	"github.com/spf13/viper"
)

type Config struct {
	DBPath string `mapstructure:"db_path"`
}

func LoadConfig() (config Config, err error) {
	configFilePath, err := xdg.ConfigFile("pjs/config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	viper.SetConfigFile(configFilePath)

	dataFilePath, err := xdg.DataFile("pjs/pjs.db")
	if err != nil {
		log.Fatal(err)
	}
	viper.Set("db_path", dataFilePath)

	viper.WriteConfig()

	err = viper.ReadInConfig()
	if err != nil {
		log.Println("Unable to read the config:", err)
		return
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		log.Println("unable to access config file:", err)
	}
	return
}
