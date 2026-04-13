package releasedoc_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yashg4509/perch/internal/testutil"
)

func TestGoReleaserConfig(t *testing.T) {
	t.Helper()
	root := testutil.RepoRoot(t)
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
