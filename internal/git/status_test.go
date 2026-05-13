package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestDirtyDetection(t *testing.T) {
	dir := t.TempDir()
	runCmd(t, dir, "git", "init")
	runCmd(t, dir, "git", "config", "user.email", "test@test.com")
	runCmd(t, dir, "git", "config", "user.name", "test")
	os.WriteFile(filepath.Join(dir, "README.md"), []byte("hello"), 0644)
	runCmd(t, dir, "git", "add", ".")
	runCmd(t, dir, "git", "commit", "-m", "initial")

	s, err := Status(dir)
	if err != nil {
		t.Fatalf("Status error: %v", err)
	}
	if s.Branch != "master" && s.Branch != "main" {
		t.Errorf("expected master/main branch, got %s", s.Branch)
	}
	if s.Dirty {
		t.Error("repo should be clean after commit")
	}

	// Make it dirty
	os.WriteFile(filepath.Join(dir, "README.md"), []byte("modified"), 0644)
	s, err = Status(dir)
	if err != nil {
		t.Fatalf("Status error after modify: %v", err)
	}
	if !s.Dirty {
		t.Error("repo should be dirty after modification")
	}
}

func runCmd(t *testing.T, dir, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Logf("runCmd %s %v: %s (err: %v)", name, args, string(out), err)
	}
}
