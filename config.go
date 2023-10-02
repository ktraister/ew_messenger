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
	LogLevel    string
	User        string
	Passwd      string
}

func fetchConfig() Configurations {
	var configuration Configurations
	//create default config
	defaultConfig := Configurations{
		RandomURL:   "randomapi.endlesswaltz.xyz",
		ExchangeURL: "exchange.endlesswaltz.xyz",
		LogLevel:    "Error",
	}

	currentUser, err := user.Current()
	if err != nil {
		fmt.Println("Unable to get current user: ", err)
		return defaultConfig
	}

	// Get the user's home directory
	configDir := fmt.Sprintf("%s/.ew", currentUser.HomeDir)
	configFile := fmt.Sprintf("%s/config.yml", configDir)

	//check if config file exists, skip if not
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
	err = viper.Unmarshal(&configuration)
	if err != nil {
		fmt.Println("Viper:Unable to decode into struct: ", err)
		return defaultConfig
	}

	return configuration
}
