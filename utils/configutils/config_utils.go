package configutils

import (
	"log"
	"os"

	"github.com/spf13/viper"
)

func LoadConfiguration() {
	cp := os.Getenv("CONFIG_PATH")
	if len(cp) == 0 {
		log.Println("CONFIG_PATH env is empty, using default config")
		cp = "config/local.yaml"
	}
	viper.SetConfigFile(cp)
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Cannot read config file: %s", err)
	}
	log.Println("Using config file:", viper.ConfigFileUsed())
}
