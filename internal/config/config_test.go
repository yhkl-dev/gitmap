package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultScanPaths(t *testing.T) {
	cfg := Default()
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, "projects")
	if len(cfg.ScanPaths) != 1 || cfg.ScanPaths[0] != expected {
		t.Errorf("expected [%s], got %v", expected, cfg.ScanPaths)
	}
}

func TestLoadFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := "scan_paths:\n  - /tmp/repos\n  - ~/code\n"
	os.WriteFile(path, []byte(content), 0644)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.ScanPaths) != 2 {
		t.Fatalf("expected 2 scan paths, got %d", len(cfg.ScanPaths))
	}
	if cfg.ScanPaths[0] != "/tmp/repos" {
		t.Errorf("expected /tmp/repos, got %s", cfg.ScanPaths[0])
	}
}

func TestExpandTilde(t *testing.T) {
	result := expandTilde("~/code")
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, "code")
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestLoadDefaultsWhenEmpty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := "scan_paths: []\n"
	os.WriteFile(path, []byte(content), 0644)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.ScanPaths) != 1 {
		t.Fatalf("expected 1 default scan path, got %d", len(cfg.ScanPaths))
	}
}

func TestLoadFileNotFound(t *testing.T) {
	cfg, err := Load("/nonexistent/path/config.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.ScanPaths) != 1 {
		t.Fatalf("expected 1 default scan path, got %d", len(cfg.ScanPaths))
	}
}
