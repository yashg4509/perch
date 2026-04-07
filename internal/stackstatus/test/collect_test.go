package stackstatus_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/yashg4509/perch/internal/config"
	"github.com/yashg4509/perch/internal/provider"
	"github.com/yashg4509/perch/internal/stackstatus"
	"github.com/yashg4509/perch/internal/testutil"
)

func TestCollect_JSONGolden(t *testing.T) {
	ctx := context.Background()
	root := testutil.RepoRoot(t)
	reg, err := provider.LoadRegistry(filepath.Join(root, "providers"))
	if err != nil {
		t.Fatal(err)
	}

	const yamlDoc = `name: golden-app
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

	got, err := stackstatus.Collect(ctx, cfg, "production", reg)
	if err != nil {
		t.Fatal(err)
	}
	out, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}

	want, err := os.ReadFile(filepath.Join("..", "testdata", "status.json"))
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
	if !jsonEqual(gotObj, wantObj) {
		t.Fatalf("JSON mismatch\ngot:  %s\nwant: %s", out, want)
	}
}

func jsonEqual(a, b any) bool {
	aj, _ := json.Marshal(a)
	bj, _ := json.Marshal(b)
	return string(aj) == string(bj)
}
