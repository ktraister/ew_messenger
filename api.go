package main

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"fyne.io/fyne/v2/widget"
	"github.com/sirupsen/logrus"
	"go.dedis.ch/kyber/v3/encrypt/ecies"
	"io"
	"math/big"
	"net/http"
	"sort"
	"strings"
	"time"
)

func removeIndex(s []string, index int) []string {
	return append(s[:index], s[index+1:]...)
}

func buildAuthHeader() (string, error) {
	i, _ := rand.Int(rand.Reader, big.NewInt(int64(len(globalConfig.KyberRemotePubKeys))))
	index := int(i.Int64())
	remotePubKey := suite.Point()
	err := remotePubKey.UnmarshalBinary(globalConfig.KyberRemotePubKeys[index])
	if err != nil {
		return "", err
	}

	payload := fmt.Sprintf("%s:%s", globalConfig.User, globalConfig.Passwd)
	//encrypt the message
	cipherText, err := ecies.Encrypt(suite, remotePubKey, []byte(payload), suite.Hash)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(cipherText), nil
}

func checkCreds() (bool, string) {
	//setup tls
	ts := tlsClient(globalConfig.PrimaryURL)

	authHeader, err := buildAuthHeader()
	if err != nil {
		return false, "Unable to encrypt credentials for transit"
	}

	//check and make sure inserted creds
	//Random and Exchange will use same mongo, so the creds will be valid for both
	health_url := fmt.Sprintf("https://%s:443/%s", globalConfig.PrimaryURL, "api/healthcheck")
	req, err := http.NewRequest("GET", health_url, nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("Auth", authHeader)
	client := http.Client{Timeout: 10 * time.Second, Transport: ts}
	resp, err := client.Do(req)
	errorText := ""
	if err != nil {
		errorText = "Couldn't Connect to RandomAPI"
		return false, errorText
	}
	if resp == nil {
		errorText = "No Response From RandomAPI"
		return false, errorText
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		errorText = "Invalid Username/Password Combination"
		return false, errorText
	}
	return true, ""
}

func getAllUsers(logger *logrus.Logger) ([]string, error) {
	//setup TLS client
	ts := tlsClient(globalConfig.PrimaryURL)

	authHeader, err := buildAuthHeader()
	if err != nil {
		return []string{}, fmt.Errorf("Unable to encrypt credentials for transit")
	}

	url := fmt.Sprintf("https://%s/%s", globalConfig.PrimaryURL, "api/userList")
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("Auth", authHeader)
	client := http.Client{Timeout: 10 * time.Second, Transport: ts}
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
	ts := tlsClient(globalConfig.PrimaryURL)

	authHeader, err := buildAuthHeader()
	if err != nil {
		return []string{}, fmt.Errorf("Unable to encrypt credentials for transit")
	}

	url := fmt.Sprintf("https://%s/%s", globalConfig.PrimaryURL, "api/friendsList")
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("Auth", authHeader)
	client := http.Client{Timeout: 10 * time.Second, Transport: ts}
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
	for _, user := range tmpFriendUsers {
		user = strings.Replace(user, " ", "", -1)
		if user != "" && user != globalConfig.User {
			final = final + user + ":"
		}
	}

	payload := []byte(final)
	logger.Debug("Putting payload to API: ", final)

	//setup TLS client
	ts := tlsClient(globalConfig.PrimaryURL)

	authHeader, err := buildAuthHeader()
	if err != nil {
		return fmt.Errorf("Unable to encrypt credentials for transit")
	}

	url := fmt.Sprintf("https://%s/%s", globalConfig.PrimaryURL, "api/updateFriendsList")
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("Auth", authHeader)

	client := http.Client{Timeout: 10 * time.Second, Transport: ts}
	_, err = client.Do(req)
	if err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

func getExUsers(logger *logrus.Logger) ([]string, error) {
	//setup TLS client
	ts := tlsClient(globalConfig.PrimaryURL)

	authHeader, err := buildAuthHeader()
	if err != nil {
		return []string{}, fmt.Errorf("Unable to encrypt credentials for transit")
	}

	url := fmt.Sprintf("https://%s/%s", globalConfig.PrimaryURL, "ws/listUsers")
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("Auth", authHeader)
	client := http.Client{Timeout: 10 * time.Second, Transport: ts}
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
	ts := tlsClient(globalConfig.PrimaryURL)

	authHeader, err := buildAuthHeader()
	if err != nil {
		return "woops", fmt.Errorf("Unable to encrypt credentials for transit")
	}

	url := fmt.Sprintf("https://%s:443/%s", globalConfig.PrimaryURL, "api/premiumCheck")
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("Auth", authHeader)
	client := http.Client{Timeout: 10 * time.Second, Transport: ts}
	var final string
	for i := 0; i <= 3; i++ {
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
		final = string(output)
		if final == "premium" || final == "basic" {
			break
		}
		time.Sleep(1 * time.Second)
	}

	return final, nil
}

func binIsCurrent(logger *logrus.Logger) bool {
	//cop out
	if globalConfig.BinVersion == "TESTING" {
		return true
	}

	//setup TLS client
	ts := tlsClient(globalConfig.PrimaryURL)

	authHeader, err := buildAuthHeader()
	if err != nil {
		logger.Error("Unable to encrypt credentials for transit")
		return false
	}

	url := fmt.Sprintf("https://%s:443/%s", globalConfig.PrimaryURL, "api/clientVersionCheck")
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("Auth", authHeader)
	client := http.Client{Timeout: 10 * time.Second, Transport: ts}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error(err)
		return false
	}

	output, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	if err != nil {
		logger.Error(err)
		return false
	}

	return globalConfig.BinVersion == string(output)
}

func apiStatusCheck(logger *logrus.Logger) {
	//setup tls
	ts := tlsClient(globalConfig.PrimaryURL)

	authHeader, err := buildAuthHeader()
	if err != nil {
		logger.Error("Unable to encrypt creds in apiStatusCheck")
		return
	}

	//check and make sure inserted creds
	//Random and Exchange will use same mongo, so the creds will be valid for both
	health_url := fmt.Sprintf("https://%s:443/%s", globalConfig.PrimaryURL, "/api/healthcheck")
	req, _ := http.NewRequest("GET", health_url, nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("Auth", authHeader)
	client := http.Client{Timeout: 10 * time.Second, Transport: ts}
	healthy := true

	for {
		time.Sleep(10 * time.Second)
		resp, err := client.Do(req)
		if err != nil {
			logger.Error("apiStatusCheck Couldn't Connect to RandomAPI")
			statusMsgChan <- statusMsg{Target: "API", Text: "ERR", Import: widget.DangerImportance, Warn: "Unable to connect to API"}
			healthy = false
			continue
		}
		if resp == nil {
			logger.Error("apiStatusCheck got no response")
			statusMsgChan <- statusMsg{Target: "API", Text: "ERR", Import: widget.DangerImportance, Warn: "No response from API"}
			healthy = false
			continue
		}

		output, err := io.ReadAll(resp.Body)
		if err != nil {
			logger.Error("apiStatusCheck error reading data from RandomAPI")
			statusMsgChan <- statusMsg{Target: "API", Text: "ERR", Import: widget.DangerImportance, Warn: "Unable to read API data"}
			healthy = false
			continue
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			logger.Error("apiStatusCheck Request failed with status: ", resp.Status)
			statusMsgChan <- statusMsg{Target: "API", Text: "ERR", Import: widget.DangerImportance, Warn: "API returned bad status code"}
			healthy = false
			continue
		}

		if string(output) != "HEALTHY" {
			logger.Warn("apiStatusCheck Request did not get programmed response")
			statusMsgChan <- statusMsg{Target: "API", Text: "WARN", Import: widget.WarningImportance, Warn: "API returned bad check value"}
			healthy = false
			continue
		} else {
			if healthy == false {
				statusMsgChan <- statusMsg{Target: "API", Text: "GO", Import: widget.SuccessImportance, Warn: "API recovered"}
				healthy = true
			} else {
				statusMsgChan <- statusMsg{Target: "API", Text: "GO", Import: widget.SuccessImportance, Warn: ""}
			}
		}
	}
}

func exStatusCheck(logger *logrus.Logger) {
	//setup tls
	ts := tlsClient(globalConfig.PrimaryURL)

	authHeader, err := buildAuthHeader()
	if err != nil {
		logger.Error("Unable to encrypt credentials for transit")
		return
	}

	//check and make sure inserted creds
	//Random and EXchange will use same mongo, so the creds will be valid for both
	health_url := fmt.Sprintf("https://%s:443/%s", globalConfig.PrimaryURL, "ws/healthcheck")
	req, _ := http.NewRequest("GET", health_url, nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("Auth", authHeader)
	client := http.Client{Timeout: 10 * time.Second, Transport: ts}
	healthy := true

	for {
		time.Sleep(10 * time.Second)
		resp, err := client.Do(req)
		if err != nil {
			logger.Error("exStatusCheck Couldn't Connect to Exchange")
			statusMsgChan <- statusMsg{Target: "EX", Text: "ERR", Import: widget.DangerImportance, Warn: "Unable to connect to Exchange"}
			healthy = false
			continue
		}
		if resp == nil {
			logger.Error("exStatusCheck got no response")
			statusMsgChan <- statusMsg{Target: "EX", Text: "ERR", Import: widget.DangerImportance, Warn: "No response from Exchange"}
			healthy = false
			continue
		}

		output, err := io.ReadAll(resp.Body)
		if err != nil {
			logger.Error("exStatusCheck error reading data from Exchange")
			statusMsgChan <- statusMsg{Target: "EX", Text: "ERR", Import: widget.DangerImportance, Warn: "Unable to read Exchange data"}
			healthy = false
			continue
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			logger.Error("exStatusCheck Request failed with status: ", resp.Status)
			statusMsgChan <- statusMsg{Target: "EX", Text: "ERR", Import: widget.DangerImportance, Warn: "Exchange returned bad status code"}
			healthy = false
			continue
		}

		if string(output) != "HEALTHY" {
			logger.Warn("exStatusCheck Request did not get programmed response")
			statusMsgChan <- statusMsg{Target: "EX", Text: "WARN", Import: widget.WarningImportance, Warn: "Exchange returned bad check value"}
			healthy = false
			continue
		} else {
			if healthy == false {
				statusMsgChan <- statusMsg{Target: "EX", Text: "GO", Import: widget.SuccessImportance, Warn: "Exchange recovered"}
				healthy = true
			} else {
				statusMsgChan <- statusMsg{Target: "EX", Text: "GO", Import: widget.SuccessImportance, Warn: ""}
			}
		}
	}
}
