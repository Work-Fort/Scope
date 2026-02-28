package stt

import (
	"fmt"
	"sync"

	"github.com/gen2brain/malgo"
)

const sampleRate = 16000

// Recorder captures audio from the default microphone using malgo (miniaudio).
// Audio is captured at 16kHz, mono, 16-bit signed PCM.
type Recorder struct {
	ctx    *malgo.AllocatedContext
	device *malgo.Device
	mu     sync.Mutex
	buf    []byte
	active bool
}

// NewRecorder initialises a malgo context for audio capture.
func NewRecorder() (*Recorder, error) {
	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, nil)
	if err != nil {
		return nil, fmt.Errorf("stt: init audio context: %w", err)
	}
	return &Recorder{ctx: ctx}, nil
}

// Start begins capturing audio from the default microphone.
func (r *Recorder) Start() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.active {
		return nil
	}

	r.buf = r.buf[:0]

	deviceConfig := malgo.DefaultDeviceConfig(malgo.Capture)
	deviceConfig.Capture.Format = malgo.FormatS16
	deviceConfig.Capture.Channels = 1
	deviceConfig.SampleRate = sampleRate

	callbacks := malgo.DeviceCallbacks{
		Data: func(pOutputSample, pInputSamples []byte, framecount uint32) {
			r.mu.Lock()
			r.buf = append(r.buf, pInputSamples...)
			r.mu.Unlock()
		},
	}

	device, err := malgo.InitDevice(r.ctx.Context, deviceConfig, callbacks)
	if err != nil {
		return fmt.Errorf("stt: init capture device: %w", err)
	}

	if err := device.Start(); err != nil {
		device.Uninit()
		return fmt.Errorf("stt: start capture: %w", err)
	}

	r.device = device
	r.active = true
	return nil
}

// Stop stops capturing and returns the recorded audio as float32 samples.
func (r *Recorder) Stop() []float32 {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.active {
		return nil
	}

	r.device.Uninit()
	r.device = nil
	r.active = false

	return Int16ToFloat32(r.buf)
}

// Snapshot returns the current audio buffer as float32 samples without
// stopping the recording. Used for streaming partial transcription.
func (r *Recorder) Snapshot() []float32 {
	r.mu.Lock()
	snapshot := make([]byte, len(r.buf))
	copy(snapshot, r.buf)
	r.mu.Unlock()

	return Int16ToFloat32(snapshot)
}

// IsActive reports whether the recorder is currently capturing.
func (r *Recorder) IsActive() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.active
}

// Close releases the malgo context.
func (r *Recorder) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.active && r.device != nil {
		r.device.Uninit()
		r.device = nil
		r.active = false
	}
	if r.ctx != nil {
		_ = r.ctx.Uninit()
		r.ctx.Free()
		r.ctx = nil
	}
}
