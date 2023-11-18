package main

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
		RandomURL:   "https://api.endlesswaltz.xyz:443/api",
		ExchangeURL: "wss://exchange.endlesswaltz.xyz:443/ws",
		SSHHost:     "endlesswaltz.xyz",
		LogLevel:    "Error",
	}

	//localdev debug config
	/*
	defaultConfig := Configurations{
		RandomURL:   "https://localhost:443/api",
		ExchangeURL: "wss://localhost:443/ws",
		SSHHost:     "localhost",
		LogLevel:    "Debug",
	}
	*/

	configuredRandomURL = defaultConfig.RandomURL
	configuredExchangeURL = defaultConfig.ExchangeURL

	return defaultConfig
}
