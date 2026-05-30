package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config holds the user configuration for gitmap.
type Config struct {
	ScanPaths []string `yaml:"scan_paths"`
	AutoFetch bool     `yaml:"auto_fetch"`
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
		return Default(), nil
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

// expandTilde replaces a leading ~ with the user's home directory.
func expandTilde(path string) string {
	if len(path) > 0 && path[0] == '~' {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[1:])
	}
	return path
}
