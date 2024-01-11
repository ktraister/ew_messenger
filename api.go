package main

import (
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"
	"bytes"
)

func removeIndex(s []string, index int) []string {
	return append(s[:index], s[index+1:]...)
}

func getAllUsers(logger *logrus.Logger) ([]string, error) {
	//setup TLS client
	ts := tlsClient(globalConfig.RandomURL)

	urlSlice := strings.Split(globalConfig.ExchangeURL, "/")
	url := "https://" + urlSlice[2] + "/api/userList"
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("User", globalConfig.User)
	req.Header.Set("Passwd", globalConfig.Passwd)
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
		if user != "" && user != globalConfig.User {
			final = append(final, user)
		}
	}

	//sort the list in alphabetical order
	sort.Strings(final)

	return final, nil
}

func getFriends(logger *logrus.Logger) ([]string, error) {
	//setup TLS client
	ts := tlsClient(globalConfig.RandomURL)

	urlSlice := strings.Split(globalConfig.ExchangeURL, "/")
	url := "https://" + urlSlice[2] + "/api/friendsList"
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("User", globalConfig.User)
	req.Header.Set("Passwd", globalConfig.Passwd)
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
		if user != "" && user != globalConfig.User {
			final = append(final, user)
		}
	}

	//sort the list in alphabetical order
	sort.Strings(final)

	return final, nil
}

func putFriends(logger *logrus.Logger) error {
	var final string
	for _, user := range friendUsers {
		user = strings.Replace(user, " ", "", -1)
		if user != "" && user != globalConfig.User {
			final = final + user + ":"
		}
	}

	payload := []byte(`{"UserList":"` + final + `"}`)
	logger.Debug("Putting payload to API: ", `{"UserList":"` + final + `"}`)

	//setup TLS client
	ts := tlsClient(globalConfig.RandomURL)
	urlSlice := strings.Split(globalConfig.ExchangeURL, "/")
	url := "https://" + urlSlice[2] + "/api/updateFriendsList"

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("User", globalConfig.User)
	req.Header.Set("Passwd", globalConfig.Passwd)

	client := http.Client{Timeout: 3 * time.Second, Transport: ts}
	_, err = client.Do(req)
	if err != nil {
		logger.Error(err)
		return err
	}

	return nil
}


func getExUsers(logger *logrus.Logger) ([]string, error) {
	//setup TLS client
	ts := tlsClient(globalConfig.RandomURL)

	urlSlice := strings.Split(globalConfig.ExchangeURL, "/")
	url := "https://" + urlSlice[2] + "/ws/listUsers"
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("User", globalConfig.User)
	req.Header.Set("Passwd", globalConfig.Passwd)
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
		if user != "" && user != globalConfig.User {
			final = append(final, user)
		}
	}

	//sort the list in alphabetical order
	sort.Strings(final)

	return final, nil
}

func getAcctType(logger *logrus.Logger) (string, error) {
	//setup TLS client
	ts := tlsClient(globalConfig.RandomURL)

	urlSlice := strings.Split(globalConfig.ExchangeURL, "/")
	url := "https://" + urlSlice[2] + "/api/premiumCheck"
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("User", globalConfig.User)
	req.Header.Set("Passwd", globalConfig.Passwd)
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
