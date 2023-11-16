package main

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"net"
)

func main() {
	// Replace these values with your SSH server details
	sshServer := "localhost"
	sshPort := 2222
	sshUser := "zero53"
	sshPassword := "@dm1nAP!"

	// Local port to forward
	localPort := 9090
	// Remote address and port to forward to
	remoteAddress := "endlesswaltz.xyz:443"

	// Create an SSH client configuration
	config := &ssh.ClientConfig{
		User: sshUser,
		Auth: []ssh.AuthMethod{
			ssh.Password(sshPassword),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Insecure; use proper verification in production
	}

	// Establish SSH connection
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", sshServer, sshPort), config)
	if err != nil {
		fmt.Println("Failed to dial:", err)
		return
	}
	defer client.Close()

	// Establish local listener
	localListener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", localPort))
	if err != nil {
		fmt.Println("Failed to listen on local port:", err)
		return
	}
	defer localListener.Close()

	fmt.Printf("Local port forwarding started on port %d...\n", localPort)

	// Accept incoming connections on local port
	for {
		localConn, err := localListener.Accept()
		if err != nil {
			fmt.Println("Failed to accept incoming connection:", err)
			return
		}

		// Connect to the remote address
		remoteConn, err := client.Dial("tcp", remoteAddress)
		if err != nil {
			fmt.Println("Failed to dial remote address:", err)
			return
		}

		// Handle data forwarding in both directions
		go forward(localConn, remoteConn)
		go forward(remoteConn, localConn)
	}
}

// forward copies data from src to dst
func forward(src, dst net.Conn) {
	defer src.Close()
	defer dst.Close()

	_, err := io.Copy(src, dst)
	if err != nil {
		fmt.Println("Error forwarding data:", err)
	}
}
