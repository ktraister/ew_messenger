package main

import (
	"fmt"
	"fyne.io/fyne/v2/widget"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"io"
	"net"
)

var quit = make(chan bool)
var proxyPort int

// if true proxy
func proxyCheck() bool {
	//fmt.Println("returning proxy true")
	return false
}

func proxyFail(pStatus *widget.Label) {
	pStatus.Text = "Proxy Error"
	pStatus.Importance = widget.DangerImportance
	pStatus.Refresh()
	globalConfig.RandomURL = configuredRandomURL
	globalConfig.ExchangeURL = configuredExchangeURL
}

func proxy(configuration Configurations, logger *logrus.Logger, pStatus *widget.Label) {
	logger.Info("Init proxy thread")
	// hard-coding proxy vars, but ingesting creds
	sshServer := configuration.SSHHost
	sshPort := 2222
	sshUser := configuration.User
	sshPassword := configuration.Passwd
	localPort := 0
	remoteAddress := "localhost:443"

	// Create an SSH client configuration
	config := &ssh.ClientConfig{
		User: sshUser,
		Auth: []ssh.AuthMethod{
			ssh.Password(sshPassword),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Insecure; use proper verification in production
	}
	logger.Info("created SSH connection config")

	// Establish SSH connection
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", sshServer, sshPort), config)
	if err != nil {
		logger.Error("Failed to dial:", err)
		proxyFail(pStatus)
		return
	}
	//defer client.Close()
	logger.Info("established SSH connetion")

	// Establish local listener
	localListener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", localPort))
	if err != nil {
		logger.Error("Failed to listen on local port:", err)
		proxyFail(pStatus)
		return
	}
	defer localListener.Close()

	proxyPort = localListener.Addr().(*net.TCPAddr).Port

	logger.Info(fmt.Sprintf("Local port forwarding started on port %d...", proxyPort))
	pStatus.Text = "Proxy Up!"
	pStatus.Importance = widget.HighImportance
	pStatus.Refresh()

	// Accept incoming connections on local port
	for {
		select {
		case <-quit:
			proxyMsgChan <- ""
			return
		default:
			localConn, err := localListener.Accept()
			if err != nil {
				logger.Error("Failed to accept incoming connection:", err)
				proxyFail(pStatus)
				return
			}

			// Connect to the remote address
			remoteConn, err := client.Dial("tcp", remoteAddress)
			if err != nil {
				logger.Error("Failed to dial remote address:", err)
				proxyFail(pStatus)
				return
			}

			// Handle data forwarding in both directions
			go forward(localConn, remoteConn, logger)
			go forward(remoteConn, localConn, logger)
		}
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
