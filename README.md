# gitmap

> Terminal-based multi-repo navigator — browse all your local repos, see
> their status at a glance, and jump in with a single keystroke.

<img src="https://img.shields.io/badge/go-1.23%2B-00ADD8?logo=go" alt="Go 1.23+">

## Quick Start

```bash
# Install
go install github.com/yhkl-dev/gitmap/cmd/gitmap@latest

# Add shell integration to ~/.bashrc or ~/.zshrc
source <(gitmap --shell-init)

# Launch
gmap
```

## Keys

| Key | Action |
|-----|--------|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `Enter` | `cd` into selected repo |
| `d` | Toggle diff preview |
| `r` | Refresh repo list |
| `q` | Quit |

## Config

Create `~/.config/gitmap/config.yaml`:

```yaml
scan_paths:
  - ~/projects
  - ~/work
  - /mnt/data/repos
```

Each path is scanned recursively for `.git` directories. Hidden folders
(`.venv`, `.cache`, etc.) and package directories (`node_modules`, `vendor`)
are skipped automatically.

If no config file exists, `~/projects` is scanned by default.

---

## Development

### Prerequisites

- **Go 1.23+** — [download](https://go.dev/dl/)
- A terminal with true color support (iTerm2, Kitty, WezTerm, Windows Terminal, etc.)

### Clone & Build

```bash
git clone git@github.com:yhkl-dev/gitmap.git
cd gitmap

# Build
make build
# → ./gitmap

# Build + run
make run
```

### Makefile

```bash
make build       # Build the binary
make run         # Build and run
make test        # Run all tests
make test-race   # Run tests with race detector
make cover       # Coverage report → coverage.html
make lint        # Static analysis (go vet)
make fmt         # Format code
make tidy        # Tidy module deps
make clean       # Remove build artifacts
make install     # Install to $GOPATH/bin
make dev         # Build, test, run (all-in-one)
make help        # Show all targets
```

### Project Structure

```
gitmap/
├── cmd/gitmap/          # Entry point — Bubble Tea TUI
│   └── main.go
├── internal/
│   ├── config/          # YAML config parsing & validation
│   │   ├── config.go
│   │   └── config_test.go
│   ├── git/             # Git status via go-git
│   │   ├── status.go
│   │   └── status_test.go
│   └── scanner/         # Recursive .git dir discovery
│       ├── scanner.go
│       └── scanner_test.go
├── scripts/             # Shell integration helpers
├── docs/plans/          # Implementation plans
├── Makefile
├── go.mod
└── README.md
```

### Tech Stack

| Layer | Library |
|-------|---------|
| TUI framework | [Bubble Tea](https://github.com/charmbracelet/bubbletea) |
| Styling | [Lip Gloss](https://github.com/charmbracelet/lipgloss) |
| Git parsing | [go-git](https://github.com/go-git/go-git) v5 (pure Go, no git binary required) |
| Config | `gopkg.in/yaml.v3` |

## License

MIT
