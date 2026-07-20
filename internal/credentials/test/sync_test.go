package credentials_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yashg4509/perch/internal/credentials"
	"github.com/yashg4509/perch/internal/provider"
)

func TestImportEnvFile_importsAndSkipsExisting(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	if err := os.WriteFile(envPath, []byte("OPENAI_API_KEY=sk-from-env\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	credPath := filepath.Join(dir, ".perch", "credentials")
	store := credentials.NewStoreAt(credPath)
	if err := store.Set("clerk_secret_key", "existing"); err != nil {
		t.Fatal(err)
	}

	specs := []provider.CredentialsSpec{
		{Key: "openai_api_key", EnvAliases: []string{"OPENAI_API_KEY"}},
		{Key: "clerk_secret_key", EnvAliases: []string{"CLERK_SECRET_KEY"}},
	}
	res, err := credentials.ImportEnvFile(store, envPath, specs, credentials.ImportOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Imported) != 1 || res.Imported[0] != "openai_api_key" {
		t.Fatalf("imported=%v", res.Imported)
	}
	if len(res.Skipped) != 0 {
		t.Fatalf("skipped=%v", res.Skipped)
	}
	v, ok, err := store.Get("openai_api_key")
	if err != nil || !ok || v != "sk-from-env" {
		t.Fatalf("openai=%q ok=%v err=%v", v, ok, err)
	}
	v, ok, _ = store.Get("clerk_secret_key")
	if !ok || v != "existing" {
		t.Fatalf("clerk should remain existing, got %q", v)
	}
}

func TestImportEnvFile_overwrite(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	if err := os.WriteFile(envPath, []byte("OPENAI_API_KEY=new\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	store := credentials.NewStoreAt(filepath.Join(dir, ".perch", "credentials"))
	_ = store.Set("openai_api_key", "old")

	specs := []provider.CredentialsSpec{{Key: "openai_api_key", EnvAliases: []string{"OPENAI_API_KEY"}}}
	res, err := credentials.ImportEnvFile(store, envPath, specs, credentials.ImportOptions{Overwrite: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Imported) != 1 {
		t.Fatalf("imported=%v", res.Imported)
	}
	v, _, _ := store.Get("openai_api_key")
	if v != "new" {
		t.Fatalf("got %q", v)
	}
}

func TestImportEnvFile_dryRun(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	if err := os.WriteFile(envPath, []byte("OPENAI_API_KEY=sk\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	store := credentials.NewStoreAt(filepath.Join(dir, ".perch", "credentials"))
	specs := []provider.CredentialsSpec{{Key: "openai_api_key", EnvAliases: []string{"OPENAI_API_KEY"}}}
	res, err := credentials.ImportEnvFile(store, envPath, specs, credentials.ImportOptions{DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Imported) != 1 {
		t.Fatalf("imported=%v", res.Imported)
	}
	_, ok, _ := store.Get("openai_api_key")
	if ok {
		t.Fatal("dry run should not write")
	}
}
