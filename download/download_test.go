package download_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/neur0map/wallf/download"
	"github.com/neur0map/wallf/source"
)

// --- helpers ---

func makeResult(id, url string) source.WallpaperResult {
	return source.WallpaperResult{
		ID:      id,
		Source:  "wallhaven",
		FullURL: url,
	}
}

// --- BuildHashIndex ---

func TestBuildHashIndex(t *testing.T) {
	dir := t.TempDir()

	// Write two small files with distinct content.
	if err := os.WriteFile(filepath.Join(dir, "a.jpg"), []byte("content-a"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "b.png"), []byte("content-b"), 0644); err != nil {
		t.Fatal(err)
	}

	index, err := download.BuildHashIndex(dir)
	if err != nil {
		t.Fatalf("BuildHashIndex: %v", err)
	}
	if len(index) != 2 {
		t.Fatalf("want 2 hashes, got %d", len(index))
	}
}

func TestBuildHashIndexEmpty(t *testing.T) {
	dir := t.TempDir()

	index, err := download.BuildHashIndex(dir)
	if err != nil {
		t.Fatalf("BuildHashIndex on empty dir: %v", err)
	}
	if len(index) != 0 {
		t.Fatalf("want 0 hashes, got %d", len(index))
	}
}

// --- Fetch ---

func TestFetchFile(t *testing.T) {
	const body = "fake-jpeg-bytes"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(body)) //nolint:errcheck
	}))
	defer srv.Close()

	dir := t.TempDir()
	d := download.New(dir, make(map[string]bool))

	result := source.WallpaperResult{
		ID:      "abc123",
		Source:  "wallhaven",
		FullURL: srv.URL + "/wallhaven-abc123.jpg",
	}

	path, err := d.Fetch(result)
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}

	wantName := "wallhaven-abc123.jpg"
	wantPath := filepath.Join(dir, wantName)
	if path != wantPath {
		t.Errorf("path = %q, want %q", path, wantPath)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(data) != body {
		t.Errorf("content = %q, want %q", string(data), body)
	}
}

func TestFetchSkipsExistingFilename(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("data")) //nolint:errcheck
	}))
	defer srv.Close()

	dir := t.TempDir()

	// Pre-create the file.
	existingPath := filepath.Join(dir, "wallhaven-abc123.jpg")
	if err := os.WriteFile(existingPath, []byte("old"), 0644); err != nil {
		t.Fatal(err)
	}

	d := download.New(dir, make(map[string]bool))
	result := source.WallpaperResult{
		ID:      "abc123",
		Source:  "wallhaven",
		FullURL: srv.URL + "/wallhaven-abc123.jpg",
	}

	_, err := d.Fetch(result)
	if !errors.Is(err, download.ErrAlreadyExists) {
		t.Errorf("want ErrAlreadyExists, got %v", err)
	}
}

func TestFetchSkipsDuplicateContent(t *testing.T) {
	const body = "duplicate-content"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(body)) //nolint:errcheck
	}))
	defer srv.Close()

	dir := t.TempDir()

	// Write an existing file with the same content so the hash index matches.
	existingPath := filepath.Join(dir, "wallhaven-existing.jpg")
	if err := os.WriteFile(existingPath, []byte(body), 0644); err != nil {
		t.Fatal(err)
	}

	// Build hash index that includes the duplicate.
	index, err := download.BuildHashIndex(dir)
	if err != nil {
		t.Fatalf("BuildHashIndex: %v", err)
	}

	d := download.New(dir, index)
	result := source.WallpaperResult{
		ID:      "newfile",
		Source:  "wallhaven",
		FullURL: srv.URL + "/wallhaven-newfile.jpg",
	}

	_, err = d.Fetch(result)
	if !errors.Is(err, download.ErrDuplicateContent) {
		t.Errorf("want ErrDuplicateContent, got %v", err)
	}

	// Ensure the new file was NOT saved.
	if _, statErr := os.Stat(filepath.Join(dir, "wallhaven-newfile.jpg")); statErr == nil {
		t.Error("new file should not have been saved")
	}
}

func TestFetchHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	dir := t.TempDir()
	d := download.New(dir, make(map[string]bool))

	result := source.WallpaperResult{
		ID:      "bad",
		Source:  "wallhaven",
		FullURL: srv.URL + "/wallhaven-bad.jpg",
	}

	_, err := d.Fetch(result)
	if err == nil {
		t.Error("expected an error for 404 response, got nil")
	}
}

// --- InvalidateSkwdCache ---

func TestInvalidateSkwdCache(t *testing.T) {
	dir := t.TempDir()
	checksumPath := filepath.Join(dir, "checksum.txt")

	if err := os.WriteFile(checksumPath, []byte("abc"), 0644); err != nil {
		t.Fatal(err)
	}

	download.InvalidateSkwdCache(checksumPath)

	if _, err := os.Stat(checksumPath); err == nil {
		t.Error("checksum.txt should have been deleted")
	}
}

func TestInvalidateSkwdCacheMissing(t *testing.T) {
	dir := t.TempDir()
	nonexistent := filepath.Join(dir, "does-not-exist.txt")

	// Must not panic.
	download.InvalidateSkwdCache(nonexistent)
}
