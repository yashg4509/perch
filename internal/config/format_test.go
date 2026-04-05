package config

import (
	"reflect"
	"testing"
)

func TestFormatYAML_roundtrip(t *testing.T) {
	c, err := Load([]byte(minimalValidPerchYAML))
	if err != nil {
		t.Fatal(err)
	}
	b, err := FormatYAML(c)
	if err != nil {
		t.Fatal(err)
	}
	c2, err := Load(b)
	if err != nil {
		t.Fatalf("reload: %v\n%s", err, string(b))
	}
	if c2.Name != c.Name {
		t.Fatalf("name %q vs %q", c2.Name, c.Name)
	}
	if !reflect.DeepEqual(c2.Environments, c.Environments) {
		t.Fatalf("environments differ")
	}
	if !reflect.DeepEqual(c2.Edges, c.Edges) {
		t.Fatalf("edges %#v vs %#v", c2.Edges, c.Edges)
	}
}

func TestFormatYAML_emptyEdges(t *testing.T) {
	c, err := Load([]byte(`name: x
environments:
  production:
    a:
      provider: vercel
      project: p
edges: []
`))
	if err != nil {
		t.Fatal(err)
	}
	b, err := FormatYAML(c)
	if err != nil {
		t.Fatal(err)
	}
	c2, err := Load(b)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if len(c2.Edges) != 0 {
		t.Fatalf("edges = %#v", c2.Edges)
	}
}

func TestAddEdge_idempotent(t *testing.T) {
	c, err := Load([]byte(minimalValidPerchYAML))
	if err != nil {
		t.Fatal(err)
	}
	if err := AddEdge(c, "frontend", "backend"); err != nil {
		t.Fatal(err)
	}
	if err := AddEdge(c, "frontend", "backend"); err != nil {
		t.Fatal(err)
	}
	if len(c.Edges) != 1 {
		t.Fatal(c.Edges)
	}
}

func TestAddEdge_new(t *testing.T) {
	c, err := Load([]byte(`name: x
environments:
  production:
    a:
      provider: vercel
      project: p
    b:
      provider: openai
edges: []
`))
	if err != nil {
		t.Fatal(err)
	}
	if err := AddEdge(c, "a", "b"); err != nil {
		t.Fatal(err)
	}
	if len(c.Edges) != 1 || c.Edges[0].From != "a" || c.Edges[0].To != "b" {
		t.Fatalf("%#v", c.Edges)
	}
}

func TestAddEdge_unknownNode(t *testing.T) {
	c, err := Load([]byte(`name: x
environments:
  production:
    a:
      provider: vercel
      project: p
edges: []
`))
	if err != nil {
		t.Fatal(err)
	}
	if err := AddEdge(c, "a", "ghost"); err == nil {
		t.Fatal("expected error")
	}
}

func TestRemoveEdge(t *testing.T) {
	c, err := Load([]byte(minimalValidPerchYAML))
	if err != nil {
		t.Fatal(err)
	}
	n := RemoveEdge(c, "frontend", "backend")
	if n != 1 || len(c.Edges) != 0 {
		t.Fatalf("n=%d edges=%#v", n, c.Edges)
	}
}
