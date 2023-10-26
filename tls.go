package main

import (
	"crypto/tls"
	"net/http"
	"strings"
)

func tlsClient(input string) *http.Transport {
	ts := &http.Transport{
		TLSClientConfig: &tls.Config{},
	}

	if strings.Contains(input, "localhost") {
		ts = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	return ts
}

func tlsConfig(input string) *tls.Config {
	ts := &tls.Config{}

	if strings.Contains(input, "localhost") {
		ts = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	return ts
}
