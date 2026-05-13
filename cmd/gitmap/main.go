package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/yhkl-dev/gitmap/internal/config"
	gitpkg "github.com/yhkl-dev/gitmap/internal/git"
	"github.com/yhkl-dev/gitmap/internal/scanner"
)

var (
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Bold(true)
	dirtyMark     = lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Render("●")
	cleanMark     = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render("○")
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("5"))
	helpStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)

type model struct {
	repos    []gitpkg.RepoStatus
	cursor   int
	diff     string
	showDiff bool
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
		case "d":
			m.showDiff = !m.showDiff
			m.diff = ""
			if m.showDiff && len(m.repos) > 0 {
				m.diff = gitpkg.Diff(m.repos[m.cursor].Path)
			}
		case "r":
			// Refresh repo list
			m.repos = loadRepos()
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.quitting {
		return ""
	}

	s := titleStyle.Render("gitmap") + helpStyle.Render(fmt.Sprintf("  %d repos", len(m.repos))) + "\n\n"

	for i, r := range m.repos {
		mark := cleanMark
		if r.Dirty {
			mark = dirtyMark
		}

		branch := ""
		if r.Branch != "" {
			branch = "  " + helpStyle.Render(r.Branch)
		}

		line := fmt.Sprintf("  %s  %s%s", mark, r.Name, branch)
		if i == m.cursor {
			line = selectedStyle.Render(line)
		}
		s += line + "\n"
	}

	s += "\n"
	if m.showDiff && m.diff != "" {
		s += helpStyle.Render("── diff ──────────────────────") + "\n"
		s += m.diff + "\n\n"
	}
	s += helpStyle.Render("↑/↓ navigate  ⏎ cd  d diff  r refresh  q quit") + "\n"
	return s
}

func loadRepos() []gitpkg.RepoStatus {
	cfg := config.Default()
	cfgPath := os.ExpandEnv("$HOME/.config/gitmap/config.yaml")
	if _, err := os.Stat(cfgPath); err == nil {
		c, cerr := config.Load(cfgPath)
		if cerr == nil {
			cfg = c
		}
	}

	repos, err := scanner.Scan(cfg.ScanPaths)
	if err != nil {
		return nil
	}

	var statuses []gitpkg.RepoStatus
	for _, r := range repos {
		s, err := gitpkg.Status(r.Path)
		if err != nil {
			s = &gitpkg.RepoStatus{Path: r.Path, Name: r.Name}
		}
		statuses = append(statuses, *s)
	}
	return statuses
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--shell-init" {
		printShellInit()
		return
	}

	repos := loadRepos()

	p := tea.NewProgram(model{repos: repos})
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
