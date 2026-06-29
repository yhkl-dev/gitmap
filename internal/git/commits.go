package git

import (
	"bytes"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// DayStats holds commit and line-change counts for a single day.
type DayStats struct {
	Commits      int
	LinesChanged int
}

// CommitsDates runs git log --all on the repo and returns a map of date
// (YYYY-MM-DD) to commit count, looking back sinceDays from now.
// If authors is non-empty, only commits matching those authors are counted.
func CommitsDates(path string, sinceDays int, authors []string) (map[string]int, error) {
	since := time.Now().AddDate(0, 0, -sinceDays).Format("2006-01-02")

	args := []string{"-C", path, "log",
		"--since=" + since,
		"--format=%ad",
		"--date=short",
		"--all",
	}
	for _, a := range authors {
		args = append(args, "--author="+a)
	}
	cmd := exec.Command("git", args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	result := make(map[string]int)
	for _, line := range strings.Split(strings.TrimSpace(out.String()), "\n") {
		date := strings.TrimSpace(line)
		if date != "" {
			result[date]++
		}
	}
	return result, nil
}

// CommitsStats runs git log --all --numstat and returns per-day commit
// counts and lines changed in a single pass.
// If authors is non-empty, only commits matching those authors are counted.
func CommitsStats(path string, sinceDays int, authors []string) (map[string]DayStats, error) {
	since := time.Now().AddDate(0, 0, -sinceDays).Format("2006-01-02")

	args := []string{"-C", path, "log",
		"--since=" + since,
		"--format=%ad",
		"--date=short",
		"--all",
		"--numstat",
	}
	for _, a := range authors {
		args = append(args, "--author="+a)
	}
	cmd := exec.Command("git", args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	result := make(map[string]DayStats)
	var curDate string

	for _, line := range strings.Split(out.String(), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// date line: YYYY-MM-DD
		if len(line) == 10 && line[4] == '-' && line[7] == '-' {
			curDate = line
			s := result[curDate]
			s.Commits++
			result[curDate] = s
			continue
		}

		// numstat line: insertions\tab\deletions\tab\filename
		if curDate != "" {
			parts := strings.Split(line, "\t")
			if len(parts) < 2 {
				continue
			}
			ins, err1 := strconv.Atoi(parts[0])
			del, err2 := strconv.Atoi(parts[1])
			if err1 != nil || err2 != nil {
				continue // binary file ("-")
			}
			s := result[curDate]
			s.LinesChanged += ins + del
			result[curDate] = s
		}
	}
	return result, nil
}
