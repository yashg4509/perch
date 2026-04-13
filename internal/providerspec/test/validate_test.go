package providerspec_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yashg4509/perch/internal/providerspec"
	"github.com/yashg4509/perch/internal/testutil"
)

func TestValidateProviderYAML_rejectsInvalidFixture(t *testing.T) {
	root := testutil.RepoRoot(t)
	p := filepath.Join(root, "internal", "providerspec", "testdata", "invalid_missing_name.yaml")
	data, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if err := providerspec.ValidateProviderYAML(data); err == nil {
		t.Fatal("expected validation error for invalid_missing_name.yaml")
	}
}

func TestValidateProviderYAML_acceptsTemplate(t *testing.T) {
	root := testutil.RepoRoot(t)
	p := filepath.Join(root, "providers", "_template.yaml")
	data, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if err := providerspec.ValidateProviderYAML(data); err != nil {
		t.Fatalf("template should validate: %v", err)
	}
}

func TestValidateAllProviderFilesUnderProviders(t *testing.T) {
	root := testutil.RepoRoot(t)
	providersDir := filepath.Join(root, "providers")
	if err := providerspec.ValidateProviderYAMLDir(providersDir); err != nil {
		t.Fatal(err)
	}
}
