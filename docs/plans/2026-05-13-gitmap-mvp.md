# Gitmap MVP Implementation Plan

> **For Hermes:** Use subagent-driven-development skill to implement this plan task-by-task.

**Goal:** Build a terminal TUI that scans local directories for git repositories, shows their status at a glance, and lets the user jump into any repo with one keystroke.

**Architecture:** Go CLI app using Bubble Tea for the TUI, go-git for pure-Go git status queries (no shelling out to `git`), and a YAML config file for scan paths. Three-panel layout: left = repo list with dirty/clean indicators, middle = branches, bottom = recent commits.

**Tech Stack:** Go 1.22+, Bubble Tea, Lip Gloss, go-git, YAML (viper or gopkg.in/yaml.v3).

---

### Task 1: Project Scaffolding

**Objective:** Set up directory structure, install dependencies, create entry point that compiles.

**Files:**
- Create: `cmd/gitmap/main.go`
- Create: `internal/tui/` (empty dir)
- Create: `internal/git/` (empty dir)
- Create: `internal/scanner/` (empty dir)
- Create: `internal/config/` (empty dir)

**Step 1: Create directory structure**

```bash
mkdir -p cmd/gitmap internal/{tui,git,scanner,config}
```

**Step 2: Create minimal main.go**

```go
package main

import "fmt"

func main() {
	fmt.Println("gitmap - git repository navigator")
}
```

**Step 3: Fetch dependencies**

```bash
go get github.com/charmbracelet/bubbletea
go get github.com/charmbracelet/lipgloss
go get github.com/charmbracelet/bubbles
go get github.com/go-git/go-git/v5
go get gopkg.in/yaml.v3
```

**Step 4: Verify build**

```bash
go build ./cmd/gitmap/
./gitmap
```

Expect: "gitmap - git repository navigator"

**Step 5: Commit**

```bash
git add -A
git commit -m "feat: scaffold project structure and dependencies"
```

---

### Task 2: Config Module — Scan Paths

**Objective:** Read YAML config file with scan paths. Default to `~/projects` if no config exists.

**Files:**
- Create: `internal/config/config.go`
- Create: `internal/config/config_test.go`

**Config format** (`~/.config/gitmap/config.yaml`):

```yaml
scan_paths:
  - ~/projects
  - /mnt/data/repos
```

**Step 1: Write failing test**

```go
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
}

func TestExpandTilde(t *testing.T) {
	result := expandTilde("~/code")
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, "code")
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}
```

**Step 2: Run test** → FAIL (package doesn't exist yet)

**Step 3: Implement config.go**

```go
package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ScanPaths []string `yaml:"scan_paths"`
}

func Default() *Config {
	home, _ := os.UserHomeDir()
	return &Config{
		ScanPaths: []string{filepath.Join(home, "projects")},
	}
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Default(), nil
	}
	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	for i, p := range cfg.ScanPaths {
		cfg.ScanPaths[i] = expandTilde(p)
	}
	return cfg, nil
}

func expandTilde(path string) string {
	if len(path) > 0 && path[0] == '~' {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[1:])
	}
	return path
}
```

**Step 4: Run test** → PASS

**Step 5: Commit**

```bash
git add -A
git commit -m "feat: add config module with YAML scan paths"
```

---

### Task 3: Scanner — Find Git Repos

**Objective:** Walk scan paths, detect `.git` directories, return list of repo paths.

**Files:**
- Create: `internal/scanner/scanner.go`
- Create: `internal/scanner/scanner_test.go`

**Step 1: Write failing test**

```go
package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanFindsRepos(t *testing.T) {
	dir := t.TempDir()
	// Create a repo-like dir
	repoPath := filepath.Join(dir, "my-repo")
	os.MkdirAll(filepath.Join(repoPath, ".git"), 0755)
	// Create a non-repo dir
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
}

func TestScanSkipsIgnored(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "node_modules", ".git"), 0755)
	os.MkdirAll(filepath.Join(dir, "real-repo", ".git"), 0755)

	repos, err := Scan([]string{dir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should skip node_modules
	for _, r := range repos {
		if filepath.Base(r.Path) == "node_modules" {
			t.Error("should have skipped node_modules")
		}
	}
}
```

**Step 2: Run test** → FAIL

**Step 3: Implement scanner.go**

```go
package scanner

import (
	"os"
	"path/filepath"
	"strings"
)

type Repo struct {
	Path string
	Name string
}

var ignoredDirs = map[string]bool{
	"node_modules": true,
	".venv":        true,
	"venv":         true,
	"__pycache__":  true,
	".cache":       true,
	"vendor":       true,
}

func Scan(paths []string) ([]Repo, error) {
	var repos []Repo
	seen := map[string]bool{}

	for _, root := range paths {
		filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // skip inaccessible paths
			}
			if !info.IsDir() {
				return nil
			}
			name := filepath.Base(path)
			if ignoredDirs[name] || strings.HasPrefix(name, ".") && name != ".git" {
				return filepath.SkipDir
			}
			if name == ".git" {
				parent := filepath.Dir(path)
				if !seen[parent] {
					seen[parent] = true
					repos = append(repos, Repo{
						Path: parent,
						Name: filepath.Base(parent),
					})
				}
				return filepath.SkipDir
			}
			return nil
		})
	}
	return repos, nil
}
```

**Step 4: Run test** → PASS

**Step 5: Commit**

```bash
git add -A
git commit -m "feat: add scanner to discover git repos"
```

---

### Task 4: Git Module — Read Status via go-git

**Objective:** Open a git repo with go-git, read working tree status (clean/dirty/modified), and current branch name.

**Files:**
- Create: `internal/git/status.go`
- Create: `internal/git/status_test.go`

**Step 1: Write failing test**

```go
package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestDirtyDetection(t *testing.T) {
	dir := t.TempDir()
	// Init a real git repo for testing
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
		t.Error("repo should be clean")
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
	cmd.Run()
}
```

**Step 2: Run test** → FAIL

**Step 3: Implement status.go**

```go
package git

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

type RepoStatus struct {
	Path   string
	Name   string
	Branch string
	Dirty  bool
	Ahead  int
	Behind int
}

func Status(path string) (*RepoStatus, error) {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return nil, err
	}

	head, err := repo.Head()
	if err != nil {
		// No commits yet — treat as clean, branch = ""
		return &RepoStatus{
			Path:  path,
			Dirty: false,
		}, nil
	}

	branch := head.Name().Short()

	wt, err := repo.Worktree()
	if err != nil {
		return nil, err
	}

	ws, err := wt.Status()
	if err != nil {
		return nil, err
	}

	return &RepoStatus{
		Path:   path,
		Branch: branch,
		Dirty:  !ws.IsClean(),
	}, nil
}
```

> Note: `Ahead`/`Behind` count requires remote fetch via go-git, deferred to later task.

**Step 4: Run test** → PASS

**Step 5: Commit**

```bash
git add -A
git commit -m "feat: add git status reader via go-git"
```

---

### Task 5: TUI — Repo List View

**Objective:** Build the Bubble Tea TUI that shows repo list with dirty/clean indicators. Keyboard navigation (j/k). Enter exits with selected repo path printed.

**Files:**
- Modify: `cmd/gitmap/main.go`
- Create: `internal/tui/model.go`

**Step 1: Write an integration test (manual verification)**

Create `cmd/gitmap/main.go`:

```go
package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"gitmap/internal/config"
	"gitmap/internal/git"
	"gitmap/internal/scanner"
)

var (
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Bold(true)
	dirtyMark     = lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Render("●")
	cleanMark     = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render("○")
)

type model struct {
	repos   []git.RepoStatus
	cursor  int
	quitting bool
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "j", "down":
			if m.cursor < len(m.repos)-1 {
				m.cursor++
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "enter":
			if len(m.repos) > 0 {
				fmt.Println(m.repos[m.cursor].Path)
			}
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.quitting {
		return ""
	}
	s := "gitmap — repos\n\n"
	for i, r := range m.repos {
		mark := cleanMark
		if r.Dirty {
			mark = dirtyMark
		}
		line := fmt.Sprintf("  %s  %s", mark, r.Name)
		if i == m.cursor {
			line = selectedStyle.Render(line)
		}
		s += line + "\n"
	}
	s += "\n↑/↓ navigate  ⏎ jump  q quit\n"
	return s
}

func main() {
	cfg := config.Default()
	cfgPath := os.ExpandEnv("$HOME/.config/gitmap/config.yaml")
	if _, err := os.Stat(cfgPath); err == nil {
		cfg, _ = config.Load(cfgPath)
	}

	repos, err := scanner.Scan(cfg.ScanPaths)
	if err != nil {
		fmt.Fprintf(os.Stderr, "scan error: %v\n", err)
		os.Exit(1)
	}

	var statuses []git.RepoStatus
	for _, r := range repos {
		s, err := git.Status(r.Path)
		if err != nil {
			s = &git.RepoStatus{Path: r.Path, Name: r.Name}
		}
		statuses = append(statuses, *s)
	}

	p := tea.NewProgram(model{repos: statuses})
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
```

**Step 2: Build and verify**

```bash
go build ./cmd/gitmap/
./gitmap
```

Manual check: TUI shows repos from `~/projects`, j/k navigate, q quits, Enter prints path and exits.

**Step 3: Commit**

```bash
git add -A
git commit -m "feat: add TUI repo list with dirty/clean indicators"
```

---

### Task 6: Shell Jump Integration

**Objective:** Print the selected repo path to stdout so a shell wrapper can `cd` into it. Add a shell alias/function for `gmap` command.

**Files:**
- Modify: `cmd/gitmap/main.go` (already outputs path on Enter — verify this works)
- Create: `scripts/gitmap.sh` (shell integration)

**Step 1: Create shell wrapper**

`scripts/gitmap.sh`:

```bash
gmap() {
    local target
    target=$(gitmap 2>/dev/null)
    if [ -n "$target" ] && [ -d "$target" ]; then
        cd "$target" || return
        echo "→ $target"
    fi
}
```

**Step 2: Verify the workflow**

```bash
# Source the wrapper
source scripts/gitmap.sh
# Run it
gmap
# Select a repo with Enter
# Should cd into that directory
```

**Step 3: Add README.md with usage**

```markdown
# gitmap

Terminal-based git repository navigator.

## Install

```bash
go install github.com/yhkl-dev/gitmap/cmd/gitmap@latest
```

## Usage

```bash
# Shell integration (add to .bashrc / .zshrc)
source <(gitmap --shell-init)

# Then:
gmap
```

## Config

`~/.config/gitmap/config.yaml`:

```yaml
scan_paths:
  - ~/projects
  - /mnt/data/repos
```
```

**Step 4: Commit**

```bash
git add -A
git commit -m "feat: add shell wrapper and README"
```

---

### Task 7: Final Cleanup — `git diff` Preview & Polish

**Objective:** Add `d` key to show git diff summary (file list) in a bottom panel. Polish colors and help text.

**Files:**
- Modify: `cmd/gitmap/main.go`

**Step 1: Add diff preview to model**

Extend the `model` struct:

```go
type model struct {
	repos     []git.RepoStatus
	cursor    int
	diff      string // diff output for selected repo
	showDiff  bool
	quitting  bool
}
```

**Step 2: Add `d` key handler** — on `d` keypress, run `go-git` diff and populate `model.diff`. Show it in the view below the repo list.

**Step 3: Commit**

```bash
git add -A
git commit -m "feat: add git diff preview panel"
```

---

## Milestones Summary

| Task | What | Est. Time |
|------|------|-----------|
| 1 | Scaffolding | 5 min |
| 2 | Config module | 10 min |
| 3 | Scanner | 15 min |
| 4 | Git status | 15 min |
| 5 | TUI view | 20 min |
| 6 | Shell integration | 10 min |
| 7 | Polish | 15 min |

**Total MVP:** ~90 min of focused work.

## What's NOT in MVP (v2 ideas)

- Branch/worktree panel (middle column)
- Commit log panel (bottom)
- Ahead/behind counts (requires remote fetch)
- tmux integration
- gh CLI PR status
- Batch operations on multiple repos
- Fuzzy filter search
- Stash management
