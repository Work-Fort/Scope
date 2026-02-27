# Whisper Speech-to-Text — CGO In-Process Design

Wire the mic button in the chat input bar to record audio via malgo, transcribe in-process via whisper.cpp CGO bindings, and insert the resulting text into the message input.

## Problem

The chat UI has a mic button that renders but does nothing. Voice input is a key usability feature for a chat app.

## Approach

Use **whisper.cpp official Go bindings** (MIT, `github.com/ggerganov/whisper.cpp/bindings/go`) for in-process transcription and **malgo** (Unlicense, `github.com/gen2brain/malgo` v0.11.24) for microphone capture. malgo vendors miniaudio as a single C file with no system deps. The whisper.cpp bindings require a cmake pre-build step to produce static libraries.

This avoids managing a separate subprocess, eliminates HTTP round-trip latency, and keeps all transcription in-process.

### License

All dependencies are GPL-3.0 compatible:

| Library | License | Compatible |
|---------|---------|------------|
| whisper.cpp + Go bindings | MIT | Yes |
| malgo (miniaudio) | Unlicense | Yes |
| GGML model weights | MIT | Yes |

The project will relicense from GPL-2.0-Only to **GPL-3.0-or-later** before merging this feature to accommodate future dependencies. MIT and Unlicense are compatible with both GPL-2.0 and GPL-3.0.

## Architecture

```
┌──────────────────────────────────────────────────────────────┐
│  Bubble Tea (ChatModel)                                      │
│                                                              │
│  ┌──────────┐   StartRecordingCmd   ┌──────────────────┐    │
│  │ Mic Btn  │ ────────────────────> │ malgo Recorder   │    │
│  │ Ctrl+M   │                       │ (goroutine)      │    │
│  └──────────┘                       │ 16kHz/mono/s16   │    │
│       │                             └───────┬──────────┘    │
│       │                                     │ audio chunks  │
│       │                                     v               │
│       │                             ┌──────────────────┐    │
│       │  StreamSegmentMsg (partial)  │ whisper.cpp      │    │
│       │ <─────────────────────────── │ (streaming CGO)  │    │
│       │  StreamDoneMsg (final)       │ segment callback │    │
│       │ <─────────────────────────── └──────────────────┘    │
│       v                                                      │
│  ┌──────────┐  input LOCKED during recording/transcribing    │
│  │ InputBar │  partial text replaces previous partial text   │
│  └──────────┘  final text stays, input unlocked              │
└──────────────────────────────────────────────────────────────┘
```

**Streaming model:** Audio is captured continuously while recording. Every ~2-3 seconds, the accumulated audio buffer is sent to whisper for transcription. Partial results replace previous partial text in the input bar in real-time. When recording stops, a final transcription pass runs on the complete audio, and the input bar is unlocked for editing.

**Input locking:** While recording or transcribing, keyboard input to the message box is **blocked**. Only the mic toggle (click or Ctrl+M) is accepted. This prevents the user from typing while partial transcription text is being updated.

All I/O happens in `tea.Cmd` goroutines. CGO calls block the calling goroutine's OS thread but the Go runtime spawns additional threads, so the Bubble Tea event loop is unaffected.

## New Dependencies

```
github.com/gen2brain/malgo            # miniaudio Go bindings (Unlicense)
github.com/ggerganov/whisper.cpp/bindings/go  # whisper.cpp CGO (MIT)
```

## Audio Capture — `pkg/stt/recorder.go`

### Verified malgo API (v0.11.24)

```go
// malgo types (verified from source):
// malgo.DefaultDeviceConfig(deviceType DeviceType) DeviceConfig
// malgo.Capture = DeviceType(2)
// malgo.FormatS16 = FormatType(2)
// malgo.InitContext(backends []Backend, config ContextConfig, logProc func(string)) (*AllocatedContext, error)
// malgo.InitDevice(context uintptr, config DeviceConfig, callbacks DeviceCallbacks) (*Device, error)
//
// DataProc callback (NOTE: output param comes FIRST, input SECOND):
//   type DataProc func(pOutputSample, pInputSamples []byte, framecount uint32)
//   For capture, audio arrives in pInputSamples (2nd param)

deviceConfig := malgo.DefaultDeviceConfig(malgo.Capture)
deviceConfig.Capture.Format   = malgo.FormatS16  // 16-bit signed PCM
deviceConfig.Capture.Channels = 1                 // mono
deviceConfig.SampleRate       = 16000             // whisper expects 16kHz
```

**System deps on Linux:** `-ldl -lpthread -lm` only (automatic via `#cgo` directives). No ALSA/PulseAudio dev headers needed — miniaudio discovers backends at runtime via dlopen.

### Recorder API

```go
package stt

type Recorder struct {
    ctx     *malgo.AllocatedContext
    device  *malgo.Device
    mu      sync.Mutex
    buf     []byte
    active  bool
}

func NewRecorder() (*Recorder, error)
func (r *Recorder) Start() error        // begin capture
func (r *Recorder) Stop() []float32     // stop capture, return all samples as float32
func (r *Recorder) Snapshot() []float32  // return current buffer copy without stopping
func (r *Recorder) IsActive() bool
func (r *Recorder) Close()              // release malgo context
```

`Stop()` and `Snapshot()` convert the accumulated `[]byte` (int16 LE) buffer to `[]float32` in the range `[-1.0, 1.0]` as whisper expects. `Snapshot()` is used for streaming — it returns the current audio buffer for partial transcription without interrupting capture.

### Max Duration

Hard cap at **30 seconds** (whisper's native window). A `time.AfterFunc` auto-stops recording at 30s and sends `RecordingDoneMsg`.

## Transcription — `pkg/stt/transcriber.go`

### Verified Whisper Go API (`github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper`)

```go
// Verified interfaces from source (bindings/go/pkg/whisper/interface.go):
//
// whisper.New(path string) (Model, error)
//
// type Model interface {
//     io.Closer
//     NewContext() (Context, error)
//     IsMultilingual() bool
//     Languages() []string
// }
//
// type Context interface {
//     SetLanguage(string) error
//     SetThreads(uint)
//     Process([]float32, EncoderBeginCallback, SegmentCallback, ProgressCallback) error
//     NextSegment() (Segment, error)  // returns io.EOF when done
//     SetVAD(bool)
//     SetVADThreshold(float32)
//     // ... many more setters
// }
//
// type Segment struct {
//     Num        int
//     Start, End time.Duration
//     Text       string
//     Tokens     []Token
// }
//
// Callback types:
//   type SegmentCallback func(Segment)       // fires per segment during Process
//   type ProgressCallback func(int)           // fires with % progress
//   type EncoderBeginCallback func() bool     // return false to abort
//
// Constants:
//   whisper.SampleRate = 16000
//   whisper.SampleBits = 32

package stt

type Transcriber struct {
    model   whisper.Model
    lang    string
    threads uint
}

func NewTranscriber(modelPath string, lang string) (*Transcriber, error)
func (t *Transcriber) Transcribe(samples []float32) (string, error)
func (t *Transcriber) Close()
```

`Transcribe` creates a context from the loaded model, sets language and threads, calls `ctx.Process(samples, nil, segmentCb, nil)`, and concatenates all segments. The `SegmentCallback` fires for each segment during processing — this is the hook for streaming partial results.

### Streaming Transcription

Two complementary mechanisms provide streaming feedback:

**1. SegmentCallback (within a single Process call):**
`Process()` accepts a `SegmentCallback func(Segment)` that fires for each segment as whisper decodes. This provides real-time text within a single transcription pass.

**2. Tick-based re-transcription (across growing audio buffer):**
While recording continues, a tick every ~2.5s snapshots the growing audio buffer and runs a fresh `Process()` call. Each call transcribes all audio accumulated so far, producing progressively better results as more speech is captured.

The combined flow:

1. A `streamTickMsg` fires every ~2.5 seconds while `recording == RecordingActive`
2. On each tick, `recorder.Snapshot()` grabs the current audio buffer
3. `transcriber.Transcribe(snapshot)` runs in a `tea.Cmd` goroutine
4. Returns `StreamSegmentMsg{Text: partial, Final: false}`
5. The input bar replaces its STT content with the partial text
6. If a tick fires while a previous transcription is still in-flight, the tick is skipped

When the user stops recording:
1. `recorder.Stop()` returns the complete audio
2. One final `transcriber.Transcribe(allSamples)` produces the definitive text
3. Returns `StreamSegmentMsg{Text: "final text", Final: true}`
4. Input bar shows final text and **unlocks** for editing

The input bar tracks `sttStart int` — the cursor position where STT text begins. Partial updates replace only text from `sttStart` onward, preserving any text typed before recording started.

### Thread Safety

whisper.cpp contexts are **not thread-safe**, but the model is shared read-only. The `Transcriber.Transcribe` method creates a fresh context per call to avoid state leaks. During streaming, only one transcription runs at a time — if a tick fires while a previous transcription is still running, the tick is skipped.

### Lazy Loading

The model is loaded on **first mic press**, not at startup. This avoids the ~125 MB memory hit for users who never use voice. The `Transcriber` is created lazily in the ChatModel.

## Model Management — `pkg/stt/model.go`

### Storage

Models stored in `$XDG_DATA_HOME/workfort/models/` (typically `~/.local/share/workfort/models/`).

### Available Models

| Model | Disk | RAM | Quality | Default |
|-------|------|-----|---------|---------|
| ggml-tiny.en.bin | 75 MB | ~125 MB | Good for English | **Yes** |
| ggml-base.en.bin | 142 MB | ~388 MB | Better | No |
| ggml-small.en.bin | 466 MB | ~743 MB | Best practical | No |

### Download

```go
func EnsureModel(name string, progressFn func(float64)) (string, error)
```

Downloads from `https://huggingface.co/ggerganov/whisper.cpp/resolve/main/<name>` if not cached. Returns the local file path. The `progressFn` callback reports download progress (0.0 → 1.0) for UI feedback.

### Config Keys

Add to `pkg/config/viper.go`:

```go
viper.SetDefault("stt-model", "ggml-tiny.en.bin")
viper.SetDefault("stt-language", "en")     // "" for auto-detect
viper.SetDefault("stt-threads", 4)
```

## TUI Integration — `internal/chat/model.go`

### Recording State

```go
type RecordingState int

const (
    RecordingIdle RecordingState = iota
    RecordingActive
    RecordingTranscribing
)
```

Add to `ChatModel`:

```go
recording     RecordingState
recorder      *stt.Recorder
transcriber   *stt.Transcriber   // lazy-loaded
```

### Tea Messages

```go
type RecordingStartedMsg struct{}
type RecordingDoneMsg struct{ Samples []float32 }
type StreamSegmentMsg struct{ Text string; Final bool }
type streamTickMsg struct{}
type TranscribeErrorMsg struct{ Err error }
type ModelDownloadProgressMsg struct{ Pct float64 }
type ModelReadyMsg struct{ Path string }
type ModelDownloadErrorMsg struct{ Err error }
```

### Input Locking

**While `recording != RecordingIdle`, keyboard input to the message box is blocked.** Only these actions are allowed:
- Mic toggle (click or Ctrl+M) — to stop recording
- Ctrl+Q — to quit

All other key presses and paste events are silently dropped. The input bar's `Blur()` is NOT called (we want the cursor visible to show where text is being inserted), but `handleKey` short-circuits with `return m, nil` for non-mic keys.

Add to `InputBar`:
```go
func (ib *InputBar) SetSTTText(text string)   // replace STT portion, preserve prefix
func (ib *InputBar) ClearSTTState()            // clear STT tracking, unlock
```

The `InputBar` tracks `sttStart int` — the cursor position where STT text begins. `SetSTTText` replaces everything from `sttStart` to end-of-input with the new transcription. `ClearSTTState` resets this so subsequent typing works normally.

### Mic Button Behavior

| State | Mic Color | Click Action |
|-------|-----------|-------------|
| Idle | TextDim | Start recording |
| Recording | Red/Accent | Stop recording → final transcribe |
| Transcribing | Primary (animated) | No-op |

### Key Binding

Add **Ctrl+M** for mic toggle. Add to `ChatKeyBindings()` and `AllShortcuts()`.

### Message Flow

1. **Mic press (idle):**
   - If transcriber is nil, kick off model download → `ModelDownloadProgressMsg` → `ModelReadyMsg` → create transcriber → then start recording
   - If transcriber exists, start recording immediately
   - `recorder.Start()` → return `RecordingStartedMsg`
   - Set `recording = RecordingActive`
   - **Lock input** — note current cursor position as `sttStart`
   - Start streaming tick: return `tea.Tick(2500ms, streamTickMsg)`

2. **streamTickMsg (while recording):**
   - If a transcription is already in-flight, skip (return next tick only)
   - `snapshot := recorder.Snapshot()` — grab current audio buffer
   - Run `transcriber.Transcribe(snapshot)` in a `tea.Cmd`
   - Return `StreamSegmentMsg{Text: partial, Final: false}`
   - Schedule next tick

3. **StreamSegmentMsg (partial):**
   - `m.input.SetSTTText(msg.Text)` — replace STT portion in input
   - Update layout if height changed

4. **Mic press (recording) or 30s timeout:**
   - `samples := recorder.Stop()` → set `recording = RecordingTranscribing`
   - Run final `transcriber.Transcribe(samples)` in a `tea.Cmd`
   - Return `StreamSegmentMsg{Text: finalText, Final: true}`

5. **StreamSegmentMsg (final):**
   - `m.input.SetSTTText(msg.Text)` — set definitive text
   - `m.input.ClearSTTState()` — unlock input
   - Set `recording = RecordingIdle`

6. **TranscribeErrorMsg:**
   - Show error toast / status message
   - `m.input.ClearSTTState()` — unlock input
   - Set `recording = RecordingIdle`

### Model Download UX

On first mic press when no model is cached:
- Show a status line below the input: "Downloading whisper model... 42%"
- Once downloaded, auto-start recording
- On error, show toast and return to idle

## Build Integration

### Build Strategy: Pre-compiled libwhisper.a via cmake

The official whisper.cpp Go bindings **require** pre-building static libraries. There is no vendored single-file option (the `paradoxe35/whisper.cpp-go` fork was evaluated and rejected — it pins to whisper.cpp ~v1.2, 2+ years behind, with hardcoded x86 SIMD flags and zero community).

### Verified whisper.cpp CGO link flags (from `bindings/go/whisper.go`):

```go
#cgo LDFLAGS: -lwhisper -lggml -lggml-base -lggml-cpu -lm -lstdc++
#cgo linux LDFLAGS: -fopenmp
```

### mise.toml — New build tasks

```toml
[tasks."whisper:setup"]
description = "Clone and build whisper.cpp static libraries"
run = """
WHISPER_TAG="v1.8.3"
WHISPER_DIR="$PWD/third_party/whisper.cpp"
if [ ! -d "$WHISPER_DIR" ]; then
  git clone --depth 1 --branch "$WHISPER_TAG" https://github.com/ggml-org/whisper.cpp.git "$WHISPER_DIR"
fi
cmake -S "$WHISPER_DIR" -B "$WHISPER_DIR/build" \
  -DCMAKE_BUILD_TYPE=Release \
  -DBUILD_SHARED_LIBS=OFF
cmake --build "$WHISPER_DIR/build" --target whisper -- -j$(nproc)
"""

[tasks.build]
description = "Build workfort binary"
depends = ["whisper:setup"]
run = """
WHISPER_DIR="$PWD/third_party/whisper.cpp"
C_INCLUDE_PATH="$WHISPER_DIR/include:$WHISPER_DIR/ggml/include" \
LIBRARY_PATH="$WHISPER_DIR/build/src:$WHISPER_DIR/build/ggml/src" \
CGO_ENABLED=1 \
go build -o build/workfort ./cmd/workfort
"""
sources = ["**/*.go", "go.mod", "go.sum"]
outputs = ["build/workfort"]
```

### System Requirements

| Platform | Requirements |
|----------|-------------|
| Linux | `cmake`, `g++` (or `clang++`), `libgomp` (for `-fopenmp`) |
| macOS | `cmake`, Xcode command line tools |

malgo adds no system requirements (vendors miniaudio inline, links only `-ldl -lpthread -lm`).

### .gitignore

Add `third_party/` to `.gitignore` — whisper.cpp sources are cloned at build time, not checked in.

## New Files

| File | Purpose |
|------|---------|
| `pkg/stt/recorder.go` | malgo microphone capture |
| `pkg/stt/transcriber.go` | whisper.cpp CGO transcription |
| `pkg/stt/model.go` | Model download and cache management |
| `pkg/stt/convert.go` | int16 → float32 PCM conversion |

## Modified Files

| File | Change |
|------|--------|
| `internal/chat/model.go` | Recording state, mic handler, streaming tick loop, input locking, lazy transcriber |
| `internal/chat/input.go` | `SetSTTText()`, `ClearSTTState()`, `sttStart` tracking |
| `internal/chat/keys.go` | Ctrl+M binding, shortcuts modal entry |
| `internal/chat/model.go` (View) | Dynamic mic button color, download progress |
| `pkg/config/viper.go` | stt-model, stt-language, stt-threads defaults |
| `go.mod` | Add malgo, whisper.cpp bindings |
| `mise.toml` | CGO_ENABLED=1 in build task |

## Implementation Order

1. `pkg/stt/convert.go` — int16 ↔ float32 conversion utilities
2. `pkg/stt/recorder.go` — malgo capture (can test independently)
3. `pkg/stt/model.go` — model download with progress callback
4. `pkg/stt/transcriber.go` — whisper CGO wrapper
5. `internal/chat/keys.go` — Ctrl+M binding
6. `internal/chat/model.go` — recording state machine, mic handler, all message types
7. `mise.toml` — CGO_ENABLED=1
8. Build and test end-to-end

Steps 1-3 are pure library code, testable without the TUI. Step 4 needs a downloaded model. Steps 5-7 wire everything into the UI.

## Error Handling

| Scenario | Behavior |
|----------|----------|
| No microphone found | Toast: "No microphone detected" |
| Model download fails | Toast: "Failed to download model", retry on next mic press |
| Transcription returns empty | No-op (don't insert empty string) |
| Transcription error | Toast with error, return to idle |
| Recording while disconnected | Still works — voice is local-only |

## Risk Assessment

| Component | Risk | Notes |
|-----------|------|-------|
| malgo (audio capture) | **Low** | Stable API, `go get` works, no system deps, actively maintained (v0.11.24, Nov 2025) |
| whisper.cpp Go bindings | **Medium** | API verified and actively maintained (VAD added Dec 2025), but cmake pre-build adds contributor friction. Issue #2689 shows users struggling with build setup. |
| CGO coexistence | **Low** | malgo and whisper.cpp use separate C/C++ codebases with no symbol conflicts. Verified compatible. |
| Streaming quality | **Medium** | Re-transcribing growing buffer every 2.5s works but whisper may produce different text as context grows. Final pass should stabilize. Needs real-world testing. |
| Build tooling | **Medium** | `mise` task wrapping cmake is straightforward but adds `cmake` + C++ compiler as build prerequisites. Contributors who only touch Go code need these installed. |

## Open Questions

- **Streaming tick interval:** Plan uses 2.5s. Shorter (1.5s) gives faster feedback but more CPU. Longer (3-4s) is cheaper but feels laggy. What feels right?
- **GPU acceleration:** whisper.cpp supports CUDA/Metal/Vulkan. Worth exposing in build flags for users with GPUs?
- **Model selection UI:** Add model picker to settings modal, or just use config file?
- **Visual feedback during recording:** Color change on mic button is the minimum. Add a recording timer or waveform in the input area?
- **Cancel recording:** Should Esc cancel a recording and discard the audio, or only the mic toggle stops it?
