package stt

/*
#include <whisper.h>

// Defined in whisper_log.c — suppresses whisper.cpp's stderr output.
extern void whisper_noop_log(enum ggml_log_level level, const char *text, void *user_data);
*/
import "C"

import (
	"fmt"
	"strings"

	"github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"
)

func init() {
	C.whisper_log_set((C.ggml_log_callback)(C.whisper_noop_log), nil)
}

// Transcriber wraps a loaded whisper model for speech-to-text transcription.
// The model is loaded once and reused; a fresh context is created per call
// to avoid state leaks (contexts are not thread-safe).
type Transcriber struct {
	model   whisper.Model
	lang    string
	threads uint
}

// NewTranscriber loads a whisper model from disk and returns a ready transcriber.
func NewTranscriber(modelPath string, lang string, threads uint) (*Transcriber, error) {
	model, err := whisper.New(modelPath)
	if err != nil {
		return nil, fmt.Errorf("stt: load model %s: %w", modelPath, err)
	}
	return &Transcriber{
		model:   model,
		lang:    lang,
		threads: threads,
	}, nil
}

// Transcribe runs speech recognition on the provided float32 audio samples
// (16kHz mono, range [-1.0, 1.0]). Returns the concatenated transcription text.
func (t *Transcriber) Transcribe(samples []float32) (string, error) {
	return t.TranscribeWithCallback(samples, nil)
}

// TranscribeWithCallback runs speech recognition and calls segmentCb for each
// decoded segment in real time. Returns the full concatenated text.
func (t *Transcriber) TranscribeWithCallback(samples []float32, segmentCb func(string)) (string, error) {
	ctx, err := t.model.NewContext()
	if err != nil {
		return "", fmt.Errorf("stt: create context: %w", err)
	}

	if t.lang != "" && ctx.IsMultilingual() {
		if err := ctx.SetLanguage(t.lang); err != nil {
			return "", fmt.Errorf("stt: set language %q: %w", t.lang, err)
		}
	}
	ctx.SetThreads(t.threads)

	var cb whisper.SegmentCallback
	if segmentCb != nil {
		cb = func(seg whisper.Segment) {
			segmentCb(seg.Text)
		}
	}

	if err := ctx.Process(samples, nil, cb, nil); err != nil {
		return "", fmt.Errorf("stt: process audio: %w", err)
	}

	var b strings.Builder
	for {
		seg, err := ctx.NextSegment()
		if err != nil {
			break
		}
		b.WriteString(seg.Text)
	}

	return strings.TrimSpace(b.String()), nil
}

// Close releases the underlying whisper model.
func (t *Transcriber) Close() {
	if t.model != nil {
		t.model.Close()
	}
}
