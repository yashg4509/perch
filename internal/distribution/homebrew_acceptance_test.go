package distribution

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHomebrewFormula(t *testing.T) {
	t.Helper()
	root := findRepoRoot(t)
	formula := filepath.Join(root, "homebrew-perch", "Formula", "perch.rb")
	b, err := os.ReadFile(formula)
	if err != nil {
		t.Fatalf("read formula: %v", err)
	}
	s := string(b)
	for _, needle := range []string{
		"class Perch",
		"github.com/yashg4509/perch",
		"cmd/perch",
	} {
		if !strings.Contains(s, needle) {
			t.Errorf("perch.rb must contain %q", needle)
		}
	}
}

func TestReadmeInstallPointer(t *testing.T) {
	t.Helper()
	root := findRepoRoot(t)
	readme := filepath.Join(root, "README.md")
	b, err := os.ReadFile(readme)
	if err != nil {
		t.Fatalf("read README: %v", err)
	}
	s := string(b)
	if !strings.Contains(s, "yashg4509/perch") {
		t.Error("README must point at canonical repo yashg4509/perch")
	}
	if !strings.Contains(s, "brew tap") && !strings.Contains(s, "Homebrew") {
		t.Error("README must mention Homebrew install path (tap or Homebrew)")
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
