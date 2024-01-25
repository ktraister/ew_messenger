package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/encrypt/ecies"
	"go.dedis.ch/kyber/v3/group/edwards25519"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

var suite = edwards25519.NewBlakeSHA256Ed25519()

// start CM
type ConnectionManager struct {
	conn         *websocket.Conn
	mu           sync.Mutex
	remotePubKey kyber.Point
	localPrivKey kyber.Scalar
	isClosed     bool
}

func (cm *ConnectionManager) Send(message []byte) error {
	cm.mu.Lock()

	if cm.isClosed {
		return fmt.Errorf("connection is closed")
	}

	//encrypt the message
	cipherText, err := ecies.Encrypt(suite, cm.remotePubKey, []byte(message), suite.Hash)
	if err != nil {
		return err
	}

	cipherTextStr := base64.StdEncoding.EncodeToString(cipherText)
	err = cm.conn.WriteMessage(websocket.TextMessage, []byte(cipherTextStr))
	if err != nil {
		return err
	}

	cm.mu.Unlock()

	return nil
}

func (cm *ConnectionManager) Read() ([]byte, error) {
	cm.mu.Lock()

	if cm.isClosed {
		return []byte{}, fmt.Errorf("connection is closed")
	}

	_, b, err := cm.conn.ReadMessage()
	if err != nil {
		return b, err
	}

	inString := fmt.Sprintf("%s", b)

	fmt.Println("Incoming raw read data --> ", inString)

	decodedBytes, err := base64.StdEncoding.DecodeString(inString)
	if err != nil {
		fmt.Println("Error decoding base64:", err)
		return b, err
	}

	fmt.Println("Incoming decoded data --> ", string(decodedBytes))

	plainText, err := ecies.Decrypt(suite, cm.localPrivKey, decodedBytes, suite.Hash)
	if err != nil {
		return b, err
	}

	fmt.Println("Incoming plaingtext data --> ", plainText)

	cm.mu.Unlock()

	return plainText, nil
}

func (cm *ConnectionManager) Close() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if !cm.isClosed {
		cm.conn.Close()
		cm.isClosed = true
	}
}

func exConnect(logger *logrus.Logger, configuration Configurations, user string) (*ConnectionManager, error) {
	// Parse the WebSocket URL
	u, err := url.Parse(fmt.Sprintf("wss://%s:443/ws", configuration.PrimaryURL))
	if err != nil {
		logger.Error(err)
		return &ConnectionManager{}, err
	}

	tlsConfig := tlsConfig(configuration.PrimaryURL)
	dialer := websocket.Dialer{
		TLSClientConfig: tlsConfig,
	}

	// Establish a WebSocket connection
	conn, _, err := dialer.Dial(u.String(), http.Header{"Passwd": []string{configuration.Passwd}, "User": []string{user}})
	if err != nil {
		logger.Error(fmt.Sprintf("Could not establish WebSocket connection with %s: %s", u.String(), err))
		return &ConnectionManager{}, err
	}
	logger.Debug("Connected to exchange server!")

	connectionManager := &ConnectionManager{
		conn:         conn,
		localPrivKey: globalConfig.KyberPrivKey,
	}

	qPubKey := globalConfig.KyberPubKey
	qPubKeyData, err := qPubKey.MarshalBinary()
	if err != nil {
		logger.Error(err)
		return &ConnectionManager{}, err
	}
	localPubKeyStr := base64.StdEncoding.EncodeToString(qPubKeyData)

	publicKeyPoint := suite.Point().Base()
	GO := false
	tries := []int{}
	//sending the startup message to map user, send computed local pubkey using remote privkey compiled in
	for i := 0; i < len(configuration.KyberRemotePubKeys); i++ {
		//choose a configured pubkey at random, equal weights -- try each only once
		var index int
		ok := false
		for !ok {
			tmpIndex, _ := rand.Int(rand.Reader, big.NewInt(int64(len(configuration.KyberRemotePubKeys))))
			if !contains(tries, int(tmpIndex.Int64())) {
				fmt.Println("Picking ", tmpIndex)
				index = int(tmpIndex.Int64())
				tries = append(tries, index)
				ok = true
			}
		}

		//lets use our connManager plumbing here
		err := publicKeyPoint.UnmarshalBinary(configuration.KyberRemotePubKeys[index])
		if err != nil {
			logger.Error("Error setting public key:", err)
			continue
		}

		connectionManager.remotePubKey = publicKeyPoint

		//connect to exchange with our username for mapping
		message := &Message{Type: "startup", User: user, Msg: localPubKeyStr}
		b, err := json.Marshal(message)
		if err != nil {
			logger.Error(err)
			continue
		}

		err = connectionManager.Send(b)
		if err != nil {
			logger.Error(err)
			continue
		}

		//now we receive a mapping reply -- either RESET or OK
		b, err = connectionManager.Read()
		if err != nil {
			return &ConnectionManager{}, err
		}

		logger.Warn(string(b))

		if strings.Contains(string(b), "GO") {
			GO = true
			break
		}
	}

	if !GO {
		err = errors.New("No acceptable Public Key found for exchange")
		return &ConnectionManager{}, err
	}

	return connectionManager, nil
}
