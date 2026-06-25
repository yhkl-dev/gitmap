# gitmap

> Terminal-based multi-repo manager with vim keybindings — scan local repos, check status at a glance, and act fast.

[中文版](README.zh-CN.md)

<img src="https://img.shields.io/badge/go-1.23%2B-00ADD8?logo=go" alt="Go 1.23+">

---

## Quick Start

```bash
# Install
go install github.com/yhkl-dev/gitmap/cmd/gitmap@latest

# Shell integration (add to ~/.bashrc / ~/.zshrc)
source <(gitmap --shell-init)

# Launch
gmap
```

`gmap` command: launches gitmap. Press `o` on a repo to print its path, or `Enter` to select and then `q` to quit — your shell will `cd` there automatically.

---

## Keys

### List View

| Key | Action |
|---|---|
| `j` / `k` / `↓` / `↑` | Move cursor |
| `5j` / `10k` | Move N lines (number prefix) |
| `ctrl+d` / `ctrl+u` | Half-page down/up |
| `ctrl+f` / `ctrl+b` | Full-page down/up |
| `gg` | Jump to top |
| `G` / `5G` | Jump to bottom / line N |
| `v` | Visual mode — select repos |
| `f` / `p` | Fetch / Pull (visual: batch selected) |
| `F` / `P` | Fetch / Pull all repos |
| `⏎` | Detail view |
| `/` | Fuzzy search (repo name / branch) |
| `n` / `N` | Next / Previous match (active filter only) |
| `s` | Cycle sort: default → name → dirty first |
| `o` | Print path and quit |
| `O` | Open in iTerm2 tab (macOS) |
| `c` | Open Claude Code in repo |
| `d` | Toggle diff preview (--stat) |
| `r` | Refresh repo list |
| `q` | Quit |

### Detail View

| Key | Action |
|---|---|
| `j` / `k` | Scroll content |
| `ctrl+d` / `ctrl+u` | Half-page scroll |
| `ctrl+f` / `ctrl+b` | Full-page scroll |
| `gg` / `G` | Scroll to top / bottom |
| `a` | Stash apply |
| `A` | Stash pop |
| `D` | Stash drop |
| `B` | Branch checkout (fuzzy search) |
| `b` | Open PR in browser |
| `d` | Toggle diff (full patch) |
| `o` | Print path and quit |
| `O` | Open in iTerm2 tab |
| `r` | Refresh detail |
| `esc` / `q` | Back to list |

---

## Status Indicators

| Indicator | Meaning |
|---|---|
| `●` (yellow) | Working tree dirty |
| `○` (green) | Working tree clean |
| `⚡` (red) | Merge / rebase / cherry-pick conflict |
| `↑N` | N commits ahead of upstream |
| `↓N` | N commits behind upstream |
| `≡N` | N stashes |
| `?N` | N untracked files |
| `[dirty]` | Uncommitted changes |
| `[conflict]` | Conflict state |

---

## Config

Create `~/.config/gitmap/config.yaml`:

```yaml
scan_paths:
  - ~/projects
  - ~/work
  - /mnt/data/repos

auto_fetch: true   # auto fetch all repos on startup
```

Each path is scanned recursively for `.git` directories. Hidden folders (`.venv`, `.cache`, etc.) and package directories (`node_modules`, `vendor`) are skipped automatically.

If no config file exists, `~/projects` is scanned by default.

---

## Development

### Prerequisites

- **Go 1.23+** — [download](https://go.dev/dl/)
- Terminal with true color support (iTerm2, Kitty, WezTerm, Windows Terminal, etc.)

### Clone & Build

```bash
git clone git@github.com:yhkl-dev/gitmap.git
cd gitmap

make build   # → ./gitmap
make run     # Build + run
```

### Makefile

```bash
make build          # Build the binary
make run            # Build and run
make test           # Run all tests
make test-race      # Run tests with race detector
make cover          # Coverage report → coverage.html
make lint           # Static analysis (go vet)
make fmt            # Format code
make tidy           # Tidy module deps
make clean          # Remove build artifacts
make install        # Install to $GOPATH/bin
make dev            # Build, test, run (all-in-one)
make release        # Create release with goreleaser
make release-snapshot  # Test release without publishing
make help           # Show all targets
```

### Project Structure

```
gitmap/
├── cmd/gitmap/           # Entry point — Bubble Tea TUI
│   ├── main.go           #   Model, types, wiring, helpers
│   ├── commands.go       #   All tea.Cmd async functions
│   ├── list.go           #   List page: keys, view, fuzzy search
│   ├── detail.go         #   Detail page: keys, scrollable view
│   └── list_test.go      #   Tests
├── internal/
│   ├── config/           # YAML config parsing
│   │   ├── config.go
│   │   └── config_test.go
│   ├── git/              # Git operations (go-git + CLI)
│   │   ├── status.go     #   Status, diff, stash, branch, remote
│   │   ├── cache.go      #   Mtime-based status cache
│   │   ├── status_test.go
│   │   └── cache_test.go
│   └── scanner/          # Recursive .git discovery
│       ├── scanner.go
│       └── scanner_test.go
├── .github/workflows/    # CI: goreleaser on tag push
├── .goreleaser.yml       # Multi-platform release config
├── Makefile
├── go.mod
└── README.md
```

### Tech Stack

| Layer | Library |
|---|---|
| TUI framework | [Bubble Tea](https://github.com/charmbracelet/bubbletea) |
| Styling | [Lip Gloss](https://github.com/charmbracelet/lipgloss) |
| Git parsing | [go-git](https://github.com/go-git/go-git) v5 (pure Go) |
| Git CLI | `git` binary (fetch, pull, stash, diff, checkout) |
| Config | `gopkg.in/yaml.v3` |

---

## License

MIT
