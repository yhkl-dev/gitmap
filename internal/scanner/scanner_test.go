package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanFindsRepos(t *testing.T) {
	dir := t.TempDir()
	repoPath := filepath.Join(dir, "my-repo")
	os.MkdirAll(filepath.Join(repoPath, ".git"), 0755)
	os.MkdirAll(filepath.Join(dir, "not-a-repo"), 0755)

	repos, err := Scan([]string{dir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repos) != 1 {
		t.Fatalf("expected 1 repo, got %d", len(repos))
	}
	if repos[0].Path != repoPath {
		t.Errorf("expected %s, got %s", repoPath, repos[0].Path)
	}
	if repos[0].Name != "my-repo" {
		t.Errorf("expected my-repo, got %s", repos[0].Name)
	}
}

func TestScanSkipsIgnored(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "node_modules", ".git"), 0755)
	os.MkdirAll(filepath.Join(dir, "real-repo", ".git"), 0755)
	os.MkdirAll(filepath.Join(dir, ".hidden", ".git"), 0755)

	repos, err := Scan([]string{dir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, r := range repos {
		base := filepath.Base(r.Path)
		if base == "node_modules" || base == ".hidden" {
			t.Errorf("should have skipped %s", base)
		}
	}
	if len(repos) != 1 {
		t.Fatalf("expected only 1 repo (real-repo), got %d", len(repos))
	}
}

func TestScanMultiplePaths(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()
	os.MkdirAll(filepath.Join(dir1, "repo-a", ".git"), 0755)
	os.MkdirAll(filepath.Join(dir2, "repo-b", ".git"), 0755)

	repos, err := Scan([]string{dir1, dir2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repos) != 2 {
		t.Fatalf("expected 2 repos across two paths, got %d", len(repos))
	}
}

func TestScanNoRepos(t *testing.T) {
	dir := t.TempDir()
	repos, err := Scan([]string{dir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repos) != 0 {
		t.Fatalf("expected 0 repos, got %d", len(repos))
	}
}
