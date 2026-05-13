package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/go-git/go-git/v5"
)

// RepoStatus holds the lightweight status of a git repository.
type RepoStatus struct {
	Path   string
	Name   string
	Branch string
	Dirty  bool
}

// Status opens a git repository at the given path and extracts its
// current branch name and working-tree cleanliness. A repository with
// no commits returns an empty Branch and Dirty=false.
func Status(path string) (*RepoStatus, error) {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return nil, err
	}

	head, err := repo.Head()
	if err != nil {
		// No commits yet
		return &RepoStatus{Path: path, Dirty: false}, nil
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

	return &RepoStatus{
		Path:   path,
		Branch: branch,
		Dirty:  !ws.IsClean(),
	}, nil
}

// Diff returns a short diff summary for the repository at path.
// Uses the system's git binary.
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

// ShortHash returns the first 7 chars of the HEAD commit hash, or an
// empty string if there are no commits.
func ShortHash(path string) string {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return ""
	}
	head, err := repo.Head()
	if err != nil {
		return ""
	}
	h := head.Hash().String()
	if len(h) >= 7 {
		return h[:7]
	}
	return h
}

// BranchInfo returns a compact branch description like "feat/foo [dirty]".
func BranchInfo(s RepoStatus) string {
	b := s.Branch
	if b == "" {
		b = "(no commits)"
	}
	if s.Dirty {
		b += " [dirty]"
	}
	return b
}

// ModifiedFiles returns a list of modified filenames from the worktree status.
func ModifiedFiles(path string) []string {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return nil
	}
	wt, err := repo.Worktree()
	if err != nil {
		return nil
	}
	ws, err := wt.Status()
	if err != nil {
		return nil
	}
	var files []string
	for f := range ws {
		status := ws.File(f)
		staging := status.Staging
		worktree := status.Worktree
		// Show staged, modified, or untracked files
		if staging != ' ' && staging != '?' {
			files = append(files, fmt.Sprintf("%c%c %s", staging, worktree, f))
		} else if worktree != ' ' {
			files = append(files, fmt.Sprintf(" %c %s", worktree, f))
		}
	}
	return files
}
