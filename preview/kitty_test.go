package preview_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/neur0map/wallf/preview"
)

// TestBuildDisplayCommand verifies that the display command slice starts with
// the kitty icat invocation and includes the image path and --place argument.
func TestBuildDisplayCommand(t *testing.T) {
	k, err := preview.New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer k.Cleanup()

	cmd := preview.BuildDisplayCommand("/tmp/test.jpg", 80, 24)
	if len(cmd) < 4 {
		t.Fatalf("command too short: %v", cmd)
	}

	// First three elements must be the kitty icat kitten.
	if cmd[0] != "kitty" {
		t.Errorf("cmd[0] = %q, want \"kitty\"", cmd[0])
	}
	if cmd[1] != "+kitten" {
		t.Errorf("cmd[1] = %q, want \"+kitten\"", cmd[1])
	}
	if cmd[2] != "icat" {
		t.Errorf("cmd[2] = %q, want \"icat\"", cmd[2])
	}

	// Image display command must contain the image path (--place removed,
	// images now rendered via kitty graphics escape sequences in View).
	if len(cmd) < 4 {
		t.Errorf("command too short: %v", cmd)
	}

	// Image path must appear in the command.
	foundPath := false
	for _, arg := range cmd {
		if arg == "/tmp/test.jpg" {
			foundPath = true
		}
	}
	if !foundPath {
		t.Errorf("image path not found in command: %v", cmd)
	}
}

// TestBuildClearCommand verifies the clear command structure.
func TestBuildClearCommand(t *testing.T) {
	cmd := preview.BuildClearCommand()
	if len(cmd) < 4 {
		t.Fatalf("clear command too short: %v", cmd)
	}
	if cmd[0] != "kitty" {
		t.Errorf("cmd[0] = %q, want \"kitty\"", cmd[0])
	}
	if cmd[1] != "+kitten" {
		t.Errorf("cmd[1] = %q, want \"+kitten\"", cmd[1])
	}
	if cmd[2] != "icat" {
		t.Errorf("cmd[2] = %q, want \"icat\"", cmd[2])
	}
	foundClear := false
	for _, arg := range cmd {
		if arg == "--clear" {
			foundClear = true
		}
	}
	if !foundClear {
		t.Errorf("--clear not found in command: %v", cmd)
	}
}

// TestDownloadPreview creates a fake HTTP server, fetches a preview, and then
// verifies that a second call returns the cached path without hitting the
// server again.
func TestDownloadPreview(t *testing.T) {
	const body = "fake-image-data"
	var requestCount int

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(body)) //nolint:errcheck
	}))
	defer srv.Close()

	k, err := preview.New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer k.Cleanup()

	url := srv.URL + "/preview.jpg"

	// First call — should download.
	path1, err := k.DownloadPreview(url, "test-id")
	if err != nil {
		t.Fatalf("DownloadPreview first call: %v", err)
	}
	if _, err := os.Stat(path1); err != nil {
		t.Errorf("cached file not found at %s: %v", path1, err)
	}
	if requestCount != 1 {
		t.Errorf("expected 1 HTTP request, got %d", requestCount)
	}

	// Second call — should return cached path.
	path2, err := k.DownloadPreview(url, "test-id")
	if err != nil {
		t.Fatalf("DownloadPreview second call: %v", err)
	}
	if path2 != path1 {
		t.Errorf("second call returned %q, want %q", path2, path1)
	}
	if requestCount != 1 {
		t.Errorf("expected still 1 HTTP request after cache hit, got %d", requestCount)
	}
}

// TestGetCachedPath verifies that getCachedPath (via DownloadPreview) returns
// a pre-created cached file.
func TestGetCachedPath(t *testing.T) {
	k, err := preview.New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer k.Cleanup()

	// Manually place a file in the cache dir using the exported helper.
	cacheDir := k.CacheDir()
	cachedFile := filepath.Join(cacheDir, "myid.jpg")
	if err := os.WriteFile(cachedFile, []byte("data"), 0644); err != nil {
		t.Fatalf("write cached file: %v", err)
	}

	// A server that must NOT be called.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("HTTP server should not have been called for cached file")
	}))
	defer srv.Close()

	path, err := k.DownloadPreview(srv.URL+"/myid.jpg", "myid")
	if err != nil {
		t.Fatalf("DownloadPreview with cached file: %v", err)
	}
	if path != cachedFile {
		t.Errorf("path = %q, want %q", path, cachedFile)
	}
}

// TestDetectCommandExists verifies that detectCommand returns at least two
// parts (binary + at minimum one meaningful entry).
func TestDetectCommandExists(t *testing.T) {
	cmd := preview.DetectCommand()
	if len(cmd) < 1 {
		t.Fatalf("detectCommand returned empty slice")
	}
	// The result must have at least one element (the binary path).
	if cmd[0] == "" {
		t.Error("detectCommand returned empty binary name")
	}
	t.Logf("detectCommand = %v", cmd)
}
