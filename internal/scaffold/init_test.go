package scaffold

import (
	"os"
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

func TestGenerate_singleVercel(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "vercel.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	reg, err := provider.LoadRegistry(filepath.Join(repoRoot(t), "providers"))
	if err != nil {
		t.Fatal(err)
	}
	raw, inf, err := Generate(root, Options{AppName: "myapp", Env: "production", Registry: reg})
	if err != nil {
		t.Fatal(err)
	}
	if len(inf.NeedsPrompt) != 0 {
		t.Fatal(inf.NeedsPrompt)
	}
	c, err := config.Load(raw)
	if err != nil {
		t.Fatal(err)
	}
	if c.Name != "myapp" {
		t.Fatal(c.Name)
	}
}

func TestGenerate_supabaseEdge(t *testing.T) {
	root := t.TempDir()
	_ = os.WriteFile(filepath.Join(root, "fly.toml"), []byte(""), 0o644)
	_ = os.WriteFile(filepath.Join(root, "package.json"), []byte(`{
  "dependencies": { "@supabase/supabase-js": "^2.0.0" }
}`), 0o644)
	reg, err := provider.LoadRegistry(filepath.Join(repoRoot(t), "providers"))
	if err != nil {
		t.Fatal(err)
	}
	_, inf, err := Generate(root, Options{AppName: "s", Env: "production", Registry: reg})
	if err != nil {
		t.Fatal(err)
	}
	if len(inf.Edges) != 1 || inf.Edges[0].From != "backend" || inf.Edges[0].To != "db" {
		t.Fatalf("%+v", inf)
	}
}

func TestWriteIfChanged_idempotent(t *testing.T) {
	root := t.TempDir()
	_ = os.WriteFile(filepath.Join(root, "vercel.json"), []byte("{}"), 0o644)
	reg, err := provider.LoadRegistry(filepath.Join(repoRoot(t), "providers"))
	if err != nil {
		t.Fatal(err)
	}
	opt := Options{AppName: "x", Env: "production", Registry: reg}
	w1, _, err := WriteIfChanged(root, opt)
	if err != nil {
		t.Fatal(err)
	}
	if !w1 {
		t.Fatal("first write should create file")
	}
	w2, _, err := WriteIfChanged(root, opt)
	if err != nil {
		t.Fatal(err)
	}
	if w2 {
		t.Fatal("second write should be skipped when identical")
	}
}
