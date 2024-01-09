package main

import (
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"
)

func removeIndex(s []string, index int) []string {
	return append(s[:index], s[index+1:]...)
}

func getExUsers(logger *logrus.Logger, configuration Configurations) ([]string, error) {
	//setup TLS client
	ts := tlsClient(configuration.RandomURL)

	urlSlice := strings.Split(configuration.ExchangeURL, "/")
	url := "https://" + urlSlice[2] + "/ws/listUsers"
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("User", configuration.User)
	req.Header.Set("Passwd", configuration.Passwd)
	client := http.Client{Timeout: 3 * time.Second, Transport: ts}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error(err)
		return []string{}, err
	}
	output, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error(err)
		return []string{}, err
	}

	tmpUsers := strings.Split(string(output), ":")
	final := []string{}
	for _, user := range tmpUsers {
		user = strings.Replace(user, " ", "", -1)
		if user != "" && user != configuration.User {
			final = append(final, user)
		}
	}

	//sort the list in alphabetical order
	sort.Strings(final)

	return final, nil
}

func getAcctType(logger *logrus.Logger, configuration Configurations) (string, error) {
	//setup TLS client
	ts := tlsClient(configuration.RandomURL)

	urlSlice := strings.Split(configuration.ExchangeURL, "/")
	url := "https://" + urlSlice[2] + "/api/premiumCheck"
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("User", configuration.User)
	req.Header.Set("Passwd", configuration.Passwd)
	client := http.Client{Timeout: 3 * time.Second, Transport: ts}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error(err)
		return "woops", err
	}
	output, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error(err)
		return "woops", err
	}

	return string(output), nil
}

func binIsCurrent(logger *logrus.Logger) bool {
	//cop out
	if globalConfig.BinVersion == "TESTING" {
		return true
	}

	//setup TLS client
	ts := tlsClient(globalConfig.RandomURL)

	urlSlice := strings.Split(globalConfig.ExchangeURL, "/")
	url := "https://" + urlSlice[2] + "/api/clientVersionCheck"
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("User", globalConfig.User)
	req.Header.Set("Passwd", globalConfig.Passwd)
	client := http.Client{Timeout: 3 * time.Second, Transport: ts}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error(err)
		return false
	}
	output, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error(err)
		return false
	}

	return globalConfig.BinVersion == string(output)
}
