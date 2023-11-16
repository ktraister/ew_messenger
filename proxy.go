package main

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"net"
	"github.com/sirupsen/logrus"
)

//if true proxy
func proxyCheck() bool {
    //fmt.Println("returning proxy true") 
    return false
}

func proxy(configuration Configurations, logger *logrus.Logger) {
        logger.Info("Init proxy thread")
	// hard-coding proxy vars, but ingesting creds
	sshServer := "localhost"
	sshPort := 2222
	sshUser := configuration.User
	sshPassword := configuration.Passwd
	localPort := 9090
	remoteAddress := "localhost:443"
        logger.Info("read in configuration")

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
		return
	}
	//defer client.Close()
        logger.Info("established SSH connetion")

	// Establish local listener
	localListener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", localPort))
	if err != nil {
		logger.Error("Failed to listen on local port:", err)
		return
	}
	defer localListener.Close()

	logger.Info("Local port forwarding started on port %d...\n", localPort)

	// Accept incoming connections on local port
	for {
		localConn, err := localListener.Accept()
		if err != nil {
			logger.Error("Failed to accept incoming connection:", err)
			return
		}

		// Connect to the remote address
		remoteConn, err := client.Dial("tcp", remoteAddress)
		if err != nil {
			logger.Error("Failed to dial remote address:", err)
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
