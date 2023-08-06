package main

import (
	"github.com/spf13/viper"
)

type Config struct {
	Token string
}

func LoadConfig() (config Config, err error) {
	viper.SetConfigType("env")
	viper.SetConfigFile(".env")

	if err = viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found
			panic("Config file not found")
		} else {
			// Config file was found but another error was produced
			panic("Error reading config file")
		}
	}

	err = viper.Unmarshal(&config)
	return config, err
}
