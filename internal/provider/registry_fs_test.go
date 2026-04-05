package provider

import (
	"path/filepath"
	"runtime"
	"testing"

	rootproviders "github.com/yashg4509/perch/providers"
)

func repoRootFS(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func TestLoadRegistryFS_embedded(t *testing.T) {
	reg, err := LoadRegistryFS(rootproviders.Files())
	if err != nil {
		t.Fatal(err)
	}
	if reg.ByName["vercel"] == nil || reg.ByName["openai"] == nil {
		t.Fatalf("missing specs: %v", reg.ByName)
	}
}

func TestLoadRegistryFS_matchesDiskProvidersDir(t *testing.T) {
	root := repoRootFS(t)
	fromDir, err := LoadRegistry(filepath.Join(root, "providers"))
	if err != nil {
		t.Fatal(err)
	}
	fromFS, err := LoadRegistryFS(rootproviders.Files())
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
