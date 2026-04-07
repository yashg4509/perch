package cli_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/yashg4509/perch/internal/cli"
	"github.com/yashg4509/perch/internal/config"
	"github.com/yashg4509/perch/internal/testutil"
)

func TestStatus_JSON_endToEnd(t *testing.T) {
	const yamlDoc = `name: golden-app
environments:
  production:
    api: { provider: openai }
    web: { provider: vercel, project: demo }
edges:
  - web -> api
`
	tmp := t.TempDir()
	repo := testutil.RepoRoot(t)
	prov := filepath.Join(repo, "providers")
	link := filepath.Join(tmp, "providers")
	if err := os.Symlink(prov, link); err != nil {
		t.Skip("symlink providers (need unix or admin):", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "perch.yaml"), []byte(yamlDoc), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Chdir(tmp)

	var buf bytes.Buffer
	cmd := cli.NewRootCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"status", "--json", "--env", "production"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var got map[string]any
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("stdout is not JSON: %q\nerr: %v", buf.String(), err)
	}
	if got["env"] != "production" {
		t.Fatalf("env = %v", got["env"])
	}
	nodes, _ := got["nodes"].([]any)
	if len(nodes) != 2 {
		t.Fatalf("nodes len = %d", len(nodes))
	}
}

func TestGraph_JSON_endToEnd(t *testing.T) {
	const yamlDoc = `name: golden-graph
environments:
  production:
    api: { provider: openai }
    web: { provider: vercel, project: demo }
edges:
  - web -> api
`
	tmp := t.TempDir()
	repo := testutil.RepoRoot(t)
	link := filepath.Join(tmp, "providers")
	if err := os.Symlink(filepath.Join(repo, "providers"), link); err != nil {
		t.Skip("symlink providers (need unix or admin):", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "perch.yaml"), []byte(yamlDoc), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(tmp)

	var buf bytes.Buffer
	cmd := cli.NewRootCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"graph", "--json", "--env", "production"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var got map[string]any
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("stdout is not JSON: %q", buf.String())
	}
	if got["appName"] != "golden-graph" {
		t.Fatalf("appName = %v", got["appName"])
	}
	edges, _ := got["edges"].([]any)
	if len(edges) != 1 {
		t.Fatalf("edges len = %d", len(edges))
	}
}

func TestInit_writesValidPerchYAML(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "vercel.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(tmp)

	var buf, errBuf bytes.Buffer
	cmd := cli.NewRootCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&errBuf)
	cmd.SetArgs([]string{"init", "--name", "cliapp", "--env", "production"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v stderr=%q", err, errBuf.String())
	}

	raw, err := os.ReadFile(filepath.Join(tmp, "perch.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	c, err := config.Load(raw)
	if err != nil {
		t.Fatalf("load generated: %v\n%s", err, raw)
	}
	if c.Name != "cliapp" {
		t.Fatalf("name = %q", c.Name)
	}
	if !bytes.Contains(buf.Bytes(), []byte("Wrote perch.yaml")) {
		t.Fatalf("stdout = %q", buf.String())
	}
}

func TestEdge_addListRemove(t *testing.T) {
	const yamlDoc = `name: edge-cli
environments:
  production:
    api: { provider: openai }
    web: { provider: vercel, project: demo }
edges: []
`
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "perch.yaml"), []byte(yamlDoc), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(tmp)

	var buf, errBuf bytes.Buffer
	root := cli.NewRootCmd()
	root.SetOut(&buf)
	root.SetErr(&errBuf)

	buf.Reset()
	errBuf.Reset()
	root.SetArgs([]string{"edge", "add", "web", "api"})
	if err := root.Execute(); err != nil {
		t.Fatalf("add: %v stderr=%q", err, errBuf.String())
	}
	raw, err := os.ReadFile(filepath.Join(tmp, "perch.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := config.Load(raw); err != nil {
		t.Fatalf("invalid yaml after add: %v\n%s", err, raw)
	}

	buf.Reset()
	errBuf.Reset()
	root.SetArgs([]string{"edge", "list"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(buf.Bytes(), []byte("web -> api")) {
		t.Fatalf("list: %q", buf.String())
	}

	buf.Reset()
	errBuf.Reset()
	root.SetArgs([]string{"edge", "rm", "web", "api"})
	if err := root.Execute(); err != nil {
		t.Fatalf("rm: %v stderr=%q", err, errBuf.String())
	}
	raw2, _ := os.ReadFile(filepath.Join(tmp, "perch.yaml"))
	c2, err := config.Load(raw2)
	if err != nil {
		t.Fatal(err)
	}
	if len(c2.Edges) != 0 {
		t.Fatalf("edges = %#v", c2.Edges)
	}

	root.SetArgs([]string{"edge", "rm", "web", "api"})
	if err := root.Execute(); err == nil {
		t.Fatal("expected error removing missing edge")
	}
}

func TestInit_JSON_output(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "vercel.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(tmp)

	var buf bytes.Buffer
	cmd := cli.NewRootCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"init", "--name", "j", "--env", "production", "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("not json: %q", buf.String())
	}
	if got["written"] != true {
		t.Fatalf("written = %v", got["written"])
	}
}
