package stt

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Work-Fort/WorkFort/pkg/config"
)

const modelBaseURL = "https://huggingface.co/ggerganov/whisper.cpp/resolve/main"

// ModelsDir returns the directory where whisper models are stored.
func ModelsDir() string {
	return filepath.Join(config.GlobalPaths.DataDir, "models")
}

// EnsureModel checks whether the named model is cached locally. If not, it
// downloads from HuggingFace. progressFn is called with values in [0.0, 1.0]
// during download. Returns the absolute path to the model file.
func EnsureModel(name string, progressFn func(float64)) (string, error) {
	dir := ModelsDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("stt: create models dir: %w", err)
	}

	path := filepath.Join(dir, name)
	if _, err := os.Stat(path); err == nil {
		return path, nil
	}

	url := modelBaseURL + "/" + name
	resp, err := http.Get(url) //nolint:gosec
	if err != nil {
		return "", fmt.Errorf("stt: download model: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("stt: download model: HTTP %d", resp.StatusCode)
	}

	tmp, err := os.CreateTemp(dir, name+".tmp.*")
	if err != nil {
		return "", fmt.Errorf("stt: create temp file: %w", err)
	}
	tmpPath := tmp.Name()
	defer func() {
		tmp.Close()
		os.Remove(tmpPath)
	}()

	var reader io.Reader = resp.Body
	if resp.ContentLength > 0 && progressFn != nil {
		reader = &progressReader{
			r:     resp.Body,
			total: resp.ContentLength,
			fn:    progressFn,
		}
	}

	if _, err := io.Copy(tmp, reader); err != nil {
		return "", fmt.Errorf("stt: write model: %w", err)
	}
	tmp.Close()

	if err := os.Rename(tmpPath, path); err != nil {
		return "", fmt.Errorf("stt: rename model: %w", err)
	}
	return path, nil
}

type progressReader struct {
	r       io.Reader
	total   int64
	current int64
	fn      func(float64)
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.r.Read(p)
	pr.current += int64(n)
	pr.fn(float64(pr.current) / float64(pr.total))
	return n, err
}
