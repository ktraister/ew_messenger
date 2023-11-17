package main

import "testing"

func TestFetchConfig(t *testing.T){

    got := fetchConfig()
    want := Configurations{
                RandomURL:   "https://api.endlesswaltz.xyz:443/api",
                ExchangeURL: "wss://exchange.endlesswaltz.xyz:443/ws",
                SSHHost:     "endlesswaltz.xyz",
                LogLevel:    "Error",
        }

    if got != want {
        t.Errorf("got %q, wanted %q", got, want)
    }
}
