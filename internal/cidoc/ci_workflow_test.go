package cidoc

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGitHubActionsWorkflow(t *testing.T) {
	t.Helper()
	root := findRepoRoot(t)
	p := filepath.Join(root, ".github", "workflows", "ci.yml")
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read CI workflow: %v", err)
	}
	s := string(b)
	for _, needle := range []string{
		"go test",
		"go vet",
		"gofmt",
		"govulncheck",
		"gosec",
		"exclude=G304",
		"GOTOOLCHAIN",
		"pull_request",
		"push",
		"main",
	} {
		if !strings.Contains(s, needle) {
			t.Errorf("ci.yml should reference %q", needle)
		}
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
