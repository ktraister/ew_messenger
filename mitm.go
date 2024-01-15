package main

import (
	"fyne.io/fyne/v2/widget"
	"github.com/sirupsen/logrus"
)

func mitm(logger *logrus.Logger) {
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

	statusMsgChan <- statusMsg{Target: "MITM", Text: "GO", Import: widget.SuccessImportance, Warn: ""}
}
