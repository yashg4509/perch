package graph

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/yashg4509/perch/internal/config"
	"github.com/yashg4509/perch/internal/provider"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func TestBuild_deployableAndReadOnly(t *testing.T) {
	const yamlDoc = `
name: app
environments:
  production:
    fe:
      provider: vercel
      project: my-app
    llm:
      provider: openai
edges:
  - fe -> llm
`
	cfg, err := config.Load([]byte(yamlDoc))
	if err != nil {
		t.Fatal(err)
	}
	root := repoRoot(t)
	reg, err := provider.LoadRegistry(filepath.Join(root, "providers"))
	if err != nil {
		t.Fatal(err)
	}
	g, err := Build(cfg, reg, "production")
	if err != nil {
		t.Fatal(err)
	}
	if g.AppName != "app" || g.Environment != "production" {
		t.Fatalf("%+v", g)
	}
	if len(g.Edges) != 1 || g.Edges[0].From != "fe" || g.Edges[0].To != "llm" {
		t.Fatalf("%+v", g.Edges)
	}
	var fe, llm *Node
	for i := range g.Nodes {
		switch g.Nodes[i].Name {
		case "fe":
			fe = &g.Nodes[i]
		case "llm":
			llm = &g.Nodes[i]
		}
	}
	if fe == nil || llm == nil {
		t.Fatalf("nodes %+v", g.Nodes)
	}
	if !fe.Deployable || fe.Provider != "vercel" || fe.Project != "my-app" {
		t.Fatalf("fe %+v", *fe)
	}
	if llm.Deployable || llm.Provider != "openai" {
		t.Fatalf("llm %+v", *llm)
	}
}

func TestBuild_unknownProvider(t *testing.T) {
	cfg, err := config.Load([]byte(`
name: x
environments:
  production:
    a:
      provider: totally-unknown-provider
edges: []
`))
	if err != nil {
		t.Fatal(err)
	}
	reg := &provider.Registry{ByName: map[string]*provider.Spec{}}
	_, err = Build(cfg, reg, "production")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestBuild_customIsDeployable(t *testing.T) {
	cfg, err := config.Load([]byte(`
name: x
environments:
  production:
    api:
      provider: custom
      status: "exit 0"
edges: []
`))
	if err != nil {
		t.Fatal(err)
	}
	reg := &provider.Registry{ByName: map[string]*provider.Spec{}}
	g, err := Build(cfg, reg, "production")
	if err != nil {
		t.Fatal(err)
	}
	if len(g.Nodes) != 1 || !g.Nodes[0].Deployable {
		t.Fatalf("%+v", g.Nodes)
	}
}
