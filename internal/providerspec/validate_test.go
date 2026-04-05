package providerspec

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	// internal/providerspec/validate_test.go -> repo root is ../..
	dir := filepath.Dir(file)
	return filepath.Clean(filepath.Join(dir, "..", ".."))
}

func TestValidateProviderYAML_rejectsInvalidFixture(t *testing.T) {
	root := repoRoot(t)
	p := filepath.Join(root, "internal", "providerspec", "testdata", "invalid_missing_name.yaml")
	data, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateProviderYAML(data); err == nil {
		t.Fatal("expected validation error for invalid_missing_name.yaml")
	}
}

func TestValidateProviderYAML_acceptsTemplate(t *testing.T) {
	root := repoRoot(t)
	p := filepath.Join(root, "providers", "_template.yaml")
	data, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateProviderYAML(data); err != nil {
		t.Fatalf("template should validate: %v", err)
	}
}

func TestValidateAllProviderFilesUnderProviders(t *testing.T) {
	root := repoRoot(t)
	providersDir := filepath.Join(root, "providers")
	if err := ValidateProviderYAMLDir(providersDir); err != nil {
		t.Fatal(err)
	}
}
