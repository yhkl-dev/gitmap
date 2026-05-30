# gitmap

> Terminal-based multi-repo manager with vim keybindings — scan local repos, check status at a glance, and act fast.

> 基于终端的 Vim 键位多仓库管理器 — 扫描本地仓库，一目了然查看状态，快速操作。

<img src="https://img.shields.io/badge/go-1.23%2B-00ADD8?logo=go" alt="Go 1.23+">

---

## Quick Start · 快速开始

```bash
# Install · 安装
go install github.com/yhkl-dev/gitmap/cmd/gitmap@latest

# Shell integration · Shell 集成 (add to ~/.bashrc / ~/.zshrc)
source <(gitmap --shell-init)

# Launch · 启动
gmap
```

`gmap` command: launches gitmap. Press `o` on a repo to `cd` into it, or `Enter` to select and then `q` to quit — your shell will `cd` there automatically.

`gmap` 命令：启动 gitmap。在仓库上按 `o` 输出路径，或 `Enter` 选中后 `q` 退出 — shell 自动 `cd` 到该目录。

---

## Keys · 键位

### List View · 列表视图

| Key · 键位 | Action · 操作 |
|---|---|
| `j` / `k` / `↓` / `↑` | Move cursor · 移动光标 |
| `5j` / `10k` | Move N lines · 移动 N 行 (number prefix · 数字前缀) |
| `ctrl+d` / `ctrl+u` | Half-page down/up · 半页上下 |
| `ctrl+f` / `ctrl+b` | Full-page down/up · 整页上下 |
| `gg` | Jump to top · 跳到顶部 |
| `G` / `5G` | Jump to bottom / line N · 跳到底部 / 第 N 行 |
| `v` | Visual mode — select repos · 可视模式 — 多选仓库 |
| `f` / `p` | Fetch / Pull (visual: batch selected) · 拉取 (可视模式：批量) |
| `F` / `P` | Fetch / Pull all repos · 拉取全部仓库 |
| `⏎` | Detail view · 进入详情 |
| `/` | Fuzzy search (repo name / branch) · 模糊搜索 (仓库名 / 分支) |
| `n` / `N` | Next / Previous match · 下一个 / 上一个匹配 (active filter only) |
| `s` | Cycle sort: default → name → dirty first · 切换排序 |
| `o` | Print path and quit · 输出路径并退出 |
| `O` | Open in iTerm2 tab · 在 iTerm2 新标签打开 (macOS) |
| `c` | Open Claude Code in repo · 在仓库中启动 Claude Code |
| `d` | Toggle diff preview (--stat) · 切换 diff 预览 |
| `r` | Refresh repo list · 刷新仓库列表 |
| `q` | Quit · 退出 |

### Detail View · 详情视图

| Key · 键位 | Action · 操作 |
|---|---|
| `j` / `k` | Scroll content · 滚动内容 |
| `ctrl+d` / `ctrl+u` | Half-page scroll · 半页滚动 |
| `ctrl+f` / `ctrl+b` | Full-page scroll · 整页滚动 |
| `gg` / `G` | Scroll to top / bottom · 滚到顶部 / 底部 |
| `a` | Stash apply · 应用最近 stash |
| `A` | Stash pop · 弹出最近 stash |
| `D` | Stash drop · 删除最近 stash |
| `B` | Branch checkout (fuzzy search) · 分支切换 (模糊搜索) |
| `b` | Open PR in browser · 浏览器打开 PR |
| `d` | Toggle diff (full patch) · 切换 diff (完整补丁) |
| `o` | Print path and quit · 输出路径并退出 |
| `O` | Open in iTerm2 tab · 在 iTerm2 新标签打开 |
| `r` | Refresh detail · 刷新详情 |
| `esc` / `q` | Back to list · 返回列表 |

---

## Status Indicators · 状态指示

| Indicator · 指示 | Meaning · 含义 |
|---|---|
| `●` (yellow) | Working tree dirty · 工作区有改动 |
| `○` (green) | Working tree clean · 工作区干净 |
| `⚡` (red) | Merge / rebase / cherry-pick conflict · 合并 / 变基冲突 |
| `↑N` | N commits ahead of upstream · 领先远程 N 个提交 |
| `↓N` | N commits behind upstream · 落后远程 N 个提交 |
| `≡N` | N stashes · N 个储藏 |
| `?N` | N untracked files · N 个未跟踪文件 |
| `[dirty]` | Uncommitted changes · 未提交的修改 |
| `[conflict]` | Conflict state · 冲突状态 |

---

## Config · 配置

Create `~/.config/gitmap/config.yaml`:

创建 `~/.config/gitmap/config.yaml`：

```yaml
scan_paths:
  - ~/projects
  - ~/work
  - /mnt/data/repos

auto_fetch: true   # auto fetch all repos on startup · 启动时自动 fetch
```

Each path is scanned recursively for `.git` directories. Hidden folders (`.venv`, `.cache`, etc.) and package directories (`node_modules`, `vendor`) are skipped automatically.

每个路径递归扫描 `.git` 目录。隐藏文件夹和包目录自动跳过。

If no config file exists, `~/projects` is scanned by default.

未配置时默认扫描 `~/projects`。

---

## Development · 开发

### Prerequisites · 环境

- **Go 1.23+** — [download](https://go.dev/dl/)
- Terminal with true color support (iTerm2, Kitty, WezTerm, Windows Terminal, etc.)

### Clone & Build · 克隆构建

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

### Project Structure · 项目结构

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

### Tech Stack · 技术栈

| Layer 层 | Library 库 |
|---|---|
| TUI framework | [Bubble Tea](https://github.com/charmbracelet/bubbletea) |
| Styling | [Lip Gloss](https://github.com/charmbracelet/lipgloss) |
| Git parsing | [go-git](https://github.com/go-git/go-git) v5 (pure Go) |
| Git CLI | `git` binary (fetch, pull, stash, diff, checkout) |
| Config | `gopkg.in/yaml.v3` |

---

## License

MIT
