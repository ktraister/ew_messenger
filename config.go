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
	KyberRemotePubKeys [][]byte
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
		KyberRemotePubKeys: [][]byte{
			[]byte{50, 222, 238, 197, 52, 192, 237, 92, 30, 134, 43, 221, 2, 229, 193, 77, 142, 200, 230, 246, 106, 122, 208, 253, 210, 130, 181, 144, 251, 250, 162, 104},
			[]byte{147, 124, 213, 177, 226, 138, 77, 111, 165, 18, 0, 155, 211, 28, 239, 233, 178, 242, 100, 39, 102, 175, 229, 99, 203, 244, 236, 101, 243, 96, 98, 80},
			[]byte{201, 145, 64, 167, 169, 129, 214, 157, 106, 53, 188, 76, 169, 41, 74, 92, 162, 143, 243, 52, 168, 40, 110, 67, 112, 72, 132, 142, 13, 111, 24, 104},
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
Test Private Key -->  ef27c112d5e0a58f1e8e82062066828087ad6af2a443950fb4d79dd6e884440d
Test Public Key -->  32deeec534c0ed5c1e862bdd02e5c14d8ec8e6f66a7ad0fdd282b590fbfaa268
Test Public Key Data -->  [50 222 238 197 52 192 237 92 30 134 43 221 2 229 193 77 142 200 230 246 106 122 208 253 210 130 181 144 251 250 162 104]
Test Private Key Data -->  [239 39 193 18 213 224 165 143 30 142 130 6 32 102 130 128 135 173 106 242 164 67 149 15 180 215 157 214 232 132 68 13]
Easy Copy Pasta
PubKey -->  []byte{50,222,238,197,52,192,237,92,30,134,43,221,2,229,193,77,142,200,230,246,106,122,208,253,210,130,181,144,251,250,162,104,}
PrivKey -->  []byte{239,39,193,18,213,224,165,143,30,142,130,6,32,102,130,128,135,173,106,242,164,67,149,15,180,215,157,214,232,132,68,13,}

Test Private Key -->  1d43d94b2ba52abe8a36b135f2ef896f10988b46b358e8ddafdd6ab1970df900
Test Public Key -->  937cd5b1e28a4d6fa512009bd31cefe9b2f2642766afe563cbf4ec65f3606250
Test Public Key Data -->  [147 124 213 177 226 138 77 111 165 18 0 155 211 28 239 233 178 242 100 39 102 175 229 99 203 244 236 101 243 96 98 80]
Test Private Key Data -->  [29 67 217 75 43 165 42 190 138 54 177 53 242 239 137 111 16 152 139 70 179 88 232 221 175 221 106 177 151 13 249 0]
Easy Copy Pasta
PubKey -->  []byte{147,124,213,177,226,138,77,111,165,18,0,155,211,28,239,233,178,242,100,39,102,175,229,99,203,244,236,101,243,96,98,80,}
PrivKey -->  []byte{29,67,217,75,43,165,42,190,138,54,177,53,242,239,137,111,16,152,139,70,179,88,232,221,175,221,106,177,151,13,249,0,}

Test Private Key -->  37061cf3655e0096d297f265ec02a726d72bf8d90cee20ae89ec386135b3e70c
Test Public Key -->  c99140a7a981d69d6a35bc4ca9294a5ca28ff334a8286e437048848e0d6f1868
Test Public Key Data -->  [201 145 64 167 169 129 214 157 106 53 188 76 169 41 74 92 162 143 243 52 168 40 110 67 112 72 132 142 13 111 24 104]
Test Private Key Data -->  [55 6 28 243 101 94 0 150 210 151 242 101 236 2 167 38 215 43 248 217 12 238 32 174 137 236 56 97 53 179 231 12]
Easy Copy Pasta
PubKey -->  []byte{201,145,64,167,169,129,214,157,106,53,188,76,169,41,74,92,162,143,243,52,168,40,110,67,112,72,132,142,13,111,24,104,}
PrivKey -->  []byte{55,6,28,243,101,94,0,150,210,151,242,101,236,2,167,38,215,43,248,217,12,238,32,174,137,236,56,97,53,179,231,12,}
*/
