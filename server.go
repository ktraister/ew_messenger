package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"go.dedis.ch/kyber/v3/encrypt/ecies"
	"go.dedis.ch/kyber/v3/group/edwards25519"
	"go.dedis.ch/kyber/v3/util/random"
)

type Client_Resp struct {
	UUID string
}

var incomingMsgChan = make(chan Post)
var outgoingMsgChan = make(chan Post)

const charset = "abcdefghijklmnopqrstuvwxyz"

func uid() string {
	rand.Seed(time.Now().UnixNano())
	sb := strings.Builder{}
	//hard-code uid of len 4
	length := 4
	sb.Grow(length)
	for i := 0; i < length; i++ {
		sb.WriteByte(charset[rand.Intn(len(charset))])
	}
	return sb.String()
}

// Change handleConnection to act as the "server side" in this transaction
// we'll pass around the websocket to accomplish this
func handleConnection(dat map[string]interface{}, logger *logrus.Logger, configuration Configurations) {
	//the entire connection will be encrypted using a single kyber key per conn
	suite := edwards25519.NewBlakeSHA256Ed25519()
	//pretty sure I saw another way to generate these keys
	qPrivKey := suite.Scalar().Pick(random.New())
	qPubKey := suite.Point().Mul(qPrivKey, nil)
	qPubKeyData, err := qPubKey.MarshalBinary()
	if err != nil {
		logger.Error(err)
		return
	}

	//debug
	logger.Debug("qPubKey data: '%d'", qPubKeyData)
	logger.Debug(fmt.Printf("Private key: %s\n", qPrivKey))
	logger.Debug(fmt.Printf("Public key: %s\n", qPubKey))

	// Encode byte slice as base64
	qPubKeyStr := base64.StdEncoding.EncodeToString(qPubKeyData)
	localUser := fmt.Sprintf("%s_server-%s", configuration.User, uid())
	targetUser := dat["from"].(string)
	cm, err := exConnect(logger, configuration, localUser)
	if err != nil {
		logger.Error(err)
		return
	}
	logger.Debug("Connected to exchange with user ", localUser)
	logger.Debug("sending base64 pubkey ", qPubKeyStr)

	//we need to respond with a HELO here
	helo := &Message{Type: "helo_reply",
		User: configuration.User,
		From: localUser,
		To:   targetUser,
		Msg:  qPubKeyStr,
	}
	b, err := json.Marshal(helo)
	if err != nil {
		logger.Error(err)
		return
	}

	err = cm.Send(b)
	if err != nil {
		logger.Error("Server:Unable to write message to websocket: ", err)
		return
	}
	logger.Debug("Responded with HELO_REPLY to ", targetUser)

	//receive the encrypted text
	_, incoming, err := cm.Read()
	if err != nil {
		logger.Error("Error reading message:", err)
		return
	}

	err = json.Unmarshal([]byte(incoming), &dat)
	if err != nil {
		logger.Error("Error unmarshalling json:", err)
		return
	}

        logger.Debug("got base64 cipherText ", dat["msg"].(string))
        decodedBytes, err := base64.StdEncoding.DecodeString(dat["msg"].(string))
        if err != nil {                      
                fmt.Println("Error decoding base64:", err)
                return 
        } 
	logger.Debug("decrypting cipherText ", decodedBytes)
	plainText, err := ecies.Decrypt(suite, qPrivKey, decodedBytes, suite.Hash)
	if err != nil {
		logger.Error(fmt.Sprintf("Ciphertext Error: %s", err))
		return
	}

	//logger.Debug("Incoming msg: ", dat["msg"].(string))
	incomingMsg := Post{User: dat["user"].(string), Msg: string(plainText), ok: true}
	incomingMsgChan <- incomingMsg
	playSound()
}
