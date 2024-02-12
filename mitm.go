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

	// Make a TLS connection to the specified domain
	for {
		time.Sleep(10 * time.Second)

		ts := tlsConfig(globalConfig.PrimaryURL)
		conn, err := tls.Dial("tcp", globalConfig.PrimaryURL, ts)
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
	}
}
