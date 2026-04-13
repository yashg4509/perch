package distribution_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yashg4509/perch/internal/testutil"
)

func TestHomebrewFormula(t *testing.T) {
	t.Helper()
	root := testutil.RepoRoot(t)
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
	root := testutil.RepoRoot(t)
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
