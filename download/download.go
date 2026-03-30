package download

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/neur0map/wallf/source"
)

// ErrAlreadyExists is returned by Fetch when a file with the same name already
// exists in the download directory.
var ErrAlreadyExists = errors.New("file already exists")

// ErrDuplicateContent is returned by Fetch when the downloaded content has a
// SHA-256 hash that matches an existing file in the download directory.
var ErrDuplicateContent = errors.New("duplicate content")

// Downloader manages downloading wallpapers to a directory with deduplication.
type Downloader struct {
	dir       string
	hashIndex map[string]bool
	client    *http.Client
}

// New creates a Downloader that saves files to dir. hashIndex is the set of
// known SHA-256 hashes (hex-encoded) already present in dir; callers should
// build it with BuildHashIndex.
func New(dir string, hashIndex map[string]bool) *Downloader {
	return &Downloader{
		dir:       dir,
		hashIndex: hashIndex,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// BuildHashIndex scans dir for image files (.jpg, .jpeg, .png, .webp) and
// returns a map of hex-encoded SHA-256 hashes to true.
func BuildHashIndex(dir string) (map[string]bool, error) {
	index := make(map[string]bool)

	entries, err := os.ReadDir(dir)
	if err != nil {
		return index, fmt.Errorf("download: read dir %s: %w", dir, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		ext := strings.ToLower(filepath.Ext(name))
		switch ext {
		case ".jpg", ".jpeg", ".png", ".webp":
			// proceed
		default:
			continue
		}

		hash, err := hashFile(filepath.Join(dir, name))
		if err != nil {
			return index, fmt.Errorf("download: hash %s: %w", name, err)
		}
		index[hash] = true
	}

	return index, nil
}

// Fetch downloads the wallpaper described by result into the Downloader's
// directory. It returns the path to the saved file. The following sentinel
// errors may be returned:
//
//   - ErrAlreadyExists: a file with the same name already exists.
//   - ErrDuplicateContent: the downloaded bytes match an existing file's hash.
func (d *Downloader) Fetch(result source.WallpaperResult) (string, error) {
	filename := result.Filename()
	destPath := filepath.Join(d.dir, filename)

	// Filename dedup.
	if _, err := os.Stat(destPath); err == nil {
		return "", ErrAlreadyExists
	}

	// Download to a temp file in the same directory so that the final rename
	// is atomic (same filesystem).
	tmp, err := os.CreateTemp(d.dir, ".wallf-tmp-*")
	if err != nil {
		return "", fmt.Errorf("download: create temp file: %w", err)
	}
	tmpPath := tmp.Name()
	// Clean up temp file on any error path.
	defer func() {
		if _, statErr := os.Stat(tmpPath); statErr == nil {
			os.Remove(tmpPath)
		}
	}()

	resp, err := d.client.Get(result.FullURL)
	if err != nil {
		tmp.Close()
		return "", fmt.Errorf("download: GET %s: %w", result.FullURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		tmp.Close()
		return "", fmt.Errorf("download: unexpected status %d for %s", resp.StatusCode, result.FullURL)
	}

	// Write to temp file while computing SHA-256 via io.MultiWriter.
	h := sha256.New()
	mw := io.MultiWriter(tmp, h)
	if _, err := io.Copy(mw, resp.Body); err != nil {
		tmp.Close()
		return "", fmt.Errorf("download: write temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return "", fmt.Errorf("download: close temp: %w", err)
	}

	// Content dedup.
	hash := hex.EncodeToString(h.Sum(nil))
	if d.hashIndex[hash] {
		return "", ErrDuplicateContent
	}

	// Atomic rename to final destination.
	if err := os.Rename(tmpPath, destPath); err != nil {
		return "", fmt.Errorf("download: rename to dest: %w", err)
	}

	// Record the new hash so subsequent calls in the same session also dedup.
	d.hashIndex[hash] = true

	return destPath, nil
}

// InvalidateSkwdCache removes the skwd checksum file at path. If the file does
// not exist the call is a no-op.
func InvalidateSkwdCache(path string) {
	os.Remove(path) //nolint:errcheck – intentional no-op on missing file
}

// hashFile computes the hex-encoded SHA-256 hash of the file at path.
func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
