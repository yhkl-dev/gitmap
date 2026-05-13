# gitmap

Terminal-based git repository navigator. Browse all your local repos, see their
status at a glance, and jump in with a single keystroke.

## Install

```bash
go install github.com/yhkl-dev/gitmap/cmd/gitmap@latest
```

## Quick Start

Add this to your `~/.bashrc` or `~/.zshrc`:

```bash
source <(gitmap --shell-init)
```

Then run:

```bash
gmap
```

**Keys:**

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
