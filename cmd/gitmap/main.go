package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/yhkl-dev/gitmap/internal/config"
	gitpkg "github.com/yhkl-dev/gitmap/internal/git"
)

var (
	green   = lipgloss.Color("2")
	yellow  = lipgloss.Color("3")
	grey    = lipgloss.Color("8")

	muted        = lipgloss.NewStyle().Foreground(grey)
	cursorStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Bold(true)
	dirtyTag     = lipgloss.NewStyle().Foreground(yellow).Render("[dirty]")
	cleanDot     = lipgloss.NewStyle().Foreground(green).Render("○")
	dirtyDot     = lipgloss.NewStyle().Foreground(yellow).Render("●")
	bold         = lipgloss.NewStyle().Bold(true)
	aheadStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Render
	behindStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("5")).Render
	stashStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Render
	untrackStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Render
)

const (
	pageList    = iota
	pageDetail
	pageHeatmap
)

const (
	sortDefault = iota
	sortByName
	sortByDirty
)

var sortNames = map[int]string{
	sortDefault: "default",
	sortByName:  "name",
	sortByDirty: "dirty first",
}

type model struct {
	page int

	allRepos     []gitpkg.RepoStatus
	cursor       int
	scrollOffset int
	filter       string
	filtering    bool
	sortMode     int

	detailRepo     *gitpkg.RepoStatus
	detailDiff     string
	detailLog      string
	detailBranches string
	detailStashes      string
	detailScroll       int
	detailStashResult  string
	prsOutput          string
	prsLoading         bool
	showDiff           bool

	branchSelect  bool
	branchFilter  string

	loading       bool
	fetching      bool
	fetchProgress string
	pulling       bool
	pullProgress  string
	scanPaths    []string
	excludeRepos []string
	author       string
	autoFetch    bool
	initDone     bool
	errorCount int // -1 = scan failed, >=0 = per-repo errors

	prefixCount int
	lastKey     string

	visualMode  bool
	visualStart int

	heatmapCommits    map[string]int
	heatmapLines      map[string]int
	heatmapLoading    bool
	heatmapTotal      int
	heatmapLinesTotal int
	heatmapRepos      int
	heatmapLoadedAt   time.Time

	lastActivity time.Time

	width    int
	height   int
	quitting bool
}

// ── message types ────────────────────────────────────────────────

type prsLoadedMsg string

type reposLoadedMsg struct {
	repos  []gitpkg.RepoStatus
	errors int // -1 = scan failed, >=0 = per-repo error count
}

type fetchDoneMsg struct{ single bool }
type pullDoneMsg struct{ single bool }
type idleTickMsg struct{}

func idleTickCmd() tea.Cmd {
	return tea.Tick(1*time.Second, func(t time.Time) tea.Msg {
		return idleTickMsg{}
	})
}

// ── bubble tea ──────────────────────────────────────────────────

func (m model) Init() tea.Cmd {
	return tea.Batch(loadReposCmd(m.scanPaths, m.excludeRepos), idleTickCmd())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case reposLoadedMsg:
		m.allRepos = msg.repos
		m.errorCount = msg.errors
		m.loading = false
		m.cursor = clamp(m.cursor, 0, len(msg.repos)-1)
		if m.autoFetch && !m.initDone {
			m.initDone = true
			m.fetching = true
			m.fetchProgress = "auto-fetching..."
			return m, batchFetchCmd(m.allRepos)
		}

	case prsLoadedMsg:
		m.prsLoading = false
		m.prsOutput = string(msg)

	case fetchDoneMsg:
		m.fetching = false
		if msg.single {
			m.fetchProgress = "fetched"
		} else {
			m.fetchProgress = "fetched all"
		}
		m.loading = true
		return m, loadReposCmd(m.scanPaths, m.excludeRepos)

	case pullDoneMsg:
		m.pulling = false
		if msg.single {
			m.pullProgress = "pulled"
		} else {
			m.pullProgress = "pulled all"
		}
		m.loading = true
		return m, loadReposCmd(m.scanPaths, m.excludeRepos)

	case heatmapLoadedMsg:
		m.heatmapCommits = msg.commits
		m.heatmapLines = msg.lines
		m.heatmapTotal = msg.total
		m.heatmapLinesTotal = msg.linesTotal
		m.heatmapRepos = msg.repos
		m.heatmapLoading = false
		m.heatmapLoadedAt = time.Now()

	case idleTickMsg:
		if m.page == pageList && time.Since(m.lastActivity) > 60*time.Second {
			m.page = pageHeatmap
			if !m.heatmapFresh() {
				m.heatmapLoading = true
				m.heatmapCommits = nil
				m.heatmapLines = nil
				return m, tea.Batch(loadHeatmapCmd(m.allRepos, m.author), idleTickCmd())
			}
		}
		return m, idleTickCmd()

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.lastActivity = time.Now()

	if msg.String() == "ctrl+c" {
		m.quitting = true
		return m, tea.Quit
	}

	if m.page == pageDetail {
		return m.handleDetailKey(msg)
	}
	if m.page == pageHeatmap {
		return m.handleHeatmapKey(msg)
	}
	return m.handleListKey(msg)
}

func (m model) View() string {
	if m.quitting {
		return ""
	}
	if m.page == pageDetail {
		return m.detailView()
	}
	if m.page == pageHeatmap {
		return m.heatmapView()
	}
	return m.listView()
}

// ── helpers ─────────────────────────────────────────────────────

func clamp(v, lo, hi int) int {
	if hi < lo {
		return lo
	}
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func (m model) visualRange() (int, int) {
	if m.cursor <= m.visualStart {
		return m.cursor, m.visualStart
	}
	return m.visualStart, m.cursor
}

func loadConfig() *config.Config {
	cfg := config.Default()
	cfgPath := os.ExpandEnv("$HOME/.config/gitmap/config.yaml")
	if _, err := os.Stat(cfgPath); err == nil {
		if c, err := config.Load(cfgPath); err == nil {
			return c
		}
	}
	return cfg
}

func openInTerm(path, cmd string) error {
	escaped := strings.ReplaceAll(path, "'", "'\\''")
	script := fmt.Sprintf(
		`tell application "iTerm2"
			activate
			try
				tell current window
					create tab with default profile
				end tell
			on error
				create window with default profile
			end try
			tell current session of current tab of current window
				write text "cd '%s' && %s"
			end tell
		end tell`, escaped, cmd)
	return exec.Command("osascript", "-e", script).Run()
}

func openInITerm(path string) {
	if err := openInTerm(path, "clear"); err != nil {
		fmt.Fprintf(os.Stderr, "gitmap: failed to open iTerm: %v\n", err)
	}
}

func openClaude(path string) {
	if err := openInTerm(path, "claude --dangerously-skip-permissions"); err != nil {
		fmt.Fprintf(os.Stderr, "gitmap: failed to open Claude: %v\n", err)
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--shell-init" {
		printShellInit()
		return
	}

	cfg := loadConfig()

	p := tea.NewProgram(
		model{loading: true, scanPaths: cfg.ScanPaths, excludeRepos: cfg.ExcludeRepos, author: cfg.Author, autoFetch: cfg.AutoFetch, lastActivity: time.Now()},
		tea.WithAltScreen(),
	)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func printShellInit() {
	fmt.Print(`# gitmap shell integration
gmap() {
    local target
    target=$(gitmap 2>/dev/null)
    if [ -n "$target" ] && [ -d "$target" ]; then
        cd "$target" || return
        echo "→ $target"
    fi
}
`)
}
