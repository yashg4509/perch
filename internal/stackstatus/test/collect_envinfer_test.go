package stackstatus_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/yashg4509/perch/internal/config"
	"github.com/yashg4509/perch/internal/credentials"
	"github.com/yashg4509/perch/internal/provider"
	"github.com/yashg4509/perch/internal/stackstatus"
	"github.com/yashg4509/perch/internal/testutil"
)

func TestCollect_PineconeProjectFromEnv(t *testing.T) {
	ctx := context.Background()
	root := testutil.RepoRoot(t)
	reg, err := provider.LoadRegistry(filepath.Join(root, "providers"))
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load([]byte(`name: t
environments:
  dev:
    vectors:
      provider: pinecone
      project: CHANGE_ME
`))
	if err != nil {
		t.Fatal(err)
	}
	opts := stackstatus.CollectOptions{ProbeAPI: false}
	without, err := stackstatus.Collect(ctx, cfg, "dev", reg, opts)
	if err != nil {
		t.Fatal(err)
	}
	if nodeConfigured(without, "vectors") {
		t.Fatal("expected unconfigured without env and credentials")
	}
	credPath := filepath.Join(t.TempDir(), "credentials.json")
	store := credentials.NewStore(credPath)
	if err := store.Set("pinecone_api_key", "test-key"); err != nil {
		t.Fatal(err)
	}
	with, err := stackstatus.Collect(ctx, cfg, "dev", reg, stackstatus.CollectOptions{
		ProjectEnv: map[string]string{"PINECONE_INDEX": "my-index"},
		CredStore:  store,
		ProbeAPI:   false,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !nodeConfigured(with, "vectors") {
		t.Fatal("expected configured when PINECONE_INDEX and credential are present")
	}
}
