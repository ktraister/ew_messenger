package main

// Configurations exported
type Configurations struct {
	RandomURL   string
	ExchangeURL string
	SSHHost     string
	SSHKey      string
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
		SSHKey:      "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDNIUPPef19VF/al9M1fot1+ltlm5eW3HTi7xe/MPmi4NQZTt9DdNZ7wpbTUdyaDRqU4exIjtGpzi8SxZk4uXGX78wfoJenTKyvmTIUXTCecwPkOHGATz1rOGBgXzkmrQSvV7d7gtkqhcfLW0j/kFjKrQYQalGdDGbzt0/KYTIC9FNFBDd6OJWYGWMdecSa9VoomAH1UfaeS5SwIj2K72Pm+KxxSi12ND9ZFXfRy6UB7O9m4oxMBsyBKR/5iU8RiMTlDQx66rRKxob1CdMplMJDv7X7UqxUJGVkc3ec7LNt1FoQPtmEXcRWogQzAeEsZ9g5o84eXV23yYMj5JQYMaodZ6N41nOvcLi/HFETmpe/u/LPseAw9/irRgkNjwDyzUo/gwLubSlSp6B9WaoujUojRM2l1gwxWJqwjK48PuV62SczyIU1gC6FabNFiHQOmaO/UcHueKiPohMN2LfL3je7s52K5WE3gUWR3clsFQEjf+0XHZsS11jaU/vwoEMovDM=",
		LogLevel:    "Debug",
	}

	/*
		//localdev debug config
		defaultConfig := Configurations{
			RandomURL:   "https://localhost:443/api",
			ExchangeURL: "wss://localhost:443/ws",
			SSHHost:     "localhost",
			SSHKey:     "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDNIUPPef19VF/al9M1fot1+ltlm5eW3HTi7xe/MPmi4NQZTt9DdNZ7wpbTUdyaDRqU4exIjtGpzi8SxZk4uXGX78wfoJenTKyvmTIUXTCecwPkOHGATz1rOGBgXzkmrQSvV7d7gtkqhcfLW0j/kFjKrQYQalGdDGbzt0/KYTIC9FNFBDd6OJWYGWMdecSa9VoomAH1UfaeS5SwIj2K72Pm+KxxSi12ND9ZFXfRy6UB7O9m4oxMBsyBKR/5iU8RiMTlDQx66rRKxob1CdMplMJDv7X7UqxUJGVkc3ec7LNt1FoQPtmEXcRWogQzAeEsZ9g5o84eXV23yYMj5JQYMaodZ6N41nOvcLi/HFETmpe/u/LPseAw9/irRgkNjwDyzUo/gwLubSlSp6B9WaoujUojRM2l1gwxWJqwjK48PuV62SczyIU1gC6FabNFiHQOmaO/UcHueKiPohMN2LfL3je7s52K5WE3gUWR3clsFQEjf+0XHZsS11jaU/vwoEMovDM=",
			LogLevel:    "Debug",
		}
	*/

	configuredRandomURL = defaultConfig.RandomURL
	configuredExchangeURL = defaultConfig.ExchangeURL

	return defaultConfig
}
