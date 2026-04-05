package releasedoc

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGoReleaserConfig(t *testing.T) {
	t.Helper()
	root := findRepoRoot(t)
	p := filepath.Join(root, ".goreleaser.yaml")
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read .goreleaser.yaml: %v", err)
	}
	s := string(b)
	for _, needle := range []string{
		"project_name:",
		"perch",
		"./cmd/perch",
		"CGO_ENABLED=0",
	} {
		if !strings.Contains(s, needle) {
			t.Errorf(".goreleaser.yaml must contain %q", needle)
		}
	}
	if !strings.Contains(s, "brews:") && !strings.Contains(s, "homebrew-perch") {
		t.Error(".goreleaser.yaml should configure Homebrew tap publish (brews and/or homebrew-perch)")
	}
}

func findRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found; run tests from module root")
		}
		dir = parent
	}
}
