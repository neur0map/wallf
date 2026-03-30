package config

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// Config is the top-level configuration structure.
type Config struct {
	General GeneralConfig `toml:"general"`
}

// GeneralConfig holds the general section of the configuration file.
type GeneralConfig struct {
	DownloadDir   string `toml:"download_dir"`
	DefaultSource string `toml:"default_source"`
	MinResolution string `toml:"min_resolution"`
}

// Default returns a Config populated with sensible defaults.
func Default() Config {
	return Config{
		General: GeneralConfig{
			DownloadDir:   "~/Pictures/Wallpapers",
			DefaultSource: "wallhaven",
			MinResolution: "2560x1440",
		},
	}
}

// ExpandTilde replaces a leading "~" in p with the user's home directory.
// Paths that do not start with "~" are returned unchanged.
func ExpandTilde(p string) string {
	if !strings.HasPrefix(p, "~") {
		return p
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return p
	}
	// Replace only the leading ~, preserving the rest of the path.
	return filepath.Join(home, p[1:])
}

// DownloadDir returns the fully-expanded download directory path.
func (c Config) DownloadDir() string {
	return ExpandTilde(c.General.DownloadDir)
}

// Path returns the default path to the configuration file.
// Respects $XDG_CONFIG_HOME, falls back to ~/.config/wallf/config.toml.
func Path() string {
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		configDir = filepath.Join(home, ".config")
	}
	return filepath.Join(configDir, "wallf", "config.toml")
}

// Exists reports whether the configuration file at Path() exists.
func Exists() bool {
	_, err := os.Stat(Path())
	return err == nil
}

// Load reads a Config from path. If the file contains malformed TOML, a
// warning is logged and the default Config is returned (no error). Any other
// OS-level error (e.g. file not found) is returned as-is.
func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	cfg := Default()
	if err := toml.Unmarshal(data, &cfg); err != nil {
		log.Printf("wallf: warning: malformed config at %s: %v — using defaults", path, err)
		return Default(), nil
	}
	return cfg, nil
}

// Save marshals c to TOML and writes it to path. Parent directories are
// created as needed.
func Save(c Config, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := toml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
