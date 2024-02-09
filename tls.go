package main

import (
	"crypto/tls"
	"net"
	"net/http"
	"strings"
	"time"
)

/*
I want an interactive button. That will require that connections get bounced

*/

// these two spots are where we'll setup the proxy rewrite (if required)
func tlsClient(input string) *http.Transport {
	ts := &http.Transport{
		TLSClientConfig:   &tls.Config{},
		DisableKeepAlives: true,
		Proxy:             http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   3 * time.Second, // 连接超时
			KeepAlive: 10 * time.Second,
			DualStack: true,
		}).DialContext,
		IdleConnTimeout:       120 * time.Second,
		ResponseHeaderTimeout: 60 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	//okay, were already there, we just have to redirect to localhost if needed
	if strings.Contains(input, "localhost") {
		ts.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
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
