package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	gitpkg "github.com/yhkl-dev/gitmap/internal/git"
)

func fuzzyMatch(query, target string) bool {
	if query == "" {
		return true
	}
	qr := []rune(strings.ToLower(query))
	tr := []rune(strings.ToLower(target))
	j := 0
	for _, r := range tr {
		if j < len(qr) && r == qr[j] {
			j++
		}
	}
	return j == len(qr)
}

func (m model) handleListKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	ks := msg.String()

	// ── filter (search) mode ───────────────────────────────────
	if m.filtering {
		switch ks {
		case "esc":
			m.filter = ""
			m.filtering = false
			m.cursor = 0
			m.scrollOffset = 0
			m.prefixCount = 0
			m.lastKey = ""

		case "enter":
			m.filtering = false
			m.prefixCount = 0
			m.lastKey = ""

		case "backspace":
			if len(m.filter) > 0 {
				m.filter = m.filter[:len(m.filter)-1]
			} else {
				m.filtering = false
			}
			m.cursor = 0
			m.scrollOffset = 0

		default:
			if len(msg.Runes) == 1 {
				m.filter += string(msg.Runes[0])
				m.cursor = 0
				m.scrollOffset = 0
			}
		}
		return m, nil
	}

	// ── normal mode ────────────────────────────────────────────

	// number prefix accumulation (0–9)
	if len(msg.Runes) == 1 && msg.Runes[0] >= '0' && msg.Runes[0] <= '9' {
		m.prefixCount = m.prefixCount*10 + int(msg.Runes[0]-'0')
		m.lastKey = ""
		return m, nil
	}

	// gg — double-tap g to go to top
	if ks == "g" {
		if m.lastKey == "g" {
			repos := m.filteredRepos()
			target := m.prefixCount
			m.prefixCount = 0
			m.lastKey = ""
			if target > 0 && target <= len(repos) {
				m.cursor = target - 1
			} else {
				m.cursor = 0
			}
			m.scrollOffset = 0
		} else {
			m.lastKey = "g"
			m.prefixCount = 0
		}
		return m, nil
	}

	repos := m.filteredRepos()
	m.cursor = clamp(m.cursor, 0, len(repos)-1)
	prefix := m.prefixCount
	m.prefixCount = 0
	m.lastKey = ""

	switch ks {

	// ── movement ──────────────────────────────────────────────

	case "j", "down":
		n := prefix
		if n == 0 {
			n = 1
		}
		m.cursor = clamp(m.cursor+n, 0, len(repos)-1)

	case "k", "up":
		n := prefix
		if n == 0 {
			n = 1
		}
		m.cursor = clamp(m.cursor-n, 0, len(repos)-1)

	case "ctrl+u":
		half := m.visibleRows() / 2
		if half < 1 {
			half = 1
		}
		m.cursor = clamp(m.cursor-half, 0, len(repos)-1)

	case "ctrl+d":
		half := m.visibleRows() / 2
		if half < 1 {
			half = 1
		}
		m.cursor = clamp(m.cursor+half, 0, len(repos)-1)

	case "ctrl+b":
		page := m.visibleRows()
		m.cursor = clamp(m.cursor-page, 0, len(repos)-1)

	case "ctrl+f":
		page := m.visibleRows()
		m.cursor = clamp(m.cursor+page, 0, len(repos)-1)

	case "G":
		if prefix > 0 && prefix <= len(repos) {
			m.cursor = prefix - 1
		} else {
			m.cursor = clamp(len(repos)-1, 0, len(repos)-1)
		}

	case "n":
		if m.filter == "" {
			break // no-op without active search
		}
		n := prefix
		if n == 0 {
			n = 1
		}
		m.cursor = clamp(m.cursor+n, 0, len(repos)-1)

	case "N":
		if m.filter == "" {
			break // no-op without active search
		}
		n := prefix
		if n == 0 {
			n = 1
		}
		m.cursor = clamp(m.cursor-n, 0, len(repos)-1)

	// ── actions ───────────────────────────────────────────────

	case "v":
		if m.visualMode {
			m.visualMode = false
		} else {
			m.visualMode = true
			m.visualStart = m.cursor
		}

	case "esc":
		if m.visualMode {
			m.visualMode = false
		}

	case "q":
		m.visualMode = false
		m.quitting = true
		return m, tea.Quit

	case "s":
		m.visualMode = false
		m.sortMode = (m.sortMode + 1) % 3
		m.cursor = 0
		m.scrollOffset = 0

	case "enter":
		m.visualMode = false
		if len(repos) > 0 {
			r := repos[m.cursor]
			m.page = pageDetail
			m.detailRepo = &r
			m.detailScroll = 0
			m.detailDiff = gitpkg.Diff(r.Path)
			m.detailLog = gitpkg.RecentCommits(r.Path, 10)
			m.detailBranches = gitpkg.BranchList(r.Path)
			m.detailStashes = gitpkg.StashList(r.Path)
			m.prsOutput = ""
			m.prsLoading = true
			m.showDiff = true
			return m, loadPRsCmd(r.Path)
		}

	case "o":
		m.visualMode = false
		if len(repos) > 0 {
			openInITerm(repos[m.cursor].Path)
		}

	case "O":
		m.visualMode = false
		if len(repos) > 0 {
			openInITerm(repos[m.cursor].Path)
		}

	case "c":
		m.visualMode = false
		if len(repos) > 0 {
			openClaude(repos[m.cursor].Path)
		}

	case "f":
		if m.visualMode {
			start, end := m.visualRange()
			selected := repos[start : end+1]
			if !m.fetching && len(selected) > 0 {
				m.visualMode = false
				m.fetching = true
				m.fetchProgress = fmt.Sprintf("fetching %d repos...", len(selected))
				return m, batchFetchCmd(selected)
			}
		} else if !m.fetching && len(repos) > 0 {
			m.fetching = true
			m.fetchProgress = "fetching..."
			return m, singleFetchCmd(repos[m.cursor].Path)
		}

	case "F":
		if !m.fetching {
			m.fetching = true
			m.fetchProgress = "fetching all..."
			return m, batchFetchCmd(m.allRepos)
		}

	case "p":
		if m.visualMode {
			start, end := m.visualRange()
			selected := repos[start : end+1]
			if !m.pulling && len(selected) > 0 {
				m.visualMode = false
				m.pulling = true
				m.pullProgress = fmt.Sprintf("pulling %d repos...", len(selected))
				return m, batchPullCmd(selected)
			}
		} else if !m.pulling && len(repos) > 0 {
			m.pulling = true
			m.pullProgress = "pulling..."
			return m, singlePullCmd(repos[m.cursor].Path)
		}

	case "P":
		if !m.pulling {
			m.pulling = true
			m.pullProgress = "pulling all..."
			return m, batchPullCmd(m.allRepos)
		}

	case "d":
		m.visualMode = false
		m.showDiff = !m.showDiff
		if m.showDiff && len(repos) > 0 {
			m.detailDiff = gitpkg.Diff(repos[m.cursor].Path)
		} else {
			m.detailDiff = ""
		}

	case "r":
		m.visualMode = false
		if !m.loading {
			m.loading = true
			m.cursor = 0
			m.scrollOffset = 0
			m.detailDiff = ""
			return m, loadReposCmd(m.scanPaths, m.excludeRepos)
		}

	case "h":
		m.visualMode = false
		m.page = pageHeatmap
		if !m.heatmapFresh() {
			m.heatmapLoading = true
			m.heatmapCommits = nil
			m.heatmapLines = nil
			return m, loadHeatmapCmd(m.allRepos, m.author)
		}
		return m, nil

	case "/":
		m.visualMode = false
		m.filtering = true
		m.filter = ""
		m.cursor = 0
		m.scrollOffset = 0
	}
	return m, nil
}

func (m model) filteredRepos() []gitpkg.RepoStatus {
	var repos []gitpkg.RepoStatus
	if m.filter == "" {
		repos = make([]gitpkg.RepoStatus, len(m.allRepos))
		copy(repos, m.allRepos)
	} else {
		for _, r := range m.allRepos {
			if fuzzyMatch(m.filter, r.Name) || fuzzyMatch(m.filter, r.Branch) {
				repos = append(repos, r)
			}
		}
	}

	switch m.sortMode {
	case sortByName:
		sort.Slice(repos, func(i, j int) bool {
			return strings.ToLower(repos[i].Name) < strings.ToLower(repos[j].Name)
		})
	case sortByDirty:
		sort.Slice(repos, func(i, j int) bool {
			if repos[i].Dirty != repos[j].Dirty {
				return repos[i].Dirty
			}
			return strings.ToLower(repos[i].Name) < strings.ToLower(repos[j].Name)
		})
	}
	return repos
}

func (m model) visibleRows() int {
	overhead := 5 // header + slogan + sep + blank before list + blank before footer
	if m.filtering {
		overhead++ // search prompt footer
	} else {
		overhead += 2 // footer + actionBar
	}
	if m.showDiff && m.detailDiff != "" {
		// diffBlock = "\n" + sep + "\n" + "── diff ──" + "\n" + detailDiff + "\n"
		overhead += strings.Count(m.detailDiff, "\n") + 4
	}
	v := m.height - overhead
	if v < 1 {
		v = 1
	}
	return v
}

func (m model) listView() string {
	if m.loading {
		header := bold.Render("gitmap") + muted.Render("  loading...")
		return header + "\n\n" + muted.Render(`"Un jour je serai de retour près de toi"`) + "\n\n" +
			muted.Render("scanning repositories in parallel...") + "\n"
	}

	repos := m.filteredRepos()

	// clamp cursor for view — mutations are discarded (value receiver)
	m.cursor = clamp(m.cursor, 0, len(repos)-1)

	dirtyCount := 0
	stashTotal := 0
	untrackedTotal := 0
	for _, r := range repos {
		if r.Dirty {
			dirtyCount++
		}
		stashTotal += r.StashCount
		untrackedTotal += r.Untracked
	}
	header := bold.Render("gitmap") +
		muted.Render(fmt.Sprintf("  %d repos", len(repos)))
	if dirtyCount > 0 {
		header += muted.Render(fmt.Sprintf("  %d dirty", dirtyCount))
	}
	if stashTotal > 0 {
		header += muted.Render(fmt.Sprintf("  %d stashes", stashTotal))
	}
	if untrackedTotal > 0 {
		header += untrackStyle(fmt.Sprintf("  %d untracked", untrackedTotal))
	}
	if m.errorCount < 0 {
		header += untrackStyle("  scan failed")
	} else if m.errorCount > 0 {
		header += untrackStyle(fmt.Sprintf("  %d errors", m.errorCount))
	}
	header += muted.Render(fmt.Sprintf("  sort: %s", sortNames[m.sortMode]))
	if m.fetching {
		header += muted.Render("  " + m.fetchProgress)
	} else if m.fetchProgress != "" {
		header += muted.Render("  " + m.fetchProgress)
	}
	if m.pulling {
		header += muted.Render("  " + m.pullProgress)
	} else if m.pullProgress != "" {
		header += muted.Render("  " + m.pullProgress)
	}

	slogan := muted.Render(`"Un jour je serai de retour près de toi"`)
	sep := muted.Render(strings.Repeat("─", max(40, m.width-2)))

	var diffBlock string
	if m.showDiff && m.detailDiff != "" {
		diffBlock = "\n" + sep + "\n" + muted.Render("── diff ──") + "\n" + m.detailDiff + "\n"
	}

	footerSep := ""
	footer := muted.Render("j/k ↓↑  ctrl+d/u ½pg  ctrl+f/b pg  gg/G ↥↧  n/N next  5j prefix  h heatmap")
	actionBar := muted.Render("/ search  ⏎ detail  s sort  h heatmap  o cd  c claude  f fetch  p pull  d diff  r refresh  v select  q quit")

	if m.filtering {
		searchPrompt := lipgloss.NewStyle().Foreground(yellow).Render("/" + m.filter + "_")
		footer = searchPrompt + muted.Render("  (esc cancel  enter confirm)")
		actionBar = ""
	} else if m.filter != "" {
		footerSep = muted.Render("  filter: " + m.filter)
	}

	visible := m.visibleRows()

	m.scrollOffset = clamp(m.scrollOffset, 0, max(0, len(repos)-visible))
	if m.cursor < m.scrollOffset {
		m.scrollOffset = m.cursor
	}
	if m.cursor >= m.scrollOffset+visible {
		m.scrollOffset = m.cursor - visible + 1
	}

	// Calculate scrollbar thumb position
	sbTrackH := 0
	sbThumbPos := 0
	if len(repos) > visible {
		sbTrackH = visible - 2
		if sbTrackH < 1 {
			sbTrackH = 1
		}
		sbThumbPos = m.scrollOffset * sbTrackH / max(1, len(repos)-visible)
	}

	nameW := 0
	for _, r := range repos {
		if len(r.Name) > nameW {
			nameW = len(r.Name)
		}
	}
	nameW = max(nameW, 4)

	var listContent strings.Builder
	if m.scrollOffset > 0 {
		listContent.WriteString(muted.Render(fmt.Sprintf("  ↑ %d more above", m.scrollOffset)) + "\n")
	}

	end := m.scrollOffset + visible
	if end > len(repos) {
		end = len(repos)
	}
	for i := m.scrollOffset; i < end; i++ {
		r := repos[i]
		dot := cleanDot
		if r.Dirty {
			dot = dirtyDot
		}

		paddedName := r.Name
		if len(paddedName) < nameW {
			paddedName += strings.Repeat(" ", nameW-len(paddedName))
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
			tags = append(tags, untrackStyle(fmt.Sprintf("?%d", r.Untracked)))
		}
		if r.Dirty {
			tags = append(tags, dirtyTag)
		}
		tagStr := ""
		if len(tags) > 0 {
			tagStr = "  " + strings.Join(tags, " ")
		}

		timeStr := ""
		if r.LastCommit != "" {
			timeStr = muted.Render("  " + r.LastCommit)
		}

		ptr := " "
		if i == m.cursor {
			ptr = "▶"
		}
		line := fmt.Sprintf("%s %s  %s  %s  %s%s%s", ptr, dot, paddedName, muted.Render(r.Dir), branch, timeStr, tagStr)
		// scrollbar char
		sbChar := " │"
		if len(repos) > visible && sbTrackH > 0 {
			relPos := (i - m.scrollOffset) * sbTrackH / visible
			if relPos >= sbThumbPos && relPos <= sbThumbPos {
				sbChar = " █"
			}
		}
		if i == m.cursor {
			line = cursorStyle.Render(line) + muted.Render(sbChar)
		} else if m.visualMode {
			vs, ve := m.visualRange()
			if i >= vs && i <= ve {
				line = lipgloss.NewStyle().
					Background(lipgloss.Color("236")).
					Render(line) + muted.Render(sbChar)
			} else {
				line = line + muted.Render(sbChar)
			}
		} else {
			line = line + muted.Render(sbChar)
		}
		listContent.WriteString(line + "\n")
	}

	if end < len(repos) {
		listContent.WriteString(muted.Render(fmt.Sprintf("  ↓ %d more below", len(repos)-end)) + "\n")
	}

result := header + "\n" + slogan + "\n" + sep + "\n" + listContent.String() + diffBlock + "\n"
	if m.filtering {
		result += footer + "\n"
	} else {
		result += footer + "  " + footerSep + "\n" + actionBar + "\n"
	}
	return result
}
