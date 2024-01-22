package main

import (
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/group/edwards25519"
	"go.dedis.ch/kyber/v3/util/random"
)

// Configurations exported
type Configurations struct {
	PrimaryURL         string
	SSHHost            string
	SSHKey             string
	SSHPort            int
	LogLevel           string
	User               string
	Passwd             string
	BinVersion         string
	KyberPubKey        kyber.Point
	KyberPrivKey       kyber.Scalar
	KyberRemotePubKeys []string
	CertData           CertData
}

type CertData struct {
	Subject  string
	Issuer   string
	DNSNames string
	SigAlgo  string
	PubAlgo  string
}

var configuredPrimaryURL = ""

func fetchConfig() Configurations {
	//generate our session kyber keypair
	suite := edwards25519.NewBlakeSHA256Ed25519()
	qPrivKey := suite.Scalar().Pick(random.New())
	qPubKey := suite.Point().Mul(qPrivKey, nil)

	/*
		//create default config
		defaultConfig := Configurations{
			PrimaryURL:  "endlesswaltz.xyz",
			SSHHost:     "ssh.endlesswaltz.xyz",
			SSHPort:     443,
			SSHKey:      "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAEAQDBPsQPG5TOY8sdoanRbdbbr09YD7INldhI6ygw4V4pdVQJ6kS5xWTYyWBkP5s5xtQjQv+ORLqko4h2suGQ/86cbbx66PX8rxHrdMC+AmNeQ19bGBvxt0LBz2FBkHa8O3nlDgwb9/qy8QpTyLrPMu666+2FjIZbfLdHVyDgOgI3u/s2p86d8ggB3/Yt0CfAjCLKQVcxekODR7Nl/eD4Lc+y4MM6YxcQOCrkqWZBlDeQ2PKg8rVVvFN2bh+wq3CwRLMQ4gNNgUIvz0a59klXRlxs8tNwO8aEC6H+DGRU1AVTHLUQPBVK17j1fmKBdwL4hDzxVwdjJ+vUtkF8Vw9Sv706wamYlsW8NccEkDrZonvHJFClxcDgMyEkNdBoaUv+8O8Rd4SOxV1PgIh0g+K+wvkZG1OVp1SRZHkhJ7ozwMmxOHVN24Q6zyv3j3s3f259SpjhAIhojLLlrkeM3GUBzpy5Px24KOZNjb7Ckafpx5gRPQqTnr+kujhYKsGvw0ksUv8MeZlHSdqhwFL6XIqrdddpqfotXEJYSyEnNRobfVKxYW6asdAdq1tTOfNKMtHJoAkmNGkulljw0D7mKJpOzQxmI+aYzhXwbjkEmOCZvGGtj0HShIZBJcnHN4NvwFGLIqGAsf86fzPm6Kw2fy6a4t1i7WD58ngcsnUviMgx1FyJ7y54h4XcIADTN5azDs+DfJgQw3QYE04Xl96PL3LHtsYA84S8+KWGjxypSfdPVFEs5zHtXMgbGuVtyLxNl8O7rzpAK4Ck9dSHhQcPH7gjGrk82e/7mYDqv0Ylf+3tOwkamO2GEkr41Nly0N8TqIQyME6pkJ+hZA8xiYQpdcIVPVoc6W0jTfmG7BVF9cwNZRyPhCtzH427hMClNbtpuipH88LO3NYQD6osTDZoVRP3/nt5eyZueo3otpwpQXwDj9L2Wt8euaa6yYGQ1CHZJr1BYEkc9H8q0kEDoV0qs061g6CkgTB5vmOoNVlahgkgbXc5FiJuQTkMfRtvNwBaIxVWULH95zW8rb1/2sAXujioLxll03tRpc1HNAcCFvUneFqvy1bJ3t5cVoLn5nWaqoBFCebzYdEc+/Yx5qndHMBZz+5BgLLoYWF696xnrQFHpiuGzz5IGM53ox6MPkXK+u4EiEXHifiZ2eH4TCzFA4+5ztsXv3DOv/mYJ0QolVt3t7JugPsu76KxK0ZHqcHbmdZGy3vvwOkbokDqnJTTrppuZzSW+FMUgyvlkJusKXdCdgqE+KQepvBaNRkWKaj75ycxszqqQgQOf2x2IpMQI1wPcEQkbivcT7eNWAzprCG1i57irbE448QKmyImPke3ZvNxIkd4JCBHVAZrB13xU9imuwTz",
			LogLevel:    "Error",
			BinVersion:  "REPLACEME",
			KyberPrivKey: qPrivKey,
			KyberPubKey: qPubKey,
			CertData: CertData{
				"CN=endlesswaltz.xyz",
				"CN=R3,O=Let's Encrypt,C=US",
				"[*.endlesswaltz.xyz endlesswaltz.xyz]",
				"SHA256-RSA",
				"ECDSA",
			},
		}
	*/
	//localdev debug config
	defaultConfig := Configurations{
		PrimaryURL:   "localhost",
		SSHHost:      "localhost",
		SSHPort:      2222,
		SSHKey:       "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDNIUPPef19VF/al9M1fot1+ltlm5eW3HTi7xe/MPmi4NQZTt9DdNZ7wpbTUdyaDRqU4exIjtGpzi8SxZk4uXGX78wfoJenTKyvmTIUXTCecwPkOHGATz1rOGBgXzkmrQSvV7d7gtkqhcfLW0j/kFjKrQYQalGdDGbzt0/KYTIC9FNFBDd6OJWYGWMdecSa9VoomAH1UfaeS5SwIj2K72Pm+KxxSi12ND9ZFXfRy6UB7O9m4oxMBsyBKR/5iU8RiMTlDQx66rRKxob1CdMplMJDv7X7UqxUJGVkc3ec7LNt1FoQPtmEXcRWogQzAeEsZ9g5o84eXV23yYMj5JQYMaodZ6N41nOvcLi/HFETmpe/u/LPseAw9/irRgkNjwDyzUo/gwLubSlSp6B9WaoujUojRM2l1gwxWJqwjK48PuV62SczyIU1gC6FabNFiHQOmaO/UcHueKiPohMN2LfL3je7s52K5WE3gUWR3clsFQEjf+0XHZsS11jaU/vwoEMovDM=",
		LogLevel:     "Debug",
		BinVersion:   "TESTING",
		KyberPrivKey: qPrivKey,
		KyberPubKey:  qPubKey,
		KyberRemotePubKeys: []string{"[9 174 135 218 199 13 39 196 75 245 181 11 34 187 212 235 136 240 87 4 128 235 171 40 174 73 215 52 145 229 221 172]",
			"[243 157 142 144 173 183 208 69 21 214 235 11 7 59 163 157 18 143 249 80 53 88 49 115 178 238 55 8 169 143 53 57]",
			"[207 94 153 73 223 42 170 49 12 204 169 252 151 235 178 158 154 247 153 134 107 93 174 231 153 22 198 125 179 5 150 3]",
		},
		CertData: CertData{
			"CN=Deathscyth,OU=XXXG-01D,O=OperationMeteor,L=ESUN,ST=SancKingdom,C=EW",
			"CN=Deathscyth,OU=XXXG-01D,O=OperationMeteor,L=ESUN,ST=SancKingdom,C=EW",
			"",
			"SHA256-RSA",
			"RSA",
		},
	}
	configuredPrimaryURL = defaultConfig.PrimaryURL

	return defaultConfig
}

/*
Test Private Key -->  9f7b2daf3912f377024de9c303335ac2f1e51dba8e7192c71ad0abd29a646a03
Test Public Key -->  cf5e9949df2aaa310ccca9fc97ebb29e9af799866b5daee79916c67db3059603
Test Public Key Data -->  [207 94 153 73 223 42 170 49 12 204 169 252 151 235 178 158 154 247 153 134 107 93 174 231 153 22 198 125 179 5 150 3]
Test Public Key Data Base64 -->  z16ZSd8qqjEMzKn8l+uynpr3mYZrXa7nmRbGfbMFlgM=
Private Key Set: d4035651fa23eb7097557c87b49d40c788aef9a1b344ab38acbd958e01cc9f0b

Test Private Key -->  1631a424d9ea41c2a1bb9477ec1172450ad93448cc94f496bc7409577968db0a
Test Public Key -->  f39d8e90adb7d04515d6eb0b073ba39d128ff95035583173b2ee3708a98f3539
Test Public Key Data -->  [243 157 142 144 173 183 208 69 21 214 235 11 7 59 163 157 18 143 249 80 53 88 49 115 178 238 55 8 169 143 53 57]
Test Public Key Data Base64 -->  852OkK230EUV1usLBzujnRKP+VA1WDFzsu43CKmPNTk=
Private Key Set: 4e87a72718ac9486e16e727f8d26b72c8d017619317f8afbd729ecc04f88b102

Test Private Key -->  12e893e566d89155399bb8918e51e711732870b3e0178f994c64960a8409b70d
Test Public Key -->  09ae87dac70d27c44bf5b50b22bbd4eb88f0570480ebab28ae49d73491e5ddac
Test Public Key Data -->  [9 174 135 218 199 13 39 196 75 245 181 11 34 187 212 235 136 240 87 4 128 235 171 40 174 73 215 52 145 229 221 172]
Test Public Key Data Base64 -->  Ca6H2scNJ8RL9bULIrvU64jwVwSA66sorknXNJHl3aw=
Private Key Set: bd6e708be31bebf8afb1af9ad972e368155f554fca133651cf3e2f0f4ec1ef06
*/
