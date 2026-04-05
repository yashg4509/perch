package stackstatus

import (
	"context"
	"runtime"
	"testing"

	"github.com/yashg4509/perch/internal/config"
	"github.com/yashg4509/perch/internal/provider"
)

func TestCollect_customProviderScript(t *testing.T) {
	ctx := context.Background()
	cmd := "true"
	if runtime.GOOS == "windows" {
		cmd = "exit /b 0"
	}
	yaml := `name: c
environments:
  production:
    svc:
      provider: custom
      status: "` + cmd + `"
edges: []
`
	cfg, err := config.Load([]byte(yaml))
	if err != nil {
		t.Fatal(err)
	}
	reg := &provider.Registry{ByName: map[string]*provider.Spec{}}
	got, err := Collect(ctx, cfg, "production", reg)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Nodes) != 1 || !got.Nodes[0].Healthy {
		t.Fatalf("%+v", got)
	}
}
