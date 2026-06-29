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

func TestExcludeRepos(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := "scan_paths:\n  - /tmp/repos\nexclude_repos:\n  - node_modules\n  - vendor-*\n  - test-?\n"
	os.WriteFile(path, []byte(content), 0644)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.ExcludeRepos) != 3 {
		t.Fatalf("expected 3 exclude patterns, got %d", len(cfg.ExcludeRepos))
	}
}

func TestAuthorFromConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := "scan_paths:\n  - /tmp/repos\nauthor:\n  - alice@example.com\n  - bob@example.com\n"
	os.WriteFile(path, []byte(content), 0644)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Author) != 2 {
		t.Fatalf("expected 2 authors, got %d", len(cfg.Author))
	}
	if cfg.Author[0] != "alice@example.com" || cfg.Author[1] != "bob@example.com" {
		t.Fatalf("unexpected authors: %v", cfg.Author)
	}
}

func TestIsExcluded(t *testing.T) {
	cfg := &Config{
		ExcludeRepos: []string{"node_modules", "vendor-*", "test-?"},
	}
	tests := []struct {
		name     string
		expected bool
	}{
		{"node_modules", true},
		{"vendor-foo", true},
		{"vendor-", true},
		{"test-a", true},
		{"my-project", false},
		{"vendor", false},
		{"test-ab", false},
	}
	for _, tt := range tests {
		if got := cfg.IsExcluded(tt.name); got != tt.expected {
			t.Errorf("IsExcluded(%q) = %v, want %v", tt.name, got, tt.expected)
		}
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
