package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"sync"
)

//start CM
type ConnectionManager struct {
	conn     *websocket.Conn
	mu       sync.Mutex
	isClosed bool
}

func (cm *ConnectionManager) Send(message []byte) error {
	cm.mu.Lock()

	if cm.isClosed {
		return fmt.Errorf("connection is closed")
	}

	err := cm.conn.WriteMessage(websocket.TextMessage, []byte(message))
	if err != nil {
		return err
	}

	cm.mu.Unlock()

	return nil
}

func (cm *ConnectionManager) Read() (int, []byte, error) {
	cm.mu.Lock()

	if cm.isClosed {
		return 0, []byte{}, fmt.Errorf("connection is closed")
	}

	i, b, err := cm.conn.ReadMessage()
	if err != nil {
		return i, b, err
	}

	cm.mu.Unlock()

	return i, b, nil
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
	u, err := url.Parse(configuration.ExchangeURL)
	if err != nil {
		logger.Fatal(err)
		return &ConnectionManager{}, err
	}

	tlsConfig := tlsConfig(configuration.ExchangeURL)
	dialer := websocket.Dialer{
	    TLSClientConfig: tlsConfig,
	}

	// Establish a WebSocket connection
	conn, _, err := dialer.Dial(u.String(), http.Header{"Passwd": []string{configuration.Passwd}, "User": []string{user}})
	if err != nil {
	        logger.Fatal(fmt.Sprintf("Could not establish WebSocket connection with %s: %s", u.String(), err))
		return &ConnectionManager{}, err
	}
	logger.Debug("Connected to exchange server!")

	connectionManager := &ConnectionManager{
		conn: conn,
	}

	//defer close from the get-go
	defer connectionManager.Close()

	//connect to exchange with our username for mapping
	message := &Message{Type: "startup", User: user}
	b, err := json.Marshal(message)
	if err != nil {
		logger.Fatal(err)
		return &ConnectionManager{}, err
	}
	err = connectionManager.Send(b)
	if err != nil {
		logger.Fatal(err)
		return &ConnectionManager{}, err
	}

	return connectionManager, nil
}
