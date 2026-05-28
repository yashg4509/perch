package cli

import (
	"os"
	"path/filepath"

	"github.com/yashg4509/perch/internal/config"
	"github.com/yashg4509/perch/internal/credentials"
	"github.com/yashg4509/perch/internal/provider"
	"github.com/yashg4509/perch/internal/stackstatus"
)

func loadCredStore() (*credentials.Store, error) {
	p, err := credentials.DefaultPath()
	if err != nil {
		return nil, err
	}
	return credentials.NewStore(p), nil
}

func credentialSpecsForEnv(cfg *config.Config, env string, reg *provider.Registry) []provider.CredentialsSpec {
	if cfg == nil || reg == nil {
		return nil
	}
	nodes, ok := cfg.Environments[env]
	if !ok {
		return nil
	}
	seen := map[string]bool{}
	var specs []provider.CredentialsSpec
	for _, n := range nodes {
		spec := reg.ByName[n.Provider]
		if spec == nil {
			continue
		}
		k := spec.Credentials.Key
		if k == "" || seen[k] {
			continue
		}
		seen[k] = true
		specs = append(specs, spec.Credentials)
	}
	return specs
}

func loadCollectOptions(perchPath string) stackstatus.CollectOptions {
	store, _ := loadCredStore()
	return stackstatus.CollectOptions{
		CredStore:        store,
		ProjectEnv:       loadProjectDotenv(perchPath),
		ProbeAPI:         true,
		ProbeConcurrency: 0,
	}
}

func loadProjectDotenv(perchPath string) map[string]string {
	envPath := filepath.Join(filepath.Dir(perchPath), ".env")
	raw, err := os.ReadFile(envPath)
	if err != nil {
		return map[string]string{}
	}
	m, err := credentials.ParseDotenv(string(raw))
	if err != nil {
		return map[string]string{}
	}
	return m
}
