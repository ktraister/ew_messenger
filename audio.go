package main

import (
	"embed"
	"time"
        "fmt"
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/effects"
	"github.com/sirupsen/logrus"
)

//go:embed audio
var content embed.FS
var volume = float64(-5)
var selectedSound = "warning_beep"

func playSound(logger *logrus.Logger) {
	f, err := content.Open(fmt.Sprintf("audio/%s.mp3", selectedSound))
	if err != nil {
		logger.Error(err)
	}

	streamer, format, err := mp3.Decode(f)
	if err != nil {
		logger.Error(err)
	}
	defer streamer.Close()

	//example -> https://github.com/faiface/beep/blob/v1.1.0/examples/tutorial/2-composing-and-controlling/d/main.go
	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	ctrl := &beep.Ctrl{Streamer: beep.Loop(1, streamer), Paused: false}
	Volume := &effects.Volume{
		Streamer: ctrl,
		Base:     2,
		Volume:   0,
		Silent:   false,
	}
	speedy := beep.ResampleRatio(4, 1, Volume)
	speaker.Lock()
	Volume.Volume = volume
	speaker.Unlock()
	speaker.Play(speedy)
}
