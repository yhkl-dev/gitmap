package scanner

import (
	"os"
	"path/filepath"
	"strings"
)

// Repo represents a discovered git repository.
type Repo struct {
	Path string
	Name string
}

// Directories to skip during scanning.
var ignoredDirs = map[string]bool{
	"node_modules": true,
	".venv":        true,
	"venv":         true,
	"__pycache__":  true,
	".cache":       true,
	"vendor":       true,
}

// Scan walks the given paths, discovers .git directories, and returns
// a deduplicated list of Repo entries. Hidden directories (except .git
// itself) and well-known package directories are skipped.
func Scan(paths []string) ([]Repo, error) {
	var repos []Repo
	seen := map[string]bool{}

	for _, root := range paths {
		_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // skip inaccessible paths
			}
			if !info.IsDir() {
				return nil
			}
			name := filepath.Base(path)
			if ignoredDirs[name] || (strings.HasPrefix(name, ".") && name != ".git") {
				return filepath.SkipDir
			}
			if name == ".git" {
				parent := filepath.Dir(path)
				if !seen[parent] {
					seen[parent] = true
					repos = append(repos, Repo{
						Path: parent,
						Name: filepath.Base(parent),
					})
				}
				return filepath.SkipDir
			}
			return nil
		})
	}
	return repos, nil
}
