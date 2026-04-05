package config

import (
	"strings"
	"testing"
)

const minimalValidPerchYAML = `
name: test-app
environments:
  production:
    frontend:
      provider: vercel
      project: my-app
    backend:
      provider: render
      service: my-api
edges:
  - frontend -> backend
`

func TestParse_minimalValidPerchYAML(t *testing.T) {
	c, err := Parse([]byte(minimalValidPerchYAML))
	if err != nil {
		t.Fatal(err)
	}
	if c.Name != "test-app" {
		t.Fatalf("Name = %q", c.Name)
	}
	env, ok := c.Environments["production"]
	if !ok {
		t.Fatal("missing production environment")
	}
	fe, ok := env["frontend"]
	if !ok || fe.Provider != "vercel" || fe.Project != "my-app" {
		t.Fatalf("frontend node: %+v", fe)
	}
	be, ok := env["backend"]
	if !ok || be.Provider != "render" || be.Service != "my-api" {
		t.Fatalf("backend node: %+v", be)
	}
	if len(c.Edges) != 1 {
		t.Fatalf("Edges = %#v", c.Edges)
	}
	if c.Edges[0].From != "frontend" || c.Edges[0].To != "backend" {
		t.Fatalf("edge = %+v", c.Edges[0])
	}
}

func TestParse_rejectsInvalidEdgeToken(t *testing.T) {
	yaml := strings.Replace(minimalValidPerchYAML, "frontend -> backend", "frontend bad backend", 1)
	_, err := Parse([]byte(yaml))
	if err == nil {
		t.Fatal("expected error for malformed edge")
	}
}
