package main

import (
	"encoding/base64"
	"fmt"
	"testing"
)

func TestFetchConfig(t *testing.T) {
	output := fetchConfig()
	got := Configurations{
		PrimaryURL:         output.PrimaryURL,
		SSHHost:            output.SSHHost,
		SSHPort:            output.SSHPort,
		SSHKey:             output.SSHKey,
		LogLevel:           output.LogLevel,
		BinVersion:         output.BinVersion,
		KyberRemotePubKeys: output.KyberRemotePubKeys,
		CertData:           output.CertData,
	}

	want := Configurations{
		PrimaryURL: "endlesswaltz.xyz",
		SSHHost:    "ssh.endlesswaltz.xyz",
		SSHPort:    443,
		SSHKey:     "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAEAQDBPsQPG5TOY8sdoanRbdbbr09YD7INldhI6ygw4V4pdVQJ6kS5xWTYyWBkP5s5xtQjQv+ORLqko4h2suGQ/86cbbx66PX8rxHrdMC+AmNeQ19bGBvxt0LBz2FBkHa8O3nlDgwb9/qy8QpTyLrPMu666+2FjIZbfLdHVyDgOgI3u/s2p86d8ggB3/Yt0CfAjCLKQVcxekODR7Nl/eD4Lc+y4MM6YxcQOCrkqWZBlDeQ2PKg8rVVvFN2bh+wq3CwRLMQ4gNNgUIvz0a59klXRlxs8tNwO8aEC6H+DGRU1AVTHLUQPBVK17j1fmKBdwL4hDzxVwdjJ+vUtkF8Vw9Sv706wamYlsW8NccEkDrZonvHJFClxcDgMyEkNdBoaUv+8O8Rd4SOxV1PgIh0g+K+wvkZG1OVp1SRZHkhJ7ozwMmxOHVN24Q6zyv3j3s3f259SpjhAIhojLLlrkeM3GUBzpy5Px24KOZNjb7Ckafpx5gRPQqTnr+kujhYKsGvw0ksUv8MeZlHSdqhwFL6XIqrdddpqfotXEJYSyEnNRobfVKxYW6asdAdq1tTOfNKMtHJoAkmNGkulljw0D7mKJpOzQxmI+aYzhXwbjkEmOCZvGGtj0HShIZBJcnHN4NvwFGLIqGAsf86fzPm6Kw2fy6a4t1i7WD58ngcsnUviMgx1FyJ7y54h4XcIADTN5azDs+DfJgQw3QYE04Xl96PL3LHtsYA84S8+KWGjxypSfdPVFEs5zHtXMgbGuVtyLxNl8O7rzpAK4Ck9dSHhQcPH7gjGrk82e/7mYDqv0Ylf+3tOwkamO2GEkr41Nly0N8TqIQyME6pkJ+hZA8xiYQpdcIVPVoc6W0jTfmG7BVF9cwNZRyPhCtzH427hMClNbtpuipH88LO3NYQD6osTDZoVRP3/nt5eyZueo3otpwpQXwDj9L2Wt8euaa6yYGQ1CHZJr1BYEkc9H8q0kEDoV0qs061g6CkgTB5vmOoNVlahgkgbXc5FiJuQTkMfRtvNwBaIxVWULH95zW8rb1/2sAXujioLxll03tRpc1HNAcCFvUneFqvy1bJ3t5cVoLn5nWaqoBFCebzYdEc+/Yx5qndHMBZz+5BgLLoYWF696xnrQFHpiuGzz5IGM53ox6MPkXK+u4EiEXHifiZ2eH4TCzFA4+5ztsXv3DOv/mYJ0QolVt3t7JugPsu76KxK0ZHqcHbmdZGy3vvwOkbokDqnJTTrppuZzSW+FMUgyvlkJusKXdCdgqE+KQepvBaNRkWKaj75ycxszqqQgQOf2x2IpMQI1wPcEQkbivcT7eNWAzprCG1i57irbE448QKmyImPke3ZvNxIkd4JCBHVAZrB13xU9imuwTz",
		LogLevel:   "Error",
		BinVersion: "REPLACEME",
		KyberRemotePubKeys: [][]byte{
			[]byte{131, 122, 34, 173, 224, 169, 82, 85, 186, 215, 94, 159, 51, 83, 126, 253, 56, 140, 208, 119, 114, 12, 30, 54, 101, 2, 18, 168, 253, 105, 104, 105},
			[]byte{116, 219, 186, 129, 172, 30, 143, 219, 162, 173, 97, 101, 104, 243, 223, 169, 16, 15, 110, 59, 163, 189, 97, 245, 136, 216, 224, 133, 201, 174, 21, 40},
			[]byte{196, 79, 64, 174, 245, 10, 251, 152, 8, 91, 220, 240, 99, 50, 135, 94, 67, 149, 14, 158, 50, 166, 20, 157, 172, 113, 149, 87, 100, 212, 94, 42},
			[]byte{123, 192, 173, 118, 254, 45, 172, 46, 88, 27, 23, 241, 196, 58, 5, 75, 146, 167, 121, 204, 170, 47, 27, 65, 252, 182, 96, 197, 187, 142, 57, 12},
			[]byte{42, 196, 144, 172, 223, 46, 48, 45, 118, 96, 40, 117, 52, 159, 162, 107, 150, 25, 182, 101, 237, 28, 178, 193, 127, 161, 232, 201, 24, 15, 29, 144},
			[]byte{40, 139, 1, 236, 134, 18, 97, 25, 246, 182, 187, 216, 205, 137, 255, 3, 38, 34, 50, 174, 25, 57, 242, 78, 172, 109, 248, 38, 85, 134, 128, 182},
		},
		CertData: CertData{
			"CN=endlesswaltz.xyz",
			"CN=R3,O=Let's Encrypt,C=US",
			"[*.endlesswaltz.xyz endlesswaltz.xyz]",
			"SHA256-RSA",
			"ECDSA",
		},
	}

	if base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%v", got))) != base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%v", want))) {
		t.Errorf("got %q, wanted %q", got, want)
	}
}
