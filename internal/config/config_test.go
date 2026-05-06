package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadDefaultsToReadOnlyAndSafeURLs(t *testing.T) {
	cfg, err := Load(Options{})
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg.Mode != "read-only" {
		t.Fatalf("Mode = %q, want read-only", cfg.Mode)
	}
	if cfg.LiveTradingEnabled {
		t.Fatal("live trading must default to disabled")
	}
	if cfg.RequestTimeout != 10*time.Second {
		t.Fatalf("RequestTimeout = %s, want 10s", cfg.RequestTimeout)
	}
}

func TestLoadReadsExplicitConfigFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte("mode: paper\npaper_state_path: /tmp/paper.json\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(Options{ConfigPath: path})
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg.Mode != "paper" {
		t.Fatalf("Mode = %q, want paper", cfg.Mode)
	}
	if cfg.PaperStatePath != "/tmp/paper.json" {
		t.Fatalf("PaperStatePath = %q", cfg.PaperStatePath)
	}
}
