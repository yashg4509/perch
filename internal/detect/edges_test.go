package detect

import (
	"reflect"
	"testing"
)

func TestInferEdges_supabaseAuto(t *testing.T) {
	nodes := map[string]string{
		"frontend": "vercel",
		"api":      "render",
		"db":       "supabase",
	}
	deps := []string{"@supabase/supabase-js"}
	inf := InferEdges(nodes, deps)
	want := []EdgePair{{From: "api", To: "db"}}
	if !reflect.DeepEqual(inf.Edges, want) {
		t.Fatalf("edges %+v want %+v", inf.Edges, want)
	}
	if len(inf.NeedsPrompt) != 0 {
		t.Fatalf("%+v", inf.NeedsPrompt)
	}
}

func TestInferEdges_supabaseDepMissingNode(t *testing.T) {
	nodes := map[string]string{"frontend": "vercel", "api": "render"}
	inf := InferEdges(nodes, []string{"@supabase/supabase-js"})
	if len(inf.Edges) != 0 {
		t.Fatal(inf.Edges)
	}
	if len(inf.NeedsPrompt) != 1 {
		t.Fatalf("%+v", inf.NeedsPrompt)
	}
}

func TestInferEdges_supabaseVercelOnly(t *testing.T) {
	nodes := map[string]string{"frontend": "vercel", "db": "supabase"}
	inf := InferEdges(nodes, []string{"@supabase/supabase-js"})
	want := []EdgePair{{From: "frontend", To: "db"}}
	if !reflect.DeepEqual(inf.Edges, want) {
		t.Fatalf("edges %+v want %+v", inf.Edges, want)
	}
	if len(inf.NeedsPrompt) != 0 {
		t.Fatalf("%+v", inf.NeedsPrompt)
	}
}

func TestInferEdges_initSignalsStyle(t *testing.T) {
	nodes := map[string]string{
		"db":       "supabase",
		"frontend": "vercel",
		"llm":      "openai",
		"payments": "stripe",
	}
	deps := []string{"@supabase/supabase-js", "openai", "stripe"}
	inf := InferEdges(nodes, deps)
	want := []EdgePair{
		{From: "frontend", To: "db"},
		{From: "frontend", To: "llm"},
		{From: "frontend", To: "payments"},
	}
	if !reflect.DeepEqual(inf.Edges, want) {
		t.Fatalf("edges %+v want %+v", inf.Edges, want)
	}
	if len(inf.NeedsPrompt) != 0 {
		t.Fatalf("%+v", inf.NeedsPrompt)
	}
}
