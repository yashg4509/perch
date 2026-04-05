package config

import (
	"fmt"
	"os"
	"path/filepath"
)

const perchFileName = "perch.yaml"

// FindPerchYAML walks upward from startDir and returns the absolute path to the first perch.yaml found.
func FindPerchYAML(startDir string) (string, error) {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return "", fmt.Errorf("config: perch.yaml search: %w", err)
	}
	for {
		cand := filepath.Join(dir, perchFileName)
		st, err := os.Stat(cand)
		if err == nil && !st.IsDir() {
			return cand, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("config: perch.yaml not found from %q", startDir)
		}
		dir = parent
	}
}

// LoadNearest finds perch.yaml upward from startDir, then [Load]s it.
func LoadNearest(startDir string) (*Config, error) {
	p, err := FindPerchYAML(startDir)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return nil, fmt.Errorf("config: read %q: %w", p, err)
	}
	return Load(data)
}
