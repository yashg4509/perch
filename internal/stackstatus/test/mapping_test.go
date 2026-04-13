package stackstatus_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yashg4509/perch/internal/provider"
	"github.com/yashg4509/perch/internal/stackstatus"
	"github.com/yashg4509/perch/internal/testutil"
)

func statusFixture(t *testing.T, name string) []byte {
	t.Helper()
	root := testutil.RepoRoot(t)
	p := filepath.Join(root, "internal", "provider", "testdata", "status", name)
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestNodeReportFromStatus_fixtureNoNetwork(t *testing.T) {
	st, err := provider.ParseStatusJSON(statusFixture(t, "render_backend.json"))
	if err != nil {
		t.Fatal(err)
	}
	nr := stackstatus.NodeReportFromStatus("backend", "render", st)
	if nr.Healthy || nr.Provider != "render" {
		t.Fatalf("%+v", nr)
	}
	if nr.LastDeploy == nil || nr.LastDeploy.SHA != "d9e1f3a" {
		t.Fatalf("%+v", nr.LastDeploy)
	}
}
