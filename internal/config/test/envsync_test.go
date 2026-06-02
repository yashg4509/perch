package config_test

import (
	"testing"

	"github.com/yashg4509/perch/internal/config"
)

func TestApplyEnvMap(t *testing.T) {
	cfg, err := config.Load([]byte(`name: x
environments:
  dev:
    vectors:
      provider: pinecone
      project: CHANGE_ME
edges: []
`))
	if err != nil {
		t.Fatal(err)
	}
	applied, err := config.ApplyEnvMap(cfg, "dev", map[string]string{
		"PINECONE_INDEX": "my-index",
	}, map[string]string{
		"vectors.project": "PINECONE_INDEX",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(applied) != 1 {
		t.Fatalf("applied=%v", applied)
	}
	if cfg.Environments["dev"]["vectors"].Project != "my-index" {
		t.Fatalf("project=%q", cfg.Environments["dev"]["vectors"].Project)
	}
}

func TestApplyDevDefaults(t *testing.T) {
	cfg, err := config.Load([]byte(`name: x
environments:
  dev:
    billing:
      provider: stripe
      project: CHANGE_ME
edges: []
`))
	if err != nil {
		t.Fatal(err)
	}
	config.ApplyDevDefaults(cfg, "dev")
	if cfg.Environments["dev"]["billing"].Project != "local-dev" {
		t.Fatalf("got %q", cfg.Environments["dev"]["billing"].Project)
	}
}
