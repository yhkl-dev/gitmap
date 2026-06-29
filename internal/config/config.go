package config

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config holds the user configuration for gitmap.
type Config struct {
	ScanPaths    []string `yaml:"scan_paths"`
	AutoFetch    bool     `yaml:"auto_fetch"`
	ExcludeRepos []string `yaml:"exclude_repos"`
	Author       string   `yaml:"author"`
}

// Default returns a config with ~/projects as the default scan path.
func Default() *Config {
	home, _ := os.UserHomeDir()
	return &Config{
		ScanPaths: []string{filepath.Join(home, "projects")},
	}
}

// Load reads config from path, falling back to Default() if the file
// does not exist.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Default(), nil
		}
		return nil, err
	}
	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	for i, p := range cfg.ScanPaths {
		cfg.ScanPaths[i] = expandTilde(p)
	}
	if len(cfg.ScanPaths) == 0 {
		return Default(), nil
	}
	return cfg, nil
}

// MatchesAny reports whether name matches any of the given filepath.Match patterns.
func MatchesAny(name string, patterns []string) bool {
	for _, pat := range patterns {
		if ok, _ := filepath.Match(pat, name); ok {
			return true
		}
	}
	return false
}

// IsExcluded returns true if name matches any pattern in the exclude list.
func (c *Config) IsExcluded(name string) bool {
	return MatchesAny(name, c.ExcludeRepos)
}

// expandTilde replaces a leading ~ with the user's home directory.
func expandTilde(path string) string {
	if len(path) > 0 && path[0] == '~' {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[1:])
	}
	return path
}
