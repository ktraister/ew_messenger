package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/encrypt/ecies"
	"go.dedis.ch/kyber/v3/group/edwards25519"
	"net/http"
	"net/url"
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

	decodedBytes, err := base64.StdEncoding.DecodeString(string(b))
	if err != nil {
		fmt.Println("Error decoding base64:", err)
		return b, err
	}

	plainText, err := ecies.Decrypt(suite, cm.localPrivKey, decodedBytes, suite.Hash)
	if err != nil {
		return b, err
	}

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
		conn: conn,
	}

	qPubKey := suite.Point()
	qPubKeyData, err := qPubKey.MarshalBinary()
	if err != nil {
		logger.Error(err)
		return &ConnectionManager{}, err
	}
	localPubKeyStr := base64.StdEncoding.EncodeToString(qPubKeyData)

	//sending the startup message to map user, send computed local pubkey using remote privkey compiled in
	for _, key := range configuration.KyberRemotePubKeys {
		//lets use our connManager plumbing here
		err = qPubKey.UnmarshalBinary([]byte(key))
		if err != nil {
			continue
		}
		connectionManager.remotePubKey = qPubKey

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

	}

	return connectionManager, nil
}
