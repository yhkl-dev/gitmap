package main

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	gitpkg "github.com/yhkl-dev/gitmap/internal/git"
)

func (m model) handleDetailKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	ks := msg.String()

	// branch select sub-mode
	if m.branchSelect {
		switch ks {
		case "esc":
			m.branchSelect = false
			m.branchFilter = ""
			return m, nil
		case "enter":
			if m.detailRepo != nil && m.branchFilter != "" {
				branches := strings.Split(m.detailBranches, "\n")
				for _, b := range branches {
					name := strings.TrimPrefix(strings.TrimSpace(b), "* ")
					if fuzzyMatch(m.branchFilter, name) {
						result := gitpkg.Checkout(m.detailRepo.Path, name)
						gitpkg.InvalidateCache(m.detailRepo.Path)
						m.branchSelect = false
						m.branchFilter = ""
						m.detailStashResult = "checkout " + name + ": " + result
						m.detailDiff = gitpkg.DetailedDiff(m.detailRepo.Path)
						m.detailLog = gitpkg.RecentCommits(m.detailRepo.Path, 10)
						m.detailBranches = gitpkg.BranchList(m.detailRepo.Path)
						m.detailStashes = gitpkg.StashList(m.detailRepo.Path)
						m.detailScroll = 0
						return m, nil
					}
				}
			}
			m.branchSelect = false
			m.branchFilter = ""
			return m, nil
		case "backspace":
			if len(m.branchFilter) > 0 {
				m.branchFilter = m.branchFilter[:len(m.branchFilter)-1]
			} else {
				m.branchSelect = false
			}
			return m, nil
		default:
			if len(msg.Runes) == 1 {
				m.branchFilter += string(msg.Runes[0])
			}
			return m, nil
		}
	}

	switch ks {
	case "esc", "q":
		m.page = pageList
		m.detailRepo = nil
		m.prsOutput = ""
		m.detailLog = ""
		m.detailBranches = ""
		m.detailStashes = ""
		m.prsLoading = false
		m.showDiff = false
		m.detailDiff = ""
		m.detailStashResult = ""
		m.detailScroll = 0
		m.lastKey = ""
		m.prefixCount = 0
		m.branchSelect = false
		m.branchFilter = ""

	// ── scrolling ─────────────────────────────────────────────

	case "j", "down":
		m.detailScroll++

	case "k", "up":
		if m.detailScroll > 0 {
			m.detailScroll--
		}

	case "ctrl+u":
		half := m.visibleDetailRows() / 2
		if half < 1 {
			half = 1
		}
		m.detailScroll -= half
		if m.detailScroll < 0 {
			m.detailScroll = 0
		}

	case "ctrl+d":
		half := m.visibleDetailRows() / 2
		if half < 1 {
			half = 1
		}
		m.detailScroll += half

	case "ctrl+b":
		m.detailScroll -= m.visibleDetailRows()
		if m.detailScroll < 0 {
			m.detailScroll = 0
		}

	case "ctrl+f":
		m.detailScroll += m.visibleDetailRows()

	case "g":
		if m.lastKey == "g" {
			m.detailScroll = 0
			m.lastKey = ""
		} else {
			m.lastKey = "g"
		}
		m.prefixCount = 0
		return m, nil

	case "G":
		m.detailScroll = 1 << 30
		m.lastKey = ""
		m.prefixCount = 0

	// ── stash ────────────────────────────────────────────────

	case "a":
		if m.detailRepo != nil {
			m.detailStashResult = gitpkg.StashApply(m.detailRepo.Path)
			gitpkg.InvalidateCache(m.detailRepo.Path)
			m.detailDiff = gitpkg.DetailedDiff(m.detailRepo.Path)
			m.detailLog = gitpkg.RecentCommits(m.detailRepo.Path, 10)
			m.detailBranches = gitpkg.BranchList(m.detailRepo.Path)
			m.detailStashes = gitpkg.StashList(m.detailRepo.Path)
			m.detailScroll = 0
		}

	case "A":
		if m.detailRepo != nil {
			m.detailStashResult = gitpkg.StashPop(m.detailRepo.Path)
			gitpkg.InvalidateCache(m.detailRepo.Path)
			m.detailDiff = gitpkg.DetailedDiff(m.detailRepo.Path)
			m.detailLog = gitpkg.RecentCommits(m.detailRepo.Path, 10)
			m.detailBranches = gitpkg.BranchList(m.detailRepo.Path)
			m.detailStashes = gitpkg.StashList(m.detailRepo.Path)
			m.detailScroll = 0
		}

	case "D":
		if m.detailRepo != nil {
			m.detailStashResult = gitpkg.StashDrop(m.detailRepo.Path)
			gitpkg.InvalidateCache(m.detailRepo.Path)
			m.detailDiff = gitpkg.DetailedDiff(m.detailRepo.Path)
			m.detailLog = gitpkg.RecentCommits(m.detailRepo.Path, 10)
			m.detailBranches = gitpkg.BranchList(m.detailRepo.Path)
			m.detailStashes = gitpkg.StashList(m.detailRepo.Path)
			m.detailScroll = 0
		}

	// ── actions ───────────────────────────────────────────────

	case "o":
		if m.detailRepo != nil {
			openInITerm(m.detailRepo.Path)
		}

	case "O":
		if m.detailRepo != nil {
			openInITerm(m.detailRepo.Path)
		}

	case "d":
		m.showDiff = !m.showDiff
		m.detailScroll = 0

	case "b":
		if m.detailRepo != nil {
			var msg string
			if ghPath, err := exec.LookPath("gh"); err == nil {
				cmd := exec.Command(ghPath, "pr", "view", "--web")
				cmd.Dir = m.detailRepo.Path
				if err := cmd.Start(); err != nil {
					msg = "gh: " + err.Error()
				}
			} else if m.detailRepo.RemoteURL != "" {
				httpsURL := gitpkg.HttpsRemote(m.detailRepo.RemoteURL)
				if openPath, err := exec.LookPath("open"); err == nil {
					if err := exec.Command(openPath, httpsURL).Start(); err != nil {
						msg = "open: " + err.Error()
					}
				} else {
					msg = "neither gh nor open found"
				}
			} else {
				msg = "no remote URL configured"
			}
			if msg != "" {
				m.detailStashResult = msg
			}
		}

	case "B":
		m.branchSelect = true
		m.branchFilter = ""
		return m, nil

	case "h":
		m.page = pageHeatmap
		if !m.heatmapFresh() {
			m.heatmapLoading = true
			m.heatmapCommits = nil
			m.heatmapLines = nil
			return m, loadHeatmapCmd(m.allRepos, m.authors)
		}
		return m, nil

	case "r":
		if m.detailRepo != nil {
			m.detailDiff = gitpkg.DetailedDiff(m.detailRepo.Path)
			m.detailLog = gitpkg.RecentCommits(m.detailRepo.Path, 10)
			m.detailBranches = gitpkg.BranchList(m.detailRepo.Path)
			m.detailStashes = gitpkg.StashList(m.detailRepo.Path)
			m.prsOutput = ""
			m.prsLoading = true
			m.detailScroll = 0
			return m, loadPRsCmd(m.detailRepo.Path)
		}

	default:
		m.lastKey = ""
		m.prefixCount = 0
		return m, nil
	}
	return m, nil
}

func (m model) visibleDetailRows() int {
	headerLines := 4
	footerLines := 1
	v := m.height - headerLines - footerLines
	if v < 1 {
		v = 1
	}
	return v
}

func (m model) detailView() string {
	r := m.detailRepo
	if r == nil {
		return ""
	}

	dot := cleanDot
	if r.Dirty {
		dot = dirtyDot
	}
	if r.Conflict {
		dot = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Render("⚡")
	}
	branch := r.Branch
	if branch == "" {
		branch = "(no commits)"
	}

	var tags []string
	if r.Ahead > 0 {
		tags = append(tags, aheadStyle(fmt.Sprintf("↑%d", r.Ahead)))
	}
	if r.Behind > 0 {
		tags = append(tags, behindStyle(fmt.Sprintf("↓%d", r.Behind)))
	}
	if r.StashCount > 0 {
		tags = append(tags, stashStyle(fmt.Sprintf("≡%d", r.StashCount)))
	}
	if r.Untracked > 0 {
		tags = append(tags, untrackStyle(fmt.Sprintf("?%d untracked", r.Untracked)))
	}
	if r.Conflict {
		tags = append(tags, untrackStyle("[conflict]"))
	}
	tagStr := ""
	if len(tags) > 0 {
		tagStr = "  " + strings.Join(tags, " ")
	}

	header := bold.Render("gitmap › "+r.Name) +
		muted.Render("  "+branch+tagStr+"  ") + dot
	if m.branchSelect {
		branches := strings.Split(m.detailBranches, "\n")
		matches := 0
		for _, b := range branches {
			name := strings.TrimPrefix(strings.TrimSpace(b), "* ")
			if fuzzyMatch(m.branchFilter, name) {
				matches++
			}
		}
		header += "\n" + muted.Render(fmt.Sprintf("  matching branches: %d", matches))
	}
	if r.RemoteURL != "" {
		header += "\n" + muted.Render("  " + r.RemoteURL)
	}

	sep := muted.Render(strings.Repeat("─", max(40, m.width-2)))

	var body strings.Builder

	// PRs
	body.WriteString(bold.Render("── PRs"))
	if m.prsLoading {
		body.WriteString(muted.Render("  loading..."))
	} else {
		body.WriteString("\n")
		if m.prsOutput != "" {
			body.WriteString("  ")
			body.WriteString(strings.ReplaceAll(m.prsOutput, "\n", "\n  "))
		} else {
			body.WriteString(muted.Render("  (gh CLI not found)"))
		}
	}
	body.WriteString("\n\n")

	// branches
	body.WriteString(bold.Render("── branches") + "\n")
	if m.detailBranches != "" {
		for _, line := range strings.Split(m.detailBranches, "\n") {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "*") {
				body.WriteString("  " + lipgloss.NewStyle().Foreground(green).Render(trimmed) + "\n")
			} else {
				body.WriteString("  " + muted.Render(trimmed) + "\n")
			}
		}
	} else {
		body.WriteString(muted.Render("  (no branches)") + "\n")
	}
	body.WriteString("\n")

	// commits
	body.WriteString(bold.Render("── commits"))
	if r.LastCommit != "" {
		body.WriteString(muted.Render("  " + r.LastCommit))
	}
	body.WriteString("\n")
	if m.detailLog != "" {
		for _, line := range strings.Split(m.detailLog, "\n") {
			body.WriteString("  " + muted.Render(line) + "\n")
		}
	} else {
		body.WriteString(muted.Render("  (no commits)") + "\n")
	}
	body.WriteString("\n")

	// stashes
	if m.detailStashes != "" {
		body.WriteString(bold.Render("── stashes") + "\n")
		for _, line := range strings.Split(m.detailStashes, "\n") {
			body.WriteString("  " + muted.Render(strings.TrimSpace(line)) + "\n")
		}
		body.WriteString("\n")
	}

	// stash result
	if m.detailStashResult != "" {
		body.WriteString(bold.Render("── stash result") + "\n")
		for _, line := range strings.Split(m.detailStashResult, "\n") {
			body.WriteString("  " + muted.Render(line) + "\n")
		}
		body.WriteString("\n")
	}

	// diff
	if m.showDiff {
		body.WriteString(bold.Render("── changes") + "\n")
		if m.detailDiff != "" {
			body.WriteString(m.detailDiff)
		} else {
			body.WriteString(muted.Render("(no changes)"))
		}
		body.WriteString("\n")
	}

	bodyLines := strings.Split(strings.TrimSuffix(body.String(), "\n"), "\n")
	visible := m.visibleDetailRows()

	m.detailScroll = clamp(m.detailScroll, 0, max(0, len(bodyLines)-visible))

	end := m.detailScroll + visible
	if end > len(bodyLines) {
		end = len(bodyLines)
	}
	scrolledBody := strings.Join(bodyLines[m.detailScroll:end], "\n")

	if m.detailScroll > 0 {
		scrolledBody = muted.Render(fmt.Sprintf("  ↑ %d more above", m.detailScroll)) + "\n" + scrolledBody
	}
	if end < len(bodyLines) {
		scrolledBody += "\n" + muted.Render(fmt.Sprintf("  ↓ %d more below", len(bodyLines)-end))
	}

	var footer string
	if m.branchSelect {
		footer = lipgloss.NewStyle().Foreground(yellow).Render("checkout: " + m.branchFilter + "_") +
			muted.Render("  (esc cancel  enter confirm)")
	} else {
		footer = muted.Render("esc/q back  j/k scroll  ctrl+d/u page  gg/G top/bot  a apply  A pop  D drop  o cd  d diff  h heatmap  r refresh  b pr-browse  B checkout")
	}

	return header + "\n" + sep + "\n\n" + scrolledBody + "\n" + footer + "\n"
}
