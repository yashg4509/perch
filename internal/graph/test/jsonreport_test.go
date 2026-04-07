package graph_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/yashg4509/perch/internal/config"
	"github.com/yashg4509/perch/internal/graph"
	"github.com/yashg4509/perch/internal/provider"
	"github.com/yashg4509/perch/internal/testutil"
)

func TestNewJSONReport_Golden(t *testing.T) {
	const yamlDoc = `name: golden-graph
environments:
  production:
    api: { provider: openai }
    web: { provider: vercel, project: demo }
edges:
  - web -> api
`
	cfg, err := config.Load([]byte(yamlDoc))
	if err != nil {
		t.Fatal(err)
	}
	reg, err := provider.LoadRegistry(filepath.Join(testutil.RepoRoot(t), "providers"))
	if err != nil {
		t.Fatal(err)
	}
	g, err := graph.Build(cfg, reg, "production")
	if err != nil {
		t.Fatal(err)
	}
	rep := graph.NewJSONReport(g)
	out, err := json.Marshal(rep)
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile(filepath.Join("..", "testdata", "graph.json"))
	if err != nil {
		t.Fatal(err)
	}
	var gotObj, wantObj any
	if err := json.Unmarshal(out, &gotObj); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(want, &wantObj); err != nil {
		t.Fatal(err)
	}
	aj, _ := json.Marshal(gotObj)
	bj, _ := json.Marshal(wantObj)
	if string(aj) != string(bj) {
		t.Fatalf("JSON mismatch\ngot:  %s\nwant: %s", out, want)
	}
}
