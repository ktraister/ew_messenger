package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"go.dedis.ch/kyber/v3/encrypt/ecies"
	"go.dedis.ch/kyber/v3/group/edwards25519"
	"strings"
	"time"
)

type Post struct {
	From string `json:"from"`
	To   string `json:"to"`
	Msg  string `json:"msg"`
	Err  error  `json:"ok"`
}

type Message struct {
	Type string `json:"type"`
	User string `json:"user,omitempty"`
	To   string `json:"to,omitempty"`
	From string `json:"from,omitempty"`
	Msg  string `json:"msg,omitempty"`
}

type Random_Req struct {
	Host string `json:"Host"`
	UUID string `json:"UUID"`
}

var dat map[string]interface{}

func ew_client(logger *logrus.Logger, configuration Configurations, message Post) error {
	sysErr := error.New("INTERNAL SYSTEM ERROR")
	user := fmt.Sprintf("%s_client-%s", configuration.User, uid())
	cm, err := exConnect(logger, configuration, user)
	if err != nil {
		return false, sysErr
	}
	defer cm.Close()
	passwd := configuration.Passwd

	targetUser := fmt.Sprintf("%s_%s", string(message.To), "server")

	logger.Debug(fmt.Sprintf("Sending msg %s from user %s to user %s!!", message.Msg, user, targetUser))

	//send HELO to target user
	helo := &Message{Type: "helo",
		User: configuration.User,
		From: user,
		To:   targetUser,
		Msg:  "HELO",
	}
	logger.Debug(helo)
	b, err := json.Marshal(helo)
	if err != nil {
		logger.Error(err)
		return false, sysErr
	}

	err = cm.Send(b)
	if err != nil {
		logger.Error("Client:Unable to write message to websocket: ", err)
		return false, sysErr
	}
	logger.Debug("Client:Sent init HELO")

	//HELO should be received within 5 seconds to proceed OR exit
	cm.conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, incoming, err := cm.Read()
	if err != nil {
		logger.Error("Client:Error reading message:", err)
		return false, sysErr
	}
	logger.Debug("Client:Read init HELO response")

	err = json.Unmarshal([]byte(incoming), &dat)
	logger.Debug(dat)
	if err != nil {
		logger.Error("Client:Error unmarshalling json:", err)
		return false, sysErr
	}

	if dat["msg"].(string) == "User not found" {
		logger.Error("Exchange couldn't route a message to ", targetUser)
		return false, errors.New("User not found")
	} else if dat["msg"].(string) == "Target user limit reached" {
		logger.Info("Exchange throttled target user")
		return false, errors.New("Target user reached message limit. Try again later")
	} else if dat["msg"].(string) == "Basic account limit reached" {
		logger.Info("Exchange throttled basic account")
		return false, errors.New("Message limit reached. Upgrade or wait until Midnight EST to continue.")
	}

	heloUser := strings.Split(dat["from"].(string), "-")[0]
	if dat["type"].(string) == "helo_reply" &&
		heloUser == targetUser {
		logger.Debug("Client received HELO from ", heloUser)
	} else {
		logger.Error(fmt.Sprintf("Didn't receive HELO_REPLY from %s in time, try again later", targetUser))
		return false, errors.New("Target user unable to receive messages.")
	}

	logger.Debug(fmt.Sprintf("shifting remote conn user from %s to %s", targetUser, dat["from"].(string)))
	targetUser = dat["from"].(string)

	//reset conn read deadline
	cm.conn.SetReadDeadline(time.Now().Add(60 * time.Second))

	suite := edwards25519.NewBlakeSHA256Ed25519()
	qPubKey := suite.Point()
	logger.Debug("got base64 pubkey ", dat["msg"].(string))
	decodedBytes, err := base64.StdEncoding.DecodeString(dat["msg"].(string))
	if err != nil {
		fmt.Println("Error decoding base64:", err)
		return false, sysErr
	}
	logger.Debug("qPubKey data: ", decodedBytes)
	err = qPubKey.UnmarshalBinary(decodedBytes)
	if err != nil {
		logger.Error(fmt.Sprintf("PubKey Marshall Error: %d", err))
		return false, sysErr
	}

	logger.Debug("qPubKey before encrypt: ", qPubKey)
	cipherText, err := ecies.Encrypt(suite, qPubKey, []byte(message.Msg), suite.Hash)
	if err != nil {
		logger.Error(fmt.Sprintf("Ciphertext Error: %d", err))
		return false, sysErr
	}

	cipherTextStr := base64.StdEncoding.EncodeToString(cipherText)
	logger.Debug("sending cipherText: ", cipherTextStr)
	//send the ciphertext to the other user throught the websocket
	outgoing := &Message{Type: "cipher",
		User: configuration.User,
		From: user,
		To:   targetUser,
		Msg:  string(cipherTextStr),
	}
	b, err = json.Marshal(outgoing)
	if err != nil {
		logger.Error(err)
		return false, sysErr
	}

	err = cm.Send(b)
	if err != nil {
		logger.Error(err)
		return false, sysErr
	}

	return true, nil
}
