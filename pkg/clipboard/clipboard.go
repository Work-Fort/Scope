package clipboard

import (
	"os/exec"
	"runtime"
	"strings"
)

// Read returns the current system clipboard contents.
// On Linux/Wayland it uses wl-paste, on Linux/X11 it uses xclip,
// on macOS it uses pbpaste.
func Read() (string, error) {
	name, args := readCmd()
	out, err := exec.Command(name, args...).Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// Write sets the system clipboard to the given text.
// On Linux/Wayland it uses wl-copy, on Linux/X11 it uses xclip,
// on macOS it uses pbcopy.
func Write(text string) error {
	name, args := writeCmd()
	cmd := exec.Command(name, args...)
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}

func readCmd() (string, []string) {
	if runtime.GOOS == "darwin" {
		return "pbpaste", nil
	}
	// Prefer Wayland tools, fall back to X11
	if _, err := exec.LookPath("wl-paste"); err == nil {
		return "wl-paste", []string{"--no-newline"}
	}
	return "xclip", []string{"-selection", "clipboard", "-o"}
}

func writeCmd() (string, []string) {
	if runtime.GOOS == "darwin" {
		return "pbcopy", nil
	}
	if _, err := exec.LookPath("wl-copy"); err == nil {
		return "wl-copy", nil
	}
	return "xclip", []string{"-selection", "clipboard"}
}
