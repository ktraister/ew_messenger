package main

import (
	"encoding/base64"
	"fmt"
	"fyne.io/fyne/v2/widget"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"io"
	"net"
)

var proxyPort int

// if true proxy
func proxyCheck() bool {
	//fmt.Println("returning proxy true")
	return false
}

func proxyFail() {
	globalConfig.PrimaryURL = configuredPrimaryURL
	statusMsgChan <- statusMsg{Target: "PROXY", Text: "WARN", Import: widget.WarningImportance, Warn: "Proxy operation failed. \nTraffic will be routed normally."}
}

// create human-readable SSH-key strings
func keyString(k ssh.PublicKey) string {
	return k.Type() + " " + base64.StdEncoding.EncodeToString(k.Marshal()) // e.g. "ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTY...."
}

func trustedHostKeyCallback(logger *logrus.Logger, trustedKey string) ssh.HostKeyCallback {
	return func(_ string, _ net.Addr, k ssh.PublicKey) error {
		ks := keyString(k)
		if trustedKey != ks {
			logger.Error("SSH-key verification FAILED for key: ", keyString(k))
			proxyFail()
			return fmt.Errorf("SSH-key verification: expected %q but got %q", trustedKey, ks)
		}
		return nil
	}
}

func proxy(logger *logrus.Logger) {
	logger.Info("Init proxy thread")

	//check account status first
	uType, err := getAcctType(logger)
	if err != nil {
		logger.Error("Failed to check account status:", err)
		proxyFail()
		return
	}
	logger.Debug("from the API for user acct type: ", uType)
	if uType != "premium" {
		logger.Info("Turning proxy off based on config")
		statusMsgChan <- statusMsg{Target: "PROXY", Text: "STBY", Import: widget.LowImportance, Warn: "Proxy disabled for basic users"}
		return
	}

	// hard-coding proxy vars, but ingesting creds
	sshServer := globalConfig.SSHHost
	sshPort := 443
	sshUser := globalConfig.User
	sshPassword := globalConfig.Passwd
	localPort := 0
	remoteAddress := "localhost:443"

	//specifies global configuration values for SSH algos -- SHOULD BE SAME IN PROXY.GO
	cipherConfig := ssh.Config{
		KeyExchanges: []string{"curve25519-sha256", "curve25519-sha256@libssh.org"},
		Ciphers:      []string{"aes128-gcm@openssh.com", "aes256-gcm@openssh.com", "aes128-ctr", "aes192-ctr", "aes256-ctr"},
		MACs:         []string{"hmac-sha2-256-etm@openssh.com", "hmac-sha2-512-etm@openssh.com", "hmac-sha2-256", "hmac-sha2-512"},
	}

	// Create an SSH client configuration
	config := &ssh.ClientConfig{
		Config: cipherConfig,
		User:   sshUser,
		Auth: []ssh.AuthMethod{
			ssh.Password(sshPassword),
		},
		HostKeyCallback: trustedHostKeyCallback(logger, globalConfig.SSHKey),
	}
	logger.Info("created SSH connection config")

	// Establish SSH connection
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", sshServer, sshPort), config)
	if err != nil {
		logger.Error("Failed to dial:", err)
		proxyFail()
		return
	}
	//defer client.Close()
	logger.Info("established SSH connetion")

	// Establish local listener
	localListener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", localPort))
	if err != nil {
		logger.Error("Failed to listen on local port:", err)
		proxyFail()
		return
	}
	defer localListener.Close()

	proxyPort = localListener.Addr().(*net.TCPAddr).Port
	globalConfig.PrimaryURL = fmt.Sprintf("localhost:%d", proxyPort)

	logger.Info(fmt.Sprintf("Local port forwarding started on port %d...", proxyPort))
	statusMsgChan <- statusMsg{Target: "PROXY", Text: "GO", Import: widget.SuccessImportance, Warn: ""}

	// Accept incoming connections on local port
	for {
		localConn, err := localListener.Accept()
		if err != nil {
			logger.Error("Failed to accept incoming connection:", err)
			proxyFail()
			return
		}

		// Connect to the remote address
		remoteConn, err := client.Dial("tcp", remoteAddress)
		if err != nil {
			logger.Error("Failed to dial remote address:", err)
			proxyFail()
			return
		}

		// Handle data forwarding in both directions
		go forward(localConn, remoteConn, logger)
		go forward(remoteConn, localConn, logger)
	}
}

// forward copies data from src to dst
func forward(src, dst net.Conn, logger *logrus.Logger) {
	defer src.Close()
	defer dst.Close()

	_, err := io.Copy(src, dst)
	if err != nil {
		logger.Error("Error forwarding data:", err)
	}
}
