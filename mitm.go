package main

import (
	"crypto/tls"
	"fmt"
	"fyne.io/fyne/v2/widget"
	"github.com/sirupsen/logrus"
	"time"
)

func mitmStatusCheck(logger *logrus.Logger) {
	logger.Info("Init MITM thread")

	//check account status first
	uType, err := getAcctType(logger)
	if err != nil {
		logger.Error("Failed to check account status:", err)
		return
	}
	if uType != "premium" {
		logger.Info("Turning MITM off based on config")
		statusMsgChan <- statusMsg{Target: "MITM", Text: "STBY", Import: widget.LowImportance, Warn: "MITM disabled for basic users"}
		return
	}

	//we up
	statusMsgChan <- statusMsg{Target: "MITM", Text: "GO", Import: widget.SuccessImportance, Warn: ""}

	ts := tlsConfig(globalConfig.PrimaryURL)
	// Make a TLS connection to the specified domain
	for {
		time.Sleep(10 * time.Second)

		conn, err := tls.Dial("tcp", fmt.Sprintf("%s/%s", globalConfig.PrimaryURL, "api/healthcheck") , ts)
		if err != nil {
			logger.Error("MITM Error dialing:", err)
			statusMsgChan <- statusMsg{Target: "MITM", Text: "WARN", Import: widget.WarningImportance, Warn: "Possible MITM attack detected"}
			continue
		}
		defer conn.Close()

		// Get the peer certificate from the TLS connection
		cert := conn.ConnectionState().PeerCertificates[0]

		if globalConfig.CertData.Subject != fmt.Sprintf("%s", cert.Subject) {
			logger.Warn("MITM switch tripped")
			statusMsgChan <- statusMsg{Target: "MITM", Text: "WARN", Import: widget.WarningImportance, Warn: "Possible MITM attack detected"}
		}

		time.Sleep(10 * time.Second)

		/*
			// Print individual fields
			fmt.Println("Serial Number:", cert.SerialNumber)
			fmt.Println("Subject:", cert.Subject)
			fmt.Println("Issuer:", cert.Issuer)
			fmt.Println("Not Before:", cert.NotBefore.Format(time.RFC3339))
			fmt.Println("Not After:", cert.NotAfter.Format(time.RFC3339))
			fmt.Println("Key Usage:", cert.KeyUsage)
			fmt.Println("Extended Key Usage:", cert.ExtKeyUsage)
			fmt.Println("DNS Names:", cert.DNSNames)
			fmt.Println("IP Addresses:", cert.IPAddresses)
			fmt.Println("Signature Algorithm:", cert.SignatureAlgorithm)
			fmt.Println("Public Key Algorithm:", cert.PublicKeyAlgorithm)
			fmt.Println("Public Key:", cert.PublicKey)
		*/
	}
}
