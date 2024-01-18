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
	BinVersion  string
	CertData    CertData
}

type CertData struct {
	Subject  string
	Issuer   string
	DNSNames string
	SigAlgo  string
	PubAlgo  string
}

var configuredRandomURL = ""
var configuredExchangeURL = ""

func fetchConfig() Configurations {
	//create default config
	defaultConfig := Configurations{
		RandomURL:   "https://api.endlesswaltz.xyz:443/api",
		ExchangeURL: "wss://exchange.endlesswaltz.xyz:443/ws",
		SSHHost:     "endlesswaltz.xyz",
		SSHKey:      "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAEAQDBPsQPG5TOY8sdoanRbdbbr09YD7INldhI6ygw4V4pdVQJ6kS5xWTYyWBkP5s5xtQjQv+ORLqko4h2suGQ/86cbbx66PX8rxHrdMC+AmNeQ19bGBvxt0LBz2FBkHa8O3nlDgwb9/qy8QpTyLrPMu666+2FjIZbfLdHVyDgOgI3u/s2p86d8ggB3/Yt0CfAjCLKQVcxekODR7Nl/eD4Lc+y4MM6YxcQOCrkqWZBlDeQ2PKg8rVVvFN2bh+wq3CwRLMQ4gNNgUIvz0a59klXRlxs8tNwO8aEC6H+DGRU1AVTHLUQPBVK17j1fmKBdwL4hDzxVwdjJ+vUtkF8Vw9Sv706wamYlsW8NccEkDrZonvHJFClxcDgMyEkNdBoaUv+8O8Rd4SOxV1PgIh0g+K+wvkZG1OVp1SRZHkhJ7ozwMmxOHVN24Q6zyv3j3s3f259SpjhAIhojLLlrkeM3GUBzpy5Px24KOZNjb7Ckafpx5gRPQqTnr+kujhYKsGvw0ksUv8MeZlHSdqhwFL6XIqrdddpqfotXEJYSyEnNRobfVKxYW6asdAdq1tTOfNKMtHJoAkmNGkulljw0D7mKJpOzQxmI+aYzhXwbjkEmOCZvGGtj0HShIZBJcnHN4NvwFGLIqGAsf86fzPm6Kw2fy6a4t1i7WD58ngcsnUviMgx1FyJ7y54h4XcIADTN5azDs+DfJgQw3QYE04Xl96PL3LHtsYA84S8+KWGjxypSfdPVFEs5zHtXMgbGuVtyLxNl8O7rzpAK4Ck9dSHhQcPH7gjGrk82e/7mYDqv0Ylf+3tOwkamO2GEkr41Nly0N8TqIQyME6pkJ+hZA8xiYQpdcIVPVoc6W0jTfmG7BVF9cwNZRyPhCtzH427hMClNbtpuipH88LO3NYQD6osTDZoVRP3/nt5eyZueo3otpwpQXwDj9L2Wt8euaa6yYGQ1CHZJr1BYEkc9H8q0kEDoV0qs061g6CkgTB5vmOoNVlahgkgbXc5FiJuQTkMfRtvNwBaIxVWULH95zW8rb1/2sAXujioLxll03tRpc1HNAcCFvUneFqvy1bJ3t5cVoLn5nWaqoBFCebzYdEc+/Yx5qndHMBZz+5BgLLoYWF696xnrQFHpiuGzz5IGM53ox6MPkXK+u4EiEXHifiZ2eH4TCzFA4+5ztsXv3DOv/mYJ0QolVt3t7JugPsu76KxK0ZHqcHbmdZGy3vvwOkbokDqnJTTrppuZzSW+FMUgyvlkJusKXdCdgqE+KQepvBaNRkWKaj75ycxszqqQgQOf2x2IpMQI1wPcEQkbivcT7eNWAzprCG1i57irbE448QKmyImPke3ZvNxIkd4JCBHVAZrB13xU9imuwTz",
		LogLevel:    "Error",
		BinVersion:  "REPLACEME",
		CertData: CertData{
			"CN=endlesswaltz.xyz",
			"CN=R3,O=Let's Encrypt,C=US",
			"[*.endlesswaltz.xyz endlesswaltz.xyz]",
			"SHA256-RSA",
			"ECDSA",
		},
	}
	/*
			//localdev debug config
			defaultConfig := Configurations{
				RandomURL:   "https://localhost:443/api",
				ExchangeURL: "wss://localhost:443/ws",
				SSHHost:     "localhost",
				SSHKey:      "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDNIUPPef19VF/al9M1fot1+ltlm5eW3HTi7xe/MPmi4NQZTt9DdNZ7wpbTUdyaDRqU4exIjtGpzi8SxZk4uXGX78wfoJenTKyvmTIUXTCecwPkOHGATz1rOGBgXzkmrQSvV7d7gtkqhcfLW0j/kFjKrQYQalGdDGbzt0/KYTIC9FNFBDd6OJWYGWMdecSa9VoomAH1UfaeS5SwIj2K72Pm+KxxSi12ND9ZFXfRy6UB7O9m4oxMBsyBKR/5iU8RiMTlDQx66rRKxob1CdMplMJDv7X7UqxUJGVkc3ec7LNt1FoQPtmEXcRWogQzAeEsZ9g5o84eXV23yYMj5JQYMaodZ6N41nOvcLi/HFETmpe/u/LPseAw9/irRgkNjwDyzUo/gwLubSlSp6B9WaoujUojRM2l1gwxWJqwjK48PuV62SczyIU1gC6FabNFiHQOmaO/UcHueKiPohMN2LfL3je7s52K5WE3gUWR3clsFQEjf+0XHZsS11jaU/vwoEMovDM=",
				LogLevel:    "Debug",
				BinVersion:  "TESTING",
				CertData:    CertData{
						"CN=Deathscyth,OU=XXXG-01D,O=OperationMeteor,L=ESUN,ST=SancKingdom,C=EW",
		                 		"CN=Deathscyth,OU=XXXG-01D,O=OperationMeteor,L=ESUN,ST=SancKingdom,C=EW",
						"",
						"SHA256-RSA",
						"RSA",
		                             },
			}
	*/
	configuredRandomURL = defaultConfig.RandomURL
	configuredExchangeURL = defaultConfig.ExchangeURL

	return defaultConfig
}
