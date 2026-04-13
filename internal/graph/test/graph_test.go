package graph_test

import (
	"path/filepath"
	"testing"

	"github.com/yashg4509/perch/internal/config"
	"github.com/yashg4509/perch/internal/graph"
	"github.com/yashg4509/perch/internal/provider"
	"github.com/yashg4509/perch/internal/testutil"
)

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
	root := testutil.RepoRoot(t)
	reg, err := provider.LoadRegistry(filepath.Join(root, "providers"))
	if err != nil {
		t.Fatal(err)
	}
	g, err := graph.Build(cfg, reg, "production")
	if err != nil {
		t.Fatal(err)
	}
	if g.AppName != "app" || g.Environment != "production" {
		t.Fatalf("%+v", g)
	}
	if len(g.Edges) != 1 || g.Edges[0].From != "fe" || g.Edges[0].To != "llm" {
		t.Fatalf("%+v", g.Edges)
	}
	var fe, llm *graph.Node
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
	_, err = graph.Build(cfg, reg, "production")
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
	g, err := graph.Build(cfg, reg, "production")
	if err != nil {
		t.Fatal(err)
	}
	if len(g.Nodes) != 1 || !g.Nodes[0].Deployable {
		t.Fatalf("%+v", g.Nodes)
	}
}
