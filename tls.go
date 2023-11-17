package main

import (
	"crypto/tls"
	"net/http"
	"strings"
)

/*
I want an interactive button. That will require that connections get bounced

*/

// these two spots are where we'll setup the proxy rewrite (if required)
func tlsClient(input string) *http.Transport {
	ts := &http.Transport{
		TLSClientConfig: &tls.Config{},
	}

	//okay, were already there, we just have to redirect to localhost if needed
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
