package main

import (
	"embed"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/sirupsen/logrus"
)

//go:embed audio
var content embed.FS

func playSound(logger *logrus.Logger) {
	f, err := content.Open("audio/warning_beep.mp3")
	if err != nil {
		logger.Error(err)
	}

	streamer, format, err := mp3.Decode(f)
	if err != nil {
		logger.Error(err)
	}
	defer streamer.Close()

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	done := make(chan bool)
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		done <- true
	})))

	<-done
}
