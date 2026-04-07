package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yashg4509/perch/internal/config"
)

func TestFindPerchYAML_findsParent(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "svc", "pkg")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	perchPath := filepath.Join(root, "perch.yaml")
	if err := os.WriteFile(perchPath, []byte(minimalValidPerchYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := config.FindPerchYAML(nested)
	if err != nil {
		t.Fatal(err)
	}
	if got != perchPath {
		t.Fatalf("FindPerchYAML = %q, want %q", got, perchPath)
	}
}

func TestFindPerchYAML_prefersStartingDir(t *testing.T) {
	root := t.TempDir()
	inner := filepath.Join(root, "inner")
	if err := os.MkdirAll(inner, 0o755); err != nil {
		t.Fatal(err)
	}
	outerFile := filepath.Join(root, "perch.yaml")
	innerFile := filepath.Join(inner, "perch.yaml")
	if err := os.WriteFile(outerFile, []byte("name: outer\nenvironments:\n  production:\n    a:\n      provider: x\nedges: []\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(innerFile, []byte(minimalValidPerchYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := config.FindPerchYAML(inner)
	if err != nil {
		t.Fatal(err)
	}
	if got != innerFile {
		t.Fatalf("want inner perch.yaml, got %q", got)
	}
}

func TestFindPerchYAML_notFound(t *testing.T) {
	root := t.TempDir()
	_, err := config.FindPerchYAML(root)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoadNearest_chdirIntegration(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "deep")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "perch.yaml"), []byte(minimalValidPerchYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Chdir(nested)
	c, err := config.LoadNearest(".")
	if err != nil {
		t.Fatal(err)
	}
	if c.Name != "test-app" {
		t.Fatal(c.Name)
	}
}
