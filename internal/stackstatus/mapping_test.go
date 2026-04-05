package stackstatus

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/yashg4509/perch/internal/provider"
)

func statusFixture(t *testing.T, name string) []byte {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	// stackstatus -> repo root -> internal/provider/testdata/status
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
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
	nr := nodeReportFromStatus("backend", "render", st)
	if nr.Healthy || nr.Provider != "render" {
		t.Fatalf("%+v", nr)
	}
	if nr.LastDeploy == nil || nr.LastDeploy.SHA != "d9e1f3a" {
		t.Fatalf("%+v", nr.LastDeploy)
	}
}
