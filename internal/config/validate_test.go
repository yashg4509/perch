package config

import (
	"strings"
	"testing"
)

func TestValidate_table(t *testing.T) {
	cases := []struct {
		name    string
		yaml    string
		wantErr string
	}{
		{
			name: "missing_name",
			yaml: `
environments:
  production:
    a:
      provider: vercel
      project: p
edges: []
`,
			wantErr: "name",
		},
		{
			name: "no_environments",
			yaml: `
name: x
edges: []
`,
			wantErr: "environments",
		},
		{
			name: "empty_environment",
			yaml: `
name: x
environments:
  production: {}
edges: []
`,
			wantErr: "production",
		},
		{
			name: "node_missing_provider",
			yaml: `
name: x
environments:
  production:
    frontend:
      project: only
edges: []
`,
			wantErr: "provider",
		},
		{
			name: "edge_unknown_from",
			yaml: `
name: x
environments:
  production:
    frontend:
      provider: vercel
      project: p
edges:
  - ghost -> frontend
`,
			wantErr: "ghost",
		},
		{
			name: "custom_missing_status",
			yaml: `
name: x
environments:
  production:
    app:
      provider: custom
edges: []
`,
			wantErr: "custom",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c, err := Parse([]byte(tc.yaml))
			if err != nil {
				t.Fatal(err)
			}
			err = Validate(c)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected err: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("err %q should mention %q", err.Error(), tc.wantErr)
			}
		})
	}
}

func TestValidate_duplicateNodeNamesProgrammatic(t *testing.T) {
	c := &Config{
		Name: "x",
		Environments: map[string]map[string]Node{
			"production": {
				"frontend": {Provider: "vercel", Project: "p"},
				"Frontend": {Provider: "render", Service: "s"},
			},
		},
		Edges: nil,
	}
	err := Validate(c)
	if err == nil {
		t.Fatal("expected duplicate node name error")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "duplicate") {
		t.Fatalf("err = %v", err)
	}
}

func TestValidate_acceptsMinimalFromT1_001(t *testing.T) {
	c, err := Parse([]byte(minimalValidPerchYAML))
	if err != nil {
		t.Fatal(err)
	}
	if err := Validate(c); err != nil {
		t.Fatal(err)
	}
}
