package main

import (
	"os/exec"
	"strings"
	"sync"

	"github.com/charmbracelet/bubbletea"

	gitpkg "github.com/yhkl-dev/gitmap/internal/git"
	"github.com/yhkl-dev/gitmap/internal/config"
	"github.com/yhkl-dev/gitmap/internal/scanner"
)

func loadPRsCmd(path string) tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("gh", "pr", "list", "--limit", "20")
		cmd.Dir = path
		out, err := cmd.Output()
		if err != nil {
			return prsLoadedMsg("")
		}
		s := strings.TrimSpace(string(out))
		if s == "" {
			s = "(no open PRs)"
		}
		return prsLoadedMsg(s)
	}
}

func singleFetchCmd(path string) tea.Cmd {
	return func() tea.Msg {
		gitpkg.Fetch(path)
		return fetchDoneMsg{single: true}
	}
}

func batchFetchCmd(repos []gitpkg.RepoStatus) tea.Cmd {
	return func() tea.Msg {
		for _, r := range repos {
			gitpkg.Fetch(r.Path)
		}
		return fetchDoneMsg{single: false}
	}
}

func singlePullCmd(path string) tea.Cmd {
	return func() tea.Msg {
		gitpkg.Pull(path)
		return pullDoneMsg{single: true}
	}
}

func batchPullCmd(repos []gitpkg.RepoStatus) tea.Cmd {
	return func() tea.Msg {
		for _, r := range repos {
			gitpkg.Pull(r.Path)
		}
		return pullDoneMsg{single: false}
	}
}

func loadReposCmd(scanPaths []string, excludeRepos []string) tea.Cmd {
	return func() tea.Msg {
		repos, err := scanner.Scan(scanPaths)
		if err != nil {
			return reposLoadedMsg{errors: -1}
		}

		if len(excludeRepos) > 0 {
			var filtered []scanner.Repo
			for _, r := range repos {
				if !config.MatchesAny(r.Name, excludeRepos) {
					filtered = append(filtered, r)
				}
			}
			repos = filtered
		}

		if len(repos) == 0 {
			return reposLoadedMsg{}
		}

		const maxConcurrent = 8
		sem := make(chan struct{}, maxConcurrent)
		var wg sync.WaitGroup
		var errMu sync.Mutex
		results := make([]gitpkg.RepoStatus, len(repos))
		var errCount int

		for i := range repos {
			wg.Add(1)
			go func(idx int, repo scanner.Repo) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				s, err := gitpkg.CachedStatus(repo.Path)
				if err != nil {
					s = &gitpkg.RepoStatus{Path: repo.Path}
					errMu.Lock()
					errCount++
					errMu.Unlock()
				}
				s.Name = repo.Name
				results[idx] = *s
			}(i, repos[i])
		}
		wg.Wait()
		return reposLoadedMsg{repos: results, errors: errCount}
	}
}
