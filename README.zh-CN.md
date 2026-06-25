# gitmap

> 基于终端的 Vim 键位多仓库管理器 — 扫描本地仓库，一目了然查看状态，快速操作。

[English](README.md)

<img src="https://img.shields.io/badge/go-1.23%2B-00ADD8?logo=go" alt="Go 1.23+">

---

## 快速开始

```bash
# 安装
go install github.com/yhkl-dev/gitmap/cmd/gitmap@latest

# Shell 集成 (添加到 ~/.bashrc / ~/.zshrc)
source <(gitmap --shell-init)

# 启动
gmap
```

`gmap` 命令：启动 gitmap。在仓库上按 `o` 输出路径，或 `Enter` 选中后 `q` 退出 — shell 自动 `cd` 到该目录。

---

## 键位

### 列表视图

| 键位 | 操作 |
|---|---|
| `j` / `k` / `↓` / `↑` | 移动光标 |
| `5j` / `10k` | 移动 N 行 (数字前缀) |
| `ctrl+d` / `ctrl+u` | 半页上下 |
| `ctrl+f` / `ctrl+b` | 整页上下 |
| `gg` | 跳到顶部 |
| `G` / `5G` | 跳到底部 / 第 N 行 |
| `v` | 可视模式 — 多选仓库 |
| `f` / `p` | 拉取 (可视模式：批量) |
| `F` / `P` | 拉取全部仓库 |
| `⏎` | 进入详情 |
| `/` | 模糊搜索 (仓库名 / 分支) |
| `n` / `N` | 下一个 / 上一个匹配 (仅活跃筛选) |
| `s` | 切换排序：默认 → 名称 → 脏优先 |
| `o` | 输出路径并退出 |
| `O` | 在 iTerm2 新标签打开 (macOS) |
| `c` | 在仓库中启动 Claude Code |
| `d` | 切换 diff 预览 (--stat) |
| `r` | 刷新仓库列表 |
| `q` | 退出 |

### 详情视图

| 键位 | 操作 |
|---|---|
| `j` / `k` | 滚动内容 |
| `ctrl+d` / `ctrl+u` | 半页滚动 |
| `ctrl+f` / `ctrl+b` | 整页滚动 |
| `gg` / `G` | 滚到顶部 / 底部 |
| `a` | 应用最近 stash |
| `A` | 弹出最近 stash |
| `D` | 删除最近 stash |
| `B` | 分支切换 (模糊搜索) |
| `b` | 浏览器打开 PR |
| `d` | 切换 diff (完整补丁) |
| `o` | 输出路径并退出 |
| `O` | 在 iTerm2 新标签打开 |
| `r` | 刷新详情 |
| `esc` / `q` | 返回列表 |

---

## 状态指示

| 指示 | 含义 |
|---|---|
| `●` (黄色) | 工作区有改动 |
| `○` (绿色) | 工作区干净 |
| `⚡` (红色) | 合并 / 变基冲突 |
| `↑N` | 领先远程 N 个提交 |
| `↓N` | 落后远程 N 个提交 |
| `≡N` | N 个储藏 |
| `?N` | N 个未跟踪文件 |
| `[dirty]` | 未提交的修改 |
| `[conflict]` | 冲突状态 |

---

## 配置

创建 `~/.config/gitmap/config.yaml`：

```yaml
scan_paths:
  - ~/projects
  - ~/work
  - /mnt/data/repos

auto_fetch: true   # 启动时自动 fetch
```

每个路径递归扫描 `.git` 目录。隐藏文件夹（`.venv`、`.cache` 等）和包目录（`node_modules`、`vendor`）自动跳过。

未配置时默认扫描 `~/projects`。

---

## 开发

### 环境

- **Go 1.23+** — [下载](https://go.dev/dl/)
- 支持真彩色的终端（iTerm2、Kitty、WezTerm、Windows Terminal 等）

### 克隆构建

```bash
git clone git@github.com:yhkl-dev/gitmap.git
cd gitmap

make build   # → ./gitmap
make run     # 构建并运行
```

### Makefile

```bash
make build          # 构建二进制文件
make run            # 构建并运行
make test           # 运行所有测试
make test-race      # 带竞态检测运行测试
make cover          # 覆盖率报告 → coverage.html
make lint           # 静态分析 (go vet)
make fmt            # 格式化代码
make tidy           # 整理模块依赖
make clean          # 删除构建产物
make install        # 安装到 $GOPATH/bin
make dev            # 构建、测试、运行 (一键)
make release        # 使用 goreleaser 创建发布
make release-snapshot  # 测试发布（不实际发布）
make help           # 显示所有目标
```

### 项目结构

```
gitmap/
├── cmd/gitmap/           # 入口 — Bubble Tea TUI
│   ├── main.go           #   模型、类型、组装、辅助函数
│   ├── commands.go       #   所有 tea.Cmd 异步函数
│   ├── list.go           #   列表页：键位、视图、模糊搜索
│   ├── detail.go         #   详情页：键位、可滚动视图
│   └── list_test.go      #   测试
├── internal/
│   ├── config/           # YAML 配置解析
│   │   ├── config.go
│   │   └── config_test.go
│   ├── git/              # Git 操作 (go-git + CLI)
│   │   ├── status.go     #   状态、diff、stash、分支、远程
│   │   ├── cache.go      #   基于 mtime 的状态缓存
│   │   ├── status_test.go
│   │   └── cache_test.go
│   └── scanner/          # 递归 .git 发现
│       ├── scanner.go
│       └── scanner_test.go
├── .github/workflows/    # CI: 标签推送触发 goreleaser
├── .goreleaser.yml       # 多平台发布配置
├── Makefile
├── go.mod
└── README.md
```

### 技术栈

| 层 | 库 |
|---|---|
| TUI 框架 | [Bubble Tea](https://github.com/charmbracelet/bubbletea) |
| 样式 | [Lip Gloss](https://github.com/charmbracelet/lipgloss) |
| Git 解析 | [go-git](https://github.com/go-git/go-git) v5 (纯 Go) |
| Git CLI | `git` 命令 (fetch, pull, stash, diff, checkout) |
| 配置 | `gopkg.in/yaml.v3` |

---

## License

MIT
