package stackstatus_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/yashg4509/perch/internal/config"
	"github.com/yashg4509/perch/internal/provider"
	"github.com/yashg4509/perch/internal/stackstatus"
	"github.com/yashg4509/perch/internal/testutil"
)

func nodeConfigured(rep *stackstatus.EnvReport, name string) bool {
	for _, n := range rep.Nodes {
		if n.Name == name {
			return n.Configured
		}
	}
	return false
}

func TestCollect_InngestDevFromProjectEnv(t *testing.T) {
	ctx := context.Background()
	root := testutil.RepoRoot(t)
	reg, err := provider.LoadRegistry(filepath.Join(root, "providers"))
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load([]byte(`name: t
environments:
  dev:
    jobs: { provider: inngest, project: local-dev }
`))
	if err != nil {
		t.Fatal(err)
	}
	without, err := stackstatus.Collect(ctx, cfg, "dev", reg, stackstatus.CollectOptions{ProbeAPI: false})
	if err != nil {
		t.Fatal(err)
	}
	if nodeConfigured(without, "jobs") {
		t.Fatal("expected jobs unconfigured without INNGEST_DEV")
	}
	with, err := stackstatus.Collect(ctx, cfg, "dev", reg, stackstatus.CollectOptions{
		ProjectEnv: map[string]string{"INNGEST_DEV": "1"},
		ProbeAPI:   false,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !nodeConfigured(with, "jobs") {
		t.Fatal("expected jobs configured when INNGEST_DEV=1")
	}
}

func TestCollect_NeonFromDatabaseURL(t *testing.T) {
	ctx := context.Background()
	root := testutil.RepoRoot(t)
	reg, err := provider.LoadRegistry(filepath.Join(root, "providers"))
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load([]byte(`name: t
environments:
  dev:
    postgres: { provider: neon, project: local-neon }
`))
	if err != nil {
		t.Fatal(err)
	}
	with, err := stackstatus.Collect(ctx, cfg, "dev", reg, stackstatus.CollectOptions{
		ProjectEnv: map[string]string{
			"DATABASE_URL": "postgresql://u:p@ep-cool-name.us-east-2.aws.neon.tech/neondb",
		},
		ProbeAPI: false,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !nodeConfigured(with, "postgres") {
		t.Fatal("expected postgres configured when DATABASE_URL is Neon")
	}
}
