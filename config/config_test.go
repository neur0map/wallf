package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/neur0map/wallf/config"
)

func TestDefaultConfig(t *testing.T) {
	cfg := config.Default()
	if cfg.General.DefaultSource != "wallhaven" {
		t.Errorf("DefaultSource = %q, want %q", cfg.General.DefaultSource, "wallhaven")
	}
	if cfg.General.MinResolution != "2560x1440" {
		t.Errorf("MinResolution = %q, want %q", cfg.General.MinResolution, "2560x1440")
	}
}

func TestExpandTilde(t *testing.T) {
	home, _ := os.UserHomeDir()

	got := config.ExpandTilde("~/Pictures/Wallpapers")
	want := filepath.Join(home, "Pictures/Wallpapers")
	if got != want {
		t.Errorf("ExpandTilde(~/Pictures/Wallpapers) = %q, want %q", got, want)
	}

	abs := "/tmp/walls"
	if got2 := config.ExpandTilde(abs); got2 != abs {
		t.Errorf("ExpandTilde(%q) = %q, want unchanged", abs, got2)
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	cfg := config.Default()
	cfg.General.DefaultSource = "reddit"
	cfg.General.MinResolution = "1920x1080"

	if err := config.Save(cfg, path); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.General.DefaultSource != "reddit" {
		t.Errorf("DefaultSource = %q, want %q", loaded.General.DefaultSource, "reddit")
	}
	if loaded.General.MinResolution != "1920x1080" {
		t.Errorf("MinResolution = %q, want %q", loaded.General.MinResolution, "1920x1080")
	}
}

func TestLoadMissing(t *testing.T) {
	_, err := config.Load("/nonexistent/path/config.toml")
	if err == nil {
		t.Error("Load of nonexistent file should return an error")
	}
}

func TestLoadInvalid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	// Write malformed TOML
	if err := os.WriteFile(path, []byte("[[[[not valid toml"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfg, err := config.Load(path)
	if err != nil {
		t.Errorf("Load of malformed TOML should return nil error, got: %v", err)
	}
	// Should return defaults
	if cfg.General.DefaultSource != "wallhaven" {
		t.Errorf("DefaultSource = %q, want default %q", cfg.General.DefaultSource, "wallhaven")
	}
}

func TestDownloadDir(t *testing.T) {
	cfg := config.Default()
	home, _ := os.UserHomeDir()
	got := cfg.DownloadDir()
	if !strings.HasPrefix(got, home) {
		t.Errorf("DownloadDir() = %q, should start with home dir %q", got, home)
	}
	if strings.Contains(got, "~") {
		t.Errorf("DownloadDir() = %q, tilde was not expanded", got)
	}
}

func TestConfigPath(t *testing.T) {
	p := config.Path()
	if p == "" {
		t.Error("Path() should not be empty")
	}
}
