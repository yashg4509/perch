package cidoc_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yashg4509/perch/internal/testutil"
)

func TestGitHubActionsWorkflow(t *testing.T) {
	t.Helper()
	root := testutil.RepoRoot(t)
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
