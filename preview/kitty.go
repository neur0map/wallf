package preview

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Kitty implements the kitty graphics protocol for terminal image preview.
type Kitty struct {
	cacheDir string
	client   *http.Client
	// lastImage holds the escape sequence string for the current preview.
	// This gets embedded directly into bubbletea's View() output.
	lastImage string
}

// New creates a Kitty previewer with a temporary cache directory.
func New() (*Kitty, error) {
	cacheDir, err := os.MkdirTemp("", "wallf-previews")
	if err != nil {
		return nil, fmt.Errorf("preview: create cache dir: %w", err)
	}
	return &Kitty{
		cacheDir: cacheDir,
		client:   &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// Supported reports whether the kitty graphics protocol is available.
func Supported() bool {
	termProg := strings.ToLower(os.Getenv("TERM_PROGRAM"))
	term := strings.ToLower(os.Getenv("TERM"))

	return strings.Contains(termProg, "kitty") ||
		strings.Contains(termProg, "wezterm") ||
		strings.Contains(termProg, "ghostty") ||
		strings.Contains(term, "kitty")
}

// CacheDir returns the cache directory path (exported for tests).
func (k *Kitty) CacheDir() string { return k.cacheDir }

// Display sends the image to the terminal using kitty graphics protocol.
// Uses t=f (file path) so kitty reads the file directly — works with JPEG/PNG.
// Writes directly to stdout, bypassing bubbletea's renderer.
func (k *Kitty) Display(imagePath string, cols, rows int) error {
	absPath, err := filepath.Abs(imagePath)
	if err != nil {
		return fmt.Errorf("preview: abs path: %w", err)
	}

	encodedPath := base64.StdEncoding.EncodeToString([]byte(absPath))

	// Delete previous images, then display new one
	// t=f means payload is a base64-encoded file path
	// a=T means transmit and display
	// c/r = columns/rows to scale into
	clear := "\x1b_Ga=d,d=A\x1b\\"
	display := fmt.Sprintf("\x1b_Ga=T,t=f,c=%d,r=%d;%s\x1b\\", cols, rows, encodedPath)

	_, err = fmt.Fprint(os.Stdout, clear+display)
	return err
}

// Clear removes all kitty graphics images from the terminal.
func (k *Kitty) Clear() error {
	_, err := fmt.Fprint(os.Stdout, "\x1b_Ga=d,d=A\x1b\\")
	return err
}

// DownloadPreview fetches the image at url and caches it locally.
func (k *Kitty) DownloadPreview(url, id string) (string, error) {
	rawURL := url
	if idx := strings.Index(rawURL, "?"); idx != -1 {
		rawURL = rawURL[:idx]
	}
	ext := filepath.Ext(rawURL)
	if ext == "" {
		ext = ".jpg"
	}

	if cached, err := k.getCachedPath(id, ext); err == nil {
		return cached, nil
	}

	destPath := filepath.Join(k.cacheDir, id+ext)
	resp, err := k.client.Get(url)
	if err != nil {
		return "", fmt.Errorf("preview: GET %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("preview: status %d for %s", resp.StatusCode, url)
	}

	f, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("preview: create file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, resp.Body); err != nil {
		return "", fmt.Errorf("preview: write: %w", err)
	}
	return destPath, nil
}

// Cleanup removes the cache directory.
func (k *Kitty) Cleanup() {
	os.RemoveAll(k.cacheDir)
}

func (k *Kitty) getCachedPath(id, ext string) (string, error) {
	p := filepath.Join(k.cacheDir, id+ext)
	if _, err := os.Stat(p); err != nil {
		return "", err
	}
	return p, nil
}

// BuildDisplayCommand is kept for test compatibility.
func BuildDisplayCommand(imagePath string, cols, rows int) []string {
	return []string{"kitty", "+kitten", "icat", "--stdin", "no", imagePath}
}

// BuildClearCommand is kept for test compatibility.
func BuildClearCommand() []string {
	return []string{"kitty", "+kitten", "icat", "--clear"}
}

// DetectCommand is kept for test compatibility.
func DetectCommand() []string {
	path, err := exec.LookPath("kitty")
	if err != nil {
		return []string{"kitty"}
	}
	return []string{path}
}
