# Whisper Speech-to-Text Integration

Wire the mic button in the chat input bar to record audio, transcribe via whisper.cpp, and insert the resulting text into the message input.

## Problem

The chat UI has a mic button (`micBtn` in `model.go` View) that renders but does nothing — the click handler has the comment `// Mic button (left) — unwired`. Voice input is a key usability feature for a chat app, especially for long-form messages or when hands are occupied.

## Approach

Use **whisper.cpp's whisper-server** (MIT license, GPL-2.0 compatible) as a local HTTP transcription service. The server is managed as a subprocess — started lazily on first mic press, kept resident for subsequent requests. Audio is captured from the system microphone, sent as a WAV payload to the OpenAI-compatible `/v1/audio/transcriptions` endpoint, and the returned text is inserted into the input bar.

This approach avoids cgo in the main binary, keeps license boundaries clean, and reuses model loads across transcription requests.

## Architecture

```
┌────────────────────────────────────────────────────────────┐
│  Bubble Tea (ChatModel)                                    │
│                                                            │
│  ┌──────────┐   RecordingStartMsg   ┌──────────────────┐  │
│  │ Mic Btn  │ ────────────────────> │ Audio Recorder   │  │
│  │ (click)  │                       │ (goroutine)      │  │
│  └──────────┘   RecordingDoneMsg    │ portaudio / sox  │  │
│       │       <──────────────────── └──────────────────┘  │
│       │                                     │              │
│       │         TranscribeCmd               │ WAV bytes    │
│       │       ─────────────────────>        v              │
│       │                             ┌──────────────────┐  │
│       │         TranscribedMsg      │ whisper-server   │  │
│       │       <──────────────────── │ (subprocess)     │  │
│       │                             │ :8178/v1/audio   │  │
│       v                             └──────────────────┘  │
│  ┌──────────┐                                              │
│  │ InputBar │ .InsertString(transcribed text)               │
│  └──────────┘                                              │
└────────────────────────────────────────────────────────────┘
```

All I/O happens in tea.Cmd goroutines. The main Update loop only processes typed messages (`RecordingStartMsg`, `RecordingDoneMsg`, `TranscribedMsg`, `TranscribeErrorMsg`).

## Audio Capture

### Option A: subprocess (recommended initially)

Shell out to `sox` (SoX — Sound eXchange) or `arecord` (ALSA) to record to a temp WAV file:

```
sox -d -r 16000 -c 1 -b 16 /tmp/workfort-rec.wav silence 1 0.1 1% 1 2.0 3%
```

- `-d` = default mic input
- `-r 16000 -c 1 -b 16` = 16kHz mono 16-bit (whisper's expected format)
- `silence ...` = auto-stop after 2s of silence (optional, can also use explicit stop)

On macOS, `sox` works the same way. On Linux, `arecord` is an alternative.

Pros: No cgo, works cross-platform with a common CLI tool.
Cons: Requires sox/arecord installed. Slight startup latency.

### Option B: portaudio cgo bindings

Use `github.com/gordonklaus/portaudio` to capture audio directly in Go. Lower latency, no external dependency at runtime, but adds cgo and complicates cross-compilation.

**Decision:** Start with subprocess (Option A). Gate on `sox` availability at runtime; show an error toast if missing. Revisit portaudio if latency becomes a problem.

### Recording Flow

1. Mic press -> start `sox` subprocess in a tea.Cmd goroutine
2. Second mic press (or timeout) -> send SIGINT to `sox`, collect WAV file
3. Return `RecordingDoneMsg{wavPath}` to Update loop

## Whisper Integration

### Server Lifecycle

A new `pkg/whisper` package manages the whisper-server subprocess:

```go
package whisper

type Server struct {
    cmd      *exec.Cmd
    port     int
    modelPath string
    ready    bool
    mu       sync.Mutex
}

func NewServer(modelPath string, port int) *Server
func (s *Server) Start() error      // launch subprocess, wait for ready
func (s *Server) Stop()              // SIGTERM + wait
func (s *Server) Transcribe(wavPath string) (string, error) // POST to /v1/audio/transcriptions
func (s *Server) EnsureRunning() error // idempotent start
```

- **Port:** Use `localhost:8178` by default (configurable via `whisper-port` config key).
- **Startup:** Launch `whisper-server -m <model> --port 8178`. Poll `/health` until 200 or 10s timeout.
- **Shutdown:** `Stop()` called from `ChatModel`'s cleanup / `tea.Quit` handler.
- **Binary location:** Look for `whisper-server` in `$PATH`. If not found, check `$XDG_DATA_HOME/workfort/bin/whisper-server`. Config key `whisper-server-path` for explicit override.

### API Usage

whisper-server exposes an OpenAI-compatible endpoint:

```
POST http://localhost:8178/v1/audio/transcriptions
Content-Type: multipart/form-data

file=@/tmp/workfort-rec.wav
model=default
response_format=text
language=en          (optional, auto-detect if omitted)
```

Response: plain text transcription.

### Error Handling

- Server not found: show one-time toast "whisper-server not found — install whisper.cpp"
- Server crash: detect via `cmd.Wait()`, set `ready=false`, retry on next mic press
- Transcription failure: show error toast, discard recording
- Empty transcription: no-op (don't insert empty string)

## TUI Changes

### Recording State

Add to `ChatModel`:

```go
type RecordingState int

const (
    RecordingIdle RecordingState = iota
    RecordingActive
    RecordingTranscribing
)

// In ChatModel:
recording      RecordingState
whisperServer  *whisper.Server
recorderCmd    *exec.Cmd
```

### Mic Button Behavior

| State | Mic Icon | Button Color | Click Action |
|-------|----------|-------------|--------------|
| Idle | `󰍬` | TextDim | Start recording |
| Recording | `󰍬` | Red (Accent) | Stop recording |
| Transcribing | `󰍬` | Primary (pulsing via tick) | No-op |

The mic button area already handles clicks in `handleMouse` — the `msg.X >= btnAreaX && msg.X < btnAreaX+8` branch currently does nothing. Wire it to toggle recording.

### Key Binding

Add **Ctrl+M** as a keyboard shortcut for mic toggle (matches the button). Update help bar and shortcuts modal.

### Message Flow (tea.Cmd chain)

```go
// 1. Mic press -> start recording
func startRecordingCmd() tea.Msg {
    // launch sox, return when started
    return RecordingStartedMsg{cmd: cmd}
}

// 2. Stop recording -> get WAV
func stopRecordingCmd(cmd *exec.Cmd, wavPath string) tea.Msg {
    cmd.Process.Signal(syscall.SIGINT)
    cmd.Wait()
    return RecordingDoneMsg{wavPath: wavPath}
}

// 3. Transcribe -> insert text
func transcribeCmd(server *whisper.Server, wavPath string) tea.Msg {
    text, err := server.Transcribe(wavPath)
    os.Remove(wavPath) // cleanup temp file
    if err != nil {
        return TranscribeErrorMsg{err: err}
    }
    return TranscribedMsg{text: text}
}
```

### InputBar Change

Add a method to insert transcribed text at cursor position:

```go
func (ib *InputBar) InsertString(s string) {
    ib.textarea.InsertString(s)
    ib.updateHeight()
}
```

## Model Management

### Download & Cache

Models are stored in `$XDG_DATA_HOME/workfort/models/` (typically `~/.local/share/workfort/models/`).

| Model | Size | Speed | Quality | Default |
|-------|------|-------|---------|---------|
| `ggml-tiny.en.bin` | 75 MB | Fastest | Good for English | Yes |
| `ggml-base.en.bin` | 142 MB | Fast | Better | No |
| `ggml-small.en.bin` | 466 MB | Moderate | Best practical | No |

Download on first use from `https://huggingface.co/ggerganov/whisper.cpp/resolve/main/`. Show progress bar in a modal (or use the message pane as status area). Config key `whisper-model` selects which model to use.

### Config Keys

```yaml
whisper-model: "ggml-tiny.en.bin"
whisper-server-path: ""           # auto-detect if empty
whisper-port: 8178
whisper-language: ""              # auto-detect if empty
```

## Dependencies & Licensing

| Dependency | License | GPL-2.0-Only Compatible | Integration |
|------------|---------|------------------------|-------------|
| whisper.cpp / whisper-server | MIT | Yes | Subprocess |
| SoX (sox) | GPL-2.0+ | Yes | Subprocess |
| arecord (alsa-utils) | GPL-2.0+ | Yes | Subprocess (Linux alt) |
| Model weights (ggml-*.bin) | MIT | Yes | Downloaded separately |

**Incompatible (do not use):**

| Library | License | Issue |
|---------|---------|-------|
| go-whisper (mutablelogic) | Apache-2.0 | Incompatible with GPL-2.0-Only |
| insanely-fast-whisper | Apache-2.0 | Incompatible with GPL-2.0-Only |
| WhisperX | BSD-4-Clause | Incompatible with GPL-2.0-Only |

## Implementation Steps

1. **`pkg/whisper/server.go`** — Server struct, Start/Stop/EnsureRunning, health poll loop
2. **`pkg/whisper/transcribe.go`** — Transcribe method (multipart POST to /v1/audio/transcriptions, parse response)
3. **`pkg/whisper/download.go`** — Model download with progress callback, XDG cache path resolution
4. **`internal/chat/recorder.go`** — sox subprocess management: `startRecording()`, `stopRecording()` returning tea.Cmd functions
5. **`internal/chat/msgtypes.go`** — Add `RecordingStartedMsg`, `RecordingDoneMsg`, `TranscribedMsg`, `TranscribeErrorMsg`
6. **`internal/chat/model.go`** — Recording state fields, mic click handler, Ctrl+M binding, Update cases for all new message types, whisper server lifecycle (start on first mic press, stop on quit)
7. **`internal/chat/input.go`** — `InsertString` method on InputBar
8. **`internal/chat/model.go` (View)** — Dynamic mic button color based on recording state
9. **`internal/chat/keys.go`** — Add Ctrl+M to help bar and shortcuts modal
10. **`pkg/config/viper.go`** — Add whisper config key defaults (`whisper-model`, `whisper-server-path`, `whisper-port`, `whisper-language`)
11. **Verify** — Build, test with real whisper-server, record/transcribe/insert flow end-to-end

## Open Questions

- **Max recording duration:** Hard cap at 30s? 60s? Whisper handles up to 30s natively; longer clips need chunking.
- **Streaming transcription:** whisper-server supports streaming mode. Worth wiring up for real-time partial results, or is batch-on-stop sufficient for v1?
- **Multi-language:** Default to English-only models (`ggml-tiny.en.bin`) for size/speed, or offer multilingual models? The `whisper-language` config key is a placeholder for this.
- **sox availability:** Should we bundle a static sox binary, or just error with install instructions? Most Linux distros have it in repos; macOS has it via Homebrew.
- **Visual feedback during recording:** Simple color change on mic button, or add a recording duration timer / waveform indicator in the input area?
- **Concurrent recordings:** Block mic press while transcribing, or allow queuing? Simplest to block.
