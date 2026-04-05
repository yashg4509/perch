package config

import "testing"

func TestEnvironmentNodes_unknown(t *testing.T) {
	c := &Config{
		Name: "a",
		Environments: map[string]map[string]Node{
			"production": {"x": {Provider: "vercel", Project: "p"}},
		},
		Edges: nil,
	}
	_, err := c.EnvironmentNodes("staging")
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() == "" {
		t.Fatal("empty error")
	}
}

func TestEnvironmentNodes_emptyName(t *testing.T) {
	c := &Config{Name: "a", Environments: map[string]map[string]Node{"production": {}}}
	_, err := c.EnvironmentNodes("")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestEnvironmentNodes_ok(t *testing.T) {
	c := &Config{
		Environments: map[string]map[string]Node{
			"production": {"fe": {Provider: "vercel", Project: "p"}},
		},
	}
	m, err := c.EnvironmentNodes("production")
	if err != nil {
		t.Fatal(err)
	}
	if len(m) != 1 {
		t.Fatal(m)
	}
}
