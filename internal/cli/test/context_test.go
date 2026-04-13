package cli_test

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yashg4509/perch/internal/cli"
)

const contextTestYAML = `name: golden-app
environments:
  production:
    api: { provider: openai }
    web: { provider: vercel, project: demo }
edges:
  - web -> api
`

func TestContext_JSON_embeddedProviders(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "perch.yaml"), []byte(contextTestYAML), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(tmp)

	var buf bytes.Buffer
	cmd := cli.NewRootCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"context", "--json", "--env", "production"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var got map[string]any
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("invalid json: %q", buf.String())
	}
	if got["stack"] != "golden-app" || got["environment"] != "production" {
		t.Fatalf("%v", got)
	}
	nodes, _ := got["nodes"].([]any)
	if len(nodes) != 2 {
		t.Fatalf("nodes = %v", nodes)
	}
}

func TestContext_forAgent_text(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "perch.yaml"), []byte(contextTestYAML), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(tmp)

	var buf bytes.Buffer
	cmd := cli.NewRootCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"context", "--for-agent", "--env", "production"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	s := buf.String()
	for _, needle := range []string{
		"Stack: golden-app",
		"Environment: production",
		"api (openai, read-only): healthy",
		"web (vercel, deployable): unhealthy",
		"Summary: 1 of 2 nodes unhealthy",
	} {
		if !strings.Contains(s, needle) {
			t.Fatalf("missing %q in:\n%s", needle, s)
		}
	}
}

func TestContext_jsonAndForAgent_rejected(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "perch.yaml"), []byte(contextTestYAML), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(tmp)

	cmd := cli.NewRootCmd()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"context", "--json", "--for-agent"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error")
	}
}
