package main

import (
	"embed"
	"log"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

//go:embed audio
var content embed.FS

func playSound() {
	f, err := content.Open("audio/warning_beep.mp3")
	if err != nil {
		log.Error(err)
	}

	streamer, format, err := mp3.Decode(f)
	if err != nil {
		log.Error(err)
	}
	defer streamer.Close()

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	done := make(chan bool)
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		done <- true
	})))

	<-done
}
