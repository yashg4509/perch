package provider_test

import (
	"path/filepath"
	"testing"

	"github.com/yashg4509/perch/internal/provider"
	"github.com/yashg4509/perch/internal/testutil"
	rootproviders "github.com/yashg4509/perch/providers"
)

func TestLoadRegistryFS_embedded(t *testing.T) {
	reg, err := provider.LoadRegistryFS(rootproviders.Files())
	if err != nil {
		t.Fatal(err)
	}
	if reg.ByName["vercel"] == nil || reg.ByName["openai"] == nil {
		t.Fatalf("missing specs: %v", reg.ByName)
	}
}

func TestLoadRegistryFS_matchesDiskProvidersDir(t *testing.T) {
	root := testutil.RepoRoot(t)
	fromDir, err := provider.LoadRegistry(filepath.Join(root, "providers"))
	if err != nil {
		t.Fatal(err)
	}
	fromFS, err := provider.LoadRegistryFS(rootproviders.Files())
	if err != nil {
		t.Fatal(err)
	}
	if len(fromDir.ByName) != len(fromFS.ByName) {
		t.Fatalf("dir=%d fs=%d", len(fromDir.ByName), len(fromFS.ByName))
	}
	for k, spec := range fromDir.ByName {
		if fromFS.ByName[k] == nil || fromFS.ByName[k].Name != spec.Name {
			t.Fatalf("mismatch on %q", k)
		}
	}
}
