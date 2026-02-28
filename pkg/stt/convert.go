package stt

import (
	"encoding/binary"
	"math"
)

// Int16ToFloat32 converts a byte buffer of little-endian int16 PCM samples
// to float32 samples normalized to [-1.0, 1.0] as whisper.cpp expects.
func Int16ToFloat32(buf []byte) []float32 {
	n := len(buf) / 2
	out := make([]float32, n)
	for i := 0; i < n; i++ {
		sample := int16(binary.LittleEndian.Uint16(buf[i*2 : i*2+2]))
		out[i] = float32(sample) / math.MaxInt16
	}
	return out
}
