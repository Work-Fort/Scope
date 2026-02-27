package audio

import (
	"bytes"
	"embed"
	"sync"
	"time"

	log "github.com/charmbracelet/log"
	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/speaker"
	"github.com/gopxl/beep/v2/wav"
)

//go:embed sounds/*.wav
var soundFS embed.FS

// Sound identifies a notification sound.
type Sound string

const (
	SoundNone  Sound = "none"
	SoundTone  Sound = "tone"
	SoundOrgan Sound = "organ"
	SoundXylo  Sound = "xylo"
)

// AllSounds returns the available notification sounds in display order.
func AllSounds() []Sound {
	return []Sound{SoundTone, SoundOrgan, SoundXylo, SoundNone}
}

// Label returns a human-readable label for the sound.
func (s Sound) Label() string {
	switch s {
	case SoundTone:
		return "Tone"
	case SoundOrgan:
		return "Organ"
	case SoundXylo:
		return "Xylo"
	case SoundNone:
		return "None"
	default:
		return string(s)
	}
}

var soundFiles = map[Sound]string{
	SoundTone:  "sounds/sound1_tone.wav",
	SoundOrgan: "sounds/sound2_organ.wav",
	SoundXylo:  "sounds/sound3_xylo.wav",
}

var speakerOnce sync.Once

func initSpeaker(rate beep.SampleRate) {
	speakerOnce.Do(func() {
		err := speaker.Init(rate, rate.N(time.Second/10))
		if err != nil {
			log.Error("audio: speaker init failed", "err", err)
		}
	})
}

// Play plays a notification sound asynchronously. Non-blocking.
func Play(s Sound) {
	if s == SoundNone {
		return
	}
	path, ok := soundFiles[s]
	if !ok {
		return
	}

	data, err := soundFS.ReadFile(path)
	if err != nil {
		log.Error("audio: failed to read embedded sound", "sound", s, "err", err)
		return
	}

	streamer, format, err := wav.Decode(bytes.NewReader(data))
	if err != nil {
		log.Error("audio: failed to decode wav", "sound", s, "err", err)
		return
	}

	initSpeaker(format.SampleRate)
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		streamer.Close()
	})))
}
