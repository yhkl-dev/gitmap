package main

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	gitpkg "github.com/yhkl-dev/gitmap/internal/git"
)

const heatmapDays = 365

// ── palette ──────────────────────────────────────────────────────

var heatEmpty = lipgloss.Color("#1e1e20")
var heatCommitsColors = []lipgloss.Color{
	lipgloss.Color("#1a4d2e"),
	lipgloss.Color("#2d8a4e"),
	lipgloss.Color("#4ade80"),
	lipgloss.Color("#86efac"),
}
var heatLinesColors = []lipgloss.Color{
	lipgloss.Color("#1a3a5c"),
	lipgloss.Color("#2d6da4"),
	lipgloss.Color("#4a9eff"),
	lipgloss.Color("#93c5fd"),
}

var dayNames = [7]string{"Su", "Mo", "Tu", "We", "Th", "Fr", "Sa"}

// ── messages ─────────────────────────────────────────────────────

type heatmapLoadedMsg struct {
	commits map[string]int
	lines   map[string]int
	total   int
	linesTotal int
	repos   int
}

// ── command ──────────────────────────────────────────────────────

func loadHeatmapCmd(repos []gitpkg.RepoStatus, authors []string) tea.Cmd {
	return func() tea.Msg {
		const maxConcurrent = 8
		sem := make(chan struct{}, maxConcurrent)
		var mu sync.Mutex
		aggCommits := make(map[string]int)
		aggLines := make(map[string]int)
		var wg sync.WaitGroup
		repoCount := 0

		for _, r := range repos {
			wg.Add(1)
			go func(path string) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				stats, err := gitpkg.CommitsStats(path, heatmapDays, authors)
				if err != nil || len(stats) == 0 {
					return
				}

				mu.Lock()
				repoCount++
				for d, s := range stats {
					aggCommits[d] += s.Commits
					aggLines[d] += s.LinesChanged
				}
				mu.Unlock()
			}(r.Path)
		}
		wg.Wait()

		total := 0
		for _, c := range aggCommits {
			total += c
		}
		linesTotal := 0
		for _, l := range aggLines {
			linesTotal += l
		}
		return heatmapLoadedMsg{
			commits:    aggCommits,
			lines:      aggLines,
			total:      total,
			linesTotal: linesTotal,
			repos:      repoCount,
		}
	}
}

// ── key handling ─────────────────────────────────────────────────

func (m model) heatmapFresh() bool {
	if len(m.heatmapCommits) == 0 && len(m.heatmapLines) == 0 {
		return false
	}
	now := time.Now()
	return now.Year() == m.heatmapLoadedAt.Year() && now.YearDay() == m.heatmapLoadedAt.YearDay()
}

func (m model) handleHeatmapKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q", "h":
		m.page = pageList
	case "r":
		if !m.heatmapLoading {
			m.heatmapLoading = true
			return m, loadHeatmapCmd(m.allRepos, m.authors)
		}
	}
	return m, nil
}

// ── grid ─────────────────────────────────────────────────────────

type monthMark struct {
	col  int
	name string
}

func buildGrid(data map[string]int) (grid [][]int, weeks int, months []monthMark, maxVal int) {
	now := time.Now()
	start := now.AddDate(0, 0, -heatmapDays)
	for start.Weekday() != time.Sunday {
		start = start.AddDate(0, 0, -1)
	}

	for _, c := range data {
		if c > maxVal {
			maxVal = c
		}
	}
	if maxVal < 1 {
		maxVal = 1
	}

	days := int(now.Sub(start).Hours()/24) + 1
	weeks = (days + 6) / 7

	grid = make([][]int, 7)
	for i := range grid {
		grid[i] = make([]int, weeks)
	}

	prevMonth := time.Month(-1)
	for col := 0; col < weeks; col++ {
		weekDate := start.AddDate(0, 0, col*7+3)
		if m := weekDate.Month(); m != prevMonth {
			months = append(months, monthMark{col, weekDate.Format("Jan")})
			prevMonth = m
		}
		for row := 0; row < 7; row++ {
			d := start.AddDate(0, 0, col*7+row)
			if d.After(now) {
				grid[row][col] = -1
				continue
			}
			grid[row][col] = data[d.Format("2006-01-02")]
		}
	}
	return
}

func cellColor(count int, maxVal int, palette []lipgloss.Color) lipgloss.Color {
	if count <= 0 {
		return heatEmpty
	}
	ratio := float64(count) / float64(maxVal)
	switch {
	case ratio <= 0.25:
		return palette[0]
	case ratio <= 0.5:
		return palette[1]
	case ratio <= 0.75:
		return palette[2]
	default:
		return palette[3]
	}
}

func heatCell(count int, maxVal int, cellChar string, palette []lipgloss.Color) string {
	if count < 0 {
		return strings.Repeat(" ", len(cellChar))
	}
	return lipgloss.NewStyle().Foreground(cellColor(count, maxVal, palette)).Render(cellChar)
}

// ── stat helpers ─────────────────────────────────────────────────

func fmtNum(n int) string {
	if n >= 1000 {
		return fmt.Sprintf("%d,%03d", n/1000, n%1000)
	}
	return fmt.Sprintf("%d", n)
}

// ── view ─────────────────────────────────────────────────────────

var (
	hmLabel  = lipgloss.NewStyle().Foreground(lipgloss.Color("#6b7280"))
	hmAccent = lipgloss.NewStyle().Foreground(lipgloss.Color("#4ade80"))
	hmBlue   = lipgloss.NewStyle().Foreground(lipgloss.Color("#4a9eff"))
	hmDim    = lipgloss.NewStyle().Foreground(lipgloss.Color("#4a4a4d"))
)

func (m model) heatmapView() string {
	header := bold.Render("gitmap › heatmap")

	if m.heatmapLoading {
		return centerIn(m.height, header+"\n\n"+muted.Render("  scanning commits across repos..."))
	}

	if len(m.heatmapCommits) == 0 && len(m.heatmapLines) == 0 {
		return centerIn(m.height, header+"\n\n"+muted.Render("  no commits found in the past year"))
	}

	// responsive cell sizing
	labelW := 4
	weeks := 53 // approximate, refined below
	cellChar := "■"
	cellW := 1
	if m.width >= 100 {
		cellChar = "■ "
		cellW = 2
	}

	// build both grids
	commitGrid, commitWeeks, months, commitMax := buildGrid(m.heatmapCommits)
	linesGrid, linesWeeks, _, linesMax := buildGrid(m.heatmapLines)
	weeks = commitWeeks
	if linesWeeks > weeks {
		weeks = linesWeeks
	}

	// ── stat bar ────────────────────────────────────────────────
	statBar := fmt.Sprintf("%s commits  %s  %s lines  %s  %s repos  %s  past year",
		hmAccent.Render(fmtNum(m.heatmapTotal)),
		hmDim.Render("·"),
		hmBlue.Render(fmtNum(m.heatmapLinesTotal)),
		hmDim.Render("·"),
		hmAccent.Render(fmtNum(m.heatmapRepos)),
		hmDim.Render("·"),
	)

	// ── month labels ───────────────────────────────────────────
	monthLine := buildMonthLine(months, cellW)

	// ── section: commits ───────────────────────────────────────
	commitsSection := hmAccent.Render("Commits") + "\n" +
		strings.Repeat(" ", labelW) + monthLine + "\n" +
		renderGrid(commitGrid, weeks, cellChar, cellW, commitMax, heatCommitsColors)

	// ── section: lines changed ─────────────────────────────────
	linesSection := hmBlue.Render("Lines Changed") + "\n" +
		strings.Repeat(" ", labelW) + monthLine + "\n" +
		renderGrid(linesGrid, weeks, cellChar, cellW, linesMax, heatLinesColors)

	// ── legend ─────────────────────────────────────────────────
	leg := renderLegend(cellChar, commitMax, heatCommitsColors)

	footer := hmLabel.Render("h/esc/q back  r refresh")

	content := header + "\n" + statBar + "\n\n" +
		commitsSection + "\n" +
		linesSection + "\n" +
		leg + "\n\n" +
		footer

	return centerIn(m.height, content)
}

func buildMonthLine(months []monthMark, cellW int) string {
	cursor := 0
	var b strings.Builder
	for _, mm := range months {
		pos := mm.col * cellW
		pad := pos - cursor
		if pad < 0 {
			pad = 0
		}
		b.WriteString(strings.Repeat(" ", pad))
		b.WriteString(hmLabel.Render(mm.name))
		cursor = pos + len(mm.name)
	}
	return b.String()
}

func renderGrid(grid [][]int, weeks int, cellChar string, cellW int, maxVal int, palette []lipgloss.Color) string {
	labelW := 4
	labelRows := map[int]bool{1: true, 3: true, 5: true}
	labelPad := strings.Repeat(" ", labelW)

	var b strings.Builder
	for row := 0; row < 7; row++ {
		if labelRows[row] {
			b.WriteString(" " + hmLabel.Render(dayNames[row]) + " ")
		} else {
			b.WriteString(labelPad)
		}
		for col := 0; col < weeks; col++ {
			b.WriteString(heatCell(grid[row][col], maxVal, cellChar, palette))
		}
		b.WriteString("\n")
	}
	return b.String()
}

func renderLegend(cellChar string, maxVal int, palette []lipgloss.Color) string {
	var b strings.Builder
	b.WriteString(hmLabel.Render("Less "))
	for _, n := range []int{0, 1, 3, 6, 10} {
		b.WriteString(heatCell(n, maxVal, cellChar, palette))
	}
	b.WriteString(hmLabel.Render(" More"))
	return b.String()
}

// centerIn pads content vertically to fill height.
func centerIn(height int, content string) string {
	lines := strings.Split(content, "\n")
	topPad := (height - len(lines)) / 2
	if topPad < 0 {
		topPad = 0
	}
	var out strings.Builder
	for i := 0; i < topPad; i++ {
		out.WriteString("\n")
	}
	out.WriteString(content)
	return out.String()
}
