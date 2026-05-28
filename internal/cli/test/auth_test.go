package cli_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/yashg4509/perch/internal/cli"
	"github.com/yashg4509/perch/internal/testutil"
)

func TestAuthSyncEnv_JSON(t *testing.T) {
	tmp := t.TempDir()
	repo := testutil.RepoRoot(t)
	if err := os.Symlink(filepath.Join(repo, "providers"), filepath.Join(tmp, "providers")); err != nil {
		t.Skip("symlink providers:", err)
	}
	envPath := filepath.Join(tmp, ".env")
	if err := os.WriteFile(envPath, []byte("OPENAI_API_KEY=sk-test\nANTHROPIC_API_KEY=sk-ant\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	credPath := filepath.Join(tmp, ".perch", "credentials")

	t.Chdir(tmp)

	var buf bytes.Buffer
	cmd := cli.NewRootCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{
		"auth", "sync-env",
		"--env-file", ".env",
		"--credentials", credPath,
		"--json",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var res struct {
		Imported []string `json:"imported"`
	}
	if err := json.Unmarshal(buf.Bytes(), &res); err != nil {
		t.Fatalf("stdout %q: %v", buf.String(), err)
	}
	found := false
	for _, k := range res.Imported {
		if k == "openai_api_key" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected openai_api_key imported, got %v", res.Imported)
	}
}
