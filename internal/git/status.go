package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/go-git/go-git/v5"
)

// RepoStatus holds the lightweight status of a git repository.
type RepoStatus struct {
	Path       string
	Name       string
	Branch     string
	Dirty      bool
	Ahead      int
	Behind     int
	StashCount int
	Untracked  int
	LastCommit string // relative time, e.g. "2 hours ago"
	RemoteURL  string
	Conflict   bool
	Dir        string
}
func shortDir(path string) string {
	i := strings.LastIndex(path, "/")
	if i <= 0 {
		return ""
	}
	parent := path[:i]
	j := strings.LastIndex(parent, "/")
	if j < 0 {
		return parent + "/"
	}
	return parent[j+1:] + "/"
}

// Status opens a git repository at the given path and extracts its status.
func Status(path string) (*RepoStatus, error) {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return nil, err
	}

	head, err := repo.Head()
	if err != nil {
		return &RepoStatus{
			Path:       path,
			Dirty:      false,
			StashCount: StashCount(path),
			RemoteURL:  RemoteURL(path),
			Conflict:   HasConflict(path),
			Dir:        shortDir(path),
		}, nil
	}

	branch := head.Name().Short()

	wt, err := repo.Worktree()
	if err != nil {
		return nil, err
	}

	ws, err := wt.Status()
	if err != nil {
		return nil, err
	}

	s := &RepoStatus{
		Path:       path,
		Branch:     branch,
		Dirty:      !ws.IsClean(),
		StashCount: StashCount(path),
		Untracked:  countUntracked(ws),
		LastCommit: LastCommitTime(path),
		RemoteURL:  RemoteURL(path),
		Conflict:   HasConflict(path),
		Dir:        shortDir(path),
	}
	if branch != "" {
		s.Ahead, s.Behind = AheadBehind(path, branch)
	}
	return s, nil
}

func countUntracked(ws git.Status) int {
	n := 0
	for _, entry := range ws {
		if entry.Staging == '?' && entry.Worktree == '?' {
			n++
		}
	}
	return n
}

// Diff returns a short diff summary for the repository at path.
func Diff(path string) string {
	cmd := exec.Command("git", "-C", path, "diff", "--stat")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return ""
	}
	result := strings.TrimSpace(out.String())
	if result == "" {
		return "(no changes)"
	}
	return result
}

// AheadBehind returns the number of commits ahead and behind upstream.
func AheadBehind(path, branch string) (int, int) {
	cmd := exec.Command("git", "-C", path, "rev-list", "--count", "--left-right", "HEAD...@{upstream}")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return 0, 0
	}
	parts := strings.Fields(strings.TrimSpace(out.String()))
	if len(parts) != 2 {
		return 0, 0
	}
	ahead, _ := strconv.Atoi(parts[0])
	behind, _ := strconv.Atoi(parts[1])
	return ahead, behind
}

// Fetch runs git fetch on the repository.
func Fetch(path string) error {
	cmd := exec.Command("git", "-C", path, "fetch")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	return cmd.Run()
}

// Pull runs git pull --ff-only on the repository.
func Pull(path string) error {
	cmd := exec.Command("git", "-C", path, "pull", "--ff-only")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	return cmd.Run()
}

// RecentCommits returns the last n commits in oneline format.
func RecentCommits(path string, n int) string {
	cmd := exec.Command("git", "-C", path, "log", "--oneline", "-n", strconv.Itoa(n))
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return ""
	}
	s := strings.TrimSpace(out.String())
	return s
}

// LastCommitTime returns the relative time of the last commit.
func LastCommitTime(path string) string {
	cmd := exec.Command("git", "-C", path, "log", "-1", "--format=%cr")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return ""
	}
	return strings.TrimSpace(out.String())
}

// StashCount returns the number of stashes.
func StashCount(path string) int {
	cmd := exec.Command("git", "-C", path, "stash", "list")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return 0
	}
	if strings.TrimSpace(out.String()) == "" {
		return 0
	}
	return strings.Count(strings.TrimSpace(out.String()), "\n") + 1
}

// StashList returns the list of stashes.
func StashList(path string) string {
	cmd := exec.Command("git", "-C", path, "stash", "list")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return ""
	}
	s := strings.TrimSpace(out.String())
	return s
}

// BranchList returns all local branches, current marked with *.
func BranchList(path string) string {
	cmd := exec.Command("git", "-C", path, "branch", "--sort=-committerdate")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return ""
	}
	s := strings.TrimSpace(out.String())
	return s
}

// StashApply runs "git stash apply" for the most recent stash.
func StashApply(path string) string {
	cmd := exec.Command("git", "-C", path, "stash", "apply")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return "error: " + err.Error()
	}
	return strings.TrimSpace(out.String())
}

// StashPop runs "git stash pop" for the most recent stash.
func StashPop(path string) string {
	cmd := exec.Command("git", "-C", path, "stash", "pop")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return "error: " + err.Error()
	}
	return strings.TrimSpace(out.String())
}

// StashDrop runs "git stash drop" for the most recent stash.
func StashDrop(path string) string {
	cmd := exec.Command("git", "-C", path, "stash", "drop")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return "error: " + err.Error()
	}
	return strings.TrimSpace(out.String())
}

// DetailedDiff returns a full unified diff, capped at 500 lines.
func DetailedDiff(path string) string {
	cmd := exec.Command("git", "-C", path, "diff", "-p", "--unified=3")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return ""
	}
	result := strings.TrimSpace(out.String())
	if result == "" {
		return "(no changes)"
	}
	lines := strings.Split(result, "\n")
	if len(lines) > 500 {
		result = strings.Join(lines[:500], "\n") +
			fmt.Sprintf("\n... (truncated, showing 500 of %d lines)", len(lines))
	}
	return result
}

// HasConflict returns true if the repo is in a merge, rebase, or cherry-pick state.
func HasConflict(path string) bool {
	for _, f := range []string{"rebase-merge", "rebase-apply", "MERGE_HEAD", "CHERRY_PICK_HEAD"} {
		if _, err := os.Stat(path + "/.git/" + f); err == nil {
			return true
		}
	}
	return false
}

// RemoteURL returns the URL of the "origin" remote.
func RemoteURL(path string) string {
	cmd := exec.Command("git", "-C", path, "remote", "get-url", "origin")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return ""
	}
	return strings.TrimSpace(out.String())
}

// HttpsRemote converts SSH remote URLs to HTTPS format for browser opening.
func HttpsRemote(url string) string {
	if strings.HasPrefix(url, "git@") {
		url = strings.Replace(url, ":", "/", 1)
		url = strings.Replace(url, "git@", "https://", 1)
	} else if strings.HasPrefix(url, "ssh://") {
		url = strings.Replace(url, "ssh://", "https://", 1)
	}
	return strings.TrimSuffix(url, ".git")
}

// Checkout runs "git checkout" for the given branch and returns output.
func Checkout(path, branch string) string {
	cmd := exec.Command("git", "-C", path, "checkout", branch)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return "error: " + strings.TrimSpace(out.String())
	}
	return strings.TrimSpace(out.String())
}
