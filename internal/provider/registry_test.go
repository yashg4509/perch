package provider

import (
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
	dir := filepath.Dir(file)
	return filepath.Clean(filepath.Join(dir, "..", ".."))
}

func TestLoadRegistry_skipsUnderscoreYAML(t *testing.T) {
	root := repoRoot(t)
	dir := filepath.Join(root, "providers")
	reg, err := LoadRegistry(dir)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := reg.ByName["_template"]; ok {
		t.Fatal("underscore-prefixed YAML must not be registered")
	}
	if _, ok := reg.ByName["template"]; ok {
		t.Fatal("template should not appear without underscore as name key")
	}
}

func TestLoadRegistry_typedDeployableAndReadOnly(t *testing.T) {
	root := repoRoot(t)
	reg, err := LoadRegistry(filepath.Join(root, "providers"))
	if err != nil {
		t.Fatal(err)
	}
	v, ok := reg.ByName["vercel"]
	if !ok {
		t.Fatal("missing vercel")
	}
	if !v.Deployable {
		t.Fatal("vercel should be deployable")
	}
	if v.CLI == nil || v.CLI.Binary != "vercel" {
		t.Fatalf("cli: %+v", v.CLI)
	}
	if v.CLI.Commands["status"] == "" {
		t.Fatal("missing status command")
	}
	if v.API.BaseURL == "" || v.API.Endpoints["status"] == "" {
		t.Fatalf("api: %+v", v.API)
	}
	if v.Credentials.Key != "vercel_token" {
		t.Fatal(v.Credentials.Key)
	}

	o, ok := reg.ByName["openai"]
	if !ok {
		t.Fatal("missing openai")
	}
	if o.Deployable {
		t.Fatal("openai is read-only")
	}
	if o.CLI != nil {
		t.Fatal("read-only provider should omit cli in fixture")
	}
	if o.API.Endpoints["status"] == "" {
		t.Fatal("missing status endpoint")
	}
}

func TestParseProviderYAML_deployableRequiresCLIStatus(t *testing.T) {
	_, err := ParseProviderYAML([]byte(`
name: bad
category: x
deployable: true
cli:
  binary: x
  commands:
    logs: "x"
api:
  base_url: "https://x"
  auth_header: "A: {token}"
  endpoints:
    status: "GET /s"
credentials:
  key: k
  prompt: p
`))
	if err == nil {
		t.Fatal("expected error: deployable requires cli.commands.status")
	}
}

func TestParseProviderYAML_readOnlyRequiresAPIStatus(t *testing.T) {
	_, err := ParseProviderYAML([]byte(`
name: bad
category: x
deployable: false
api:
  base_url: "https://x"
  auth_header: "A: {token}"
  endpoints:
    usage: "GET /u"
credentials:
  key: k
  prompt: p
`))
	if err == nil {
		t.Fatal("expected error: api.endpoints.status required")
	}
}
