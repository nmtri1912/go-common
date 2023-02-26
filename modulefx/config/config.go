package config

import (
	"log"
	"os"

	"github.com/spf13/viper"
)

func LoadConfiguration() {
	cp := os.Getenv("CONFIG_PATH")
	if len(cp) == 0 {
		log.Println("CONFIG_PATH env is empty!. Using local config")
		cp = "config/local.yaml"
	}
	viper.SetConfigFile(cp)
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal("Cannot read config file: ", err)
	}
	log.Println("Using config file:", viper.ConfigFileUsed())
}
