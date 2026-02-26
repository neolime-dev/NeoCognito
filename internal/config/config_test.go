package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/neolime-dev/neocognito/internal/config"
)

// isolate redirects XDG paths to temp dirs so tests never touch real config.
func isolate(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "config"))
	t.Setenv("XDG_DATA_HOME", filepath.Join(dir, "data"))
}

// --- Default() ---

func TestDefault_EditorFromEnv(t *testing.T) {
	t.Setenv("EDITOR", "nano")
	cfg := config.Default()
	if cfg.Editor != "nano" {
		t.Errorf("editor: want nano, got %q", cfg.Editor)
	}
}

func TestDefault_EditorFallback(t *testing.T) {
	t.Setenv("EDITOR", "")
	cfg := config.Default()
	if cfg.Editor != "vi" {
		t.Errorf("editor fallback: want vi, got %q", cfg.Editor)
	}
}

func TestDefault_StaleDays(t *testing.T) {
	cfg := config.Default()
	if cfg.StaleDays != 14 {
		t.Errorf("stale_days default: want 14, got %d", cfg.StaleDays)
	}
}

func TestDefault_DataDirFromXDG(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmp)
	cfg := config.Default()
	want := filepath.Join(tmp, "neocognito")
	if cfg.DataDir != want {
		t.Errorf("DataDir: want %q, got %q", want, cfg.DataDir)
	}
}

// --- ConfigPath() ---

func TestConfigPath_XDGOverride(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	got := config.ConfigPath()
	want := filepath.Join(tmp, "neocognito", "config.toml")
	if got != want {
		t.Errorf("ConfigPath: want %q, got %q", want, got)
	}
}

// --- Load() ---

func TestLoad_NoFileReturnsDefaults(t *testing.T) {
	isolate(t)
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load with no file: %v", err)
	}
	if cfg.StaleDays != 14 {
		t.Errorf("stale_days default: want 14, got %d", cfg.StaleDays)
	}
}

func TestLoad_ParsesValidTOML(t *testing.T) {
	isolate(t)
	dir := filepath.Dir(config.ConfigPath())
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	content := "stale_days = 30\ntheme = \"solarized\"\n"
	if err := os.WriteFile(config.ConfigPath(), []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.StaleDays != 30 {
		t.Errorf("stale_days: want 30, got %d", cfg.StaleDays)
	}
	if cfg.Theme != "solarized" {
		t.Errorf("theme: want solarized, got %q", cfg.Theme)
	}
}

func TestLoad_InvalidTOMLReturnsError(t *testing.T) {
	isolate(t)
	dir := filepath.Dir(config.ConfigPath())
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(config.ConfigPath(), []byte("stale_days = %%%"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := config.Load()
	if err == nil {
		t.Error("expected error for invalid TOML, got nil")
	}
}

func TestLoad_TildeExpansion(t *testing.T) {
	isolate(t)
	dir := filepath.Dir(config.ConfigPath())
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	content := "data_dir = \"~/myfolder\"\n"
	if err := os.WriteFile(config.ConfigPath(), []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if strings.HasPrefix(cfg.DataDir, "~") {
		t.Errorf("DataDir still has tilde: %q", cfg.DataDir)
	}
	home, _ := os.UserHomeDir()
	if !strings.HasPrefix(cfg.DataDir, home) {
		t.Errorf("DataDir does not start with home %q: %q", home, cfg.DataDir)
	}
}

// --- Save() round-trip ---

func TestSave_RoundTrip(t *testing.T) {
	isolate(t)
	cfg := config.Default()
	cfg.StaleDays = 99
	cfg.Theme = "gruvbox"

	if err := cfg.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := config.Load()
	if err != nil {
		t.Fatalf("Load after Save: %v", err)
	}
	if loaded.StaleDays != 99 {
		t.Errorf("stale_days: want 99, got %d", loaded.StaleDays)
	}
	if loaded.Theme != "gruvbox" {
		t.Errorf("theme: want gruvbox, got %q", loaded.Theme)
	}
}

// --- WriteDefault() ---

func TestWriteDefault_CreatesFile(t *testing.T) {
	isolate(t)
	if err := config.WriteDefault(); err != nil {
		t.Fatalf("WriteDefault: %v", err)
	}
	if _, err := os.Stat(config.ConfigPath()); os.IsNotExist(err) {
		t.Error("config file was not created")
	}
}

func TestWriteDefault_IdempotentDoesNotOverwrite(t *testing.T) {
	isolate(t)
	if err := config.WriteDefault(); err != nil {
		t.Fatalf("first WriteDefault: %v", err)
	}

	// Manually write a sentinel value
	if err := os.WriteFile(config.ConfigPath(), []byte("sentinel"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	if err := config.WriteDefault(); err != nil {
		t.Fatalf("second WriteDefault: %v", err)
	}

	data, err := os.ReadFile(config.ConfigPath())
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(data) != "sentinel" {
		t.Errorf("WriteDefault overwrote existing file; content: %q", data)
	}
}
