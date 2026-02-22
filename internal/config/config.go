// Package config loads and provides the NeoCognito configuration.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config holds all user-configurable settings.
type Config struct {
	// Editor to invoke for deep editing (default: $EDITOR or "vi")
	Editor string `toml:"editor"`

	// DataDir is the root of the block/asset/db storage.
	DataDir string `toml:"data_dir"`

	// GitCommit enables automatic git commits after every block write.
	GitCommit bool `toml:"git_commit"`

	// StaleDays is the number of days before a block is considered stale.
	StaleDays int `toml:"stale_days"`

	// Theme name to load from the themes directory.
	Theme string `toml:"theme"`

	// DefaultTags are prepended to every new block.
	DefaultTags []string `toml:"default_tags"`

	// SyncCmd is the command run when invoking neocognito sync.
	SyncCmd string `toml:"sync_cmd"`
}

// Default returns a Config with sensible defaults.
func Default() *Config {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}
	return &Config{
		Editor:    editor,
		DataDir:   defaultDataDir(),
		GitCommit: false,
		StaleDays: 14,
		Theme:     "default",
		SyncCmd:   "git add . && git commit -m 'sync' && git push",
	}
}

// Load reads the config file at the standard path, falling back to defaults
// for any missing fields.
func Load() (*Config, error) {
	cfg := Default()
	path := configPath()

	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		// No config file — return defaults silently
		return cfg, nil
	}

	if _, err := toml.DecodeFile(path, cfg); err != nil {
		return nil, fmt.Errorf("parsing config file %s: %w", path, err)
	}

	// Expand ~ in DataDir
	if len(cfg.DataDir) > 1 && cfg.DataDir[:2] == "~/" {
		home, _ := os.UserHomeDir()
		cfg.DataDir = filepath.Join(home, cfg.DataDir[2:])
	}

	return cfg, nil
}

// WriteDefault creates the config file with default values if it doesn't exist.
func WriteDefault() error {
	path := configPath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	if _, err := os.Stat(path); err == nil {
		return nil // already exists
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating config file: %w", err)
	}
	defer f.Close()

	tmpl := `# NeoCognito configuration
# See: https://github.com/lemondesk/neocognito

# editor = "nvim"
# data_dir = "~/.local/share/neocognito"
git_commit  = false
stale_days  = 14
theme       = "default"
# default_tags = ["@inbox"]
# sync_cmd = "git add . && git commit -m 'sync' && git push"
`
	_, err = f.WriteString(tmpl)
	return err
}

// Save writes the current configuration back to the config file.
func (c *Config) Save() error {
	path := configPath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating config file for save: %w", err)
	}
	defer f.Close()

	return toml.NewEncoder(f).Encode(c)
}

// ConfigPath returns the path to the config file.
func ConfigPath() string { return configPath() }

func configPath() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "neocognito", "config.toml")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "neocognito", "config.toml")
}

func defaultDataDir() string {
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return filepath.Join(xdg, "neocognito")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share", "neocognito")
}
