package main

import (
	"fmt"
	"github.com/spf13/viper"
	"os"
	"os/user"
)

// Configurations exported
type Configurations struct {
	RandomURL   string
	ExchangeURL string
	SSHHost     string
	LogLevel    string
	User        string
	Passwd      string
}

var configuredRandomURL = ""
var configuredExchangeURL = ""

func fetchConfig() Configurations {
	//create default config
	defaultConfig := Configurations{
		RandomURL:   "https://api.endlesswaltz.xyz:443/api/otp",
		ExchangeURL: "wss://exchange.endlesswaltz.xyz:443/ws",
		SSHHost:     "endlesswaltz.xyz",
		LogLevel:    "Debug",
	}

	configuredRandomURL = defaultConfig.RandomURL
	configuredExchangeURL = defaultConfig.ExchangeURL

	//Config file override code
	currentUser, err := user.Current()
	if err != nil {
		fmt.Println("Unable to get current user: ", err)
		return defaultConfig
	}

	// Get the user's home directory
	configDir := fmt.Sprintf("%s/.ew", currentUser.HomeDir)
	configFile := fmt.Sprintf("%s/config.yml", configDir)

	//check if config file exists, return if not
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return defaultConfig
	}

	fmt.Println("Config file exists, reading...")

	viper.SetConfigName("config")
	viper.AddConfigPath("$HOME/.ew/")
	viper.SetConfigType("yml")

	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Viper:Error reading config file: ", err)
		return defaultConfig
	}
	err = viper.Unmarshal(&defaultConfig)
	if err != nil {
		fmt.Println("Viper:Unable to decode into struct: ", err)
		return defaultConfig
	}

	configuredRandomURL = defaultConfig.RandomURL
	configuredExchangeURL = defaultConfig.ExchangeURL

	return defaultConfig
}
