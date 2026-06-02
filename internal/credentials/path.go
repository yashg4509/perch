package credentials

import (
	"os"
	"path/filepath"
	"strings"
)

// DefaultPath returns $HOME/.perch/credentials (or PERCH_CREDENTIALS_PATH if set).
func DefaultPath() (string, error) {
	if p := strings.TrimSpace(os.Getenv("PERCH_CREDENTIALS_PATH")); p != "" {
		return p, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".perch", "credentials"), nil
}
