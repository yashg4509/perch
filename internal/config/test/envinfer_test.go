package config_test

import (
	"testing"

	"github.com/yashg4509/perch/internal/config"
	"github.com/yashg4509/perch/internal/provider"
)

func TestInferProject_pineconeAlias(t *testing.T) {
	spec := &provider.Spec{
		Name:              "pinecone",
		ProjectEnvAliases: []string{"PINECONE_INDEX"},
	}
	got := config.InferProject("pinecone", spec, config.Node{}, map[string]string{
		"PINECONE_INDEX": "my-index",
	})
	if got != "my-index" {
		t.Fatalf("got %q", got)
	}
}

func TestEffectiveProject_placeholder(t *testing.T) {
	spec := &provider.Spec{
		Name:              "pinecone",
		ProjectEnvAliases: []string{"PINECONE_INDEX"},
	}
	n := config.Node{Provider: "pinecone", Project: "CHANGE_ME"}
	got := config.EffectiveProject(n, spec, map[string]string{"PINECONE_INDEX": "idx-1"})
	if got != "idx-1" {
		t.Fatalf("got %q", got)
	}
}

func TestEffectiveProject_nodeOverride(t *testing.T) {
	spec := &provider.Spec{ProjectEnvAliases: []string{"PINECONE_INDEX"}}
	n := config.Node{Provider: "pinecone", Project: "CHANGE_ME", EnvProject: "CUSTOM_INDEX"}
	got := config.EffectiveProject(n, spec, map[string]string{
		"PINECONE_INDEX": "ignored",
		"CUSTOM_INDEX":   "custom",
	})
	if got != "custom" {
		t.Fatalf("got %q", got)
	}
}

func TestInferNeonFromDatabaseURL(t *testing.T) {
	got := config.InferProject("neon", &provider.Spec{Name: "neon"}, config.Node{}, map[string]string{
		"DATABASE_URL": "postgresql://u:p@ep-cool-name.us-east-2.aws.neon.tech/neondb",
	})
	if got != "ep-cool-name" {
		t.Fatalf("got %q", got)
	}
}

func TestApplyEnvInference(t *testing.T) {
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
	reg := &provider.Registry{
		ByName: map[string]*provider.Spec{
			"pinecone": {
				Name:              "pinecone",
				ProjectEnvAliases: []string{"PINECONE_INDEX"},
			},
		},
	}
	applied := config.ApplyEnvInference(cfg, "dev", map[string]string{"PINECONE_INDEX": "from-env"}, reg)
	if len(applied) != 1 {
		t.Fatalf("applied=%v", applied)
	}
	if cfg.Environments["dev"]["vectors"].Project != "from-env" {
		t.Fatalf("project=%q", cfg.Environments["dev"]["vectors"].Project)
	}
}

func TestProviderConfigured_pineconeWithEnv(t *testing.T) {
	spec := &provider.Spec{
		Name:              "pinecone",
		ProjectEnvAliases: []string{"PINECONE_INDEX"},
		Credentials:       provider.CredentialsSpec{Key: "pinecone_api_key"},
	}
	n := config.Node{Provider: "pinecone", Project: "CHANGE_ME"}
	ok, _ := config.ProviderConfigured(n, spec, map[string]string{"PINECONE_INDEX": "idx"}, func(string) (bool, error) {
		return true, nil
	})
	if !ok {
		t.Fatal("expected configured")
	}
}
