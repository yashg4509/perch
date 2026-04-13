package config_test

import (
	"runtime"
	"testing"

	"github.com/yashg4509/perch/internal/config"
)

func TestLoad_validCombinesParseAndValidate(t *testing.T) {
	c, err := config.Load([]byte(minimalValidPerchYAML))
	if err != nil {
		t.Fatal(err)
	}
	if c.Name != "test-app" {
		t.Fatal(c.Name)
	}
}

func TestLoad_invalidYAML(t *testing.T) {
	_, err := config.Load([]byte("name: [\n"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoad_customProviderWithStatus(t *testing.T) {
	status := "exit 0"
	if runtime.GOOS == "windows" {
		status = "exit /b 0"
	}
	_, err := config.Load([]byte(`name: x
environments:
  production:
    svc:
      provider: custom
      status: "` + status + `"
edges: []
`))
	if err != nil {
		t.Fatal(err)
	}
}

func TestLoad_invalidSemantics(t *testing.T) {
	_, err := config.Load([]byte(`name: ""
environments:
  production:
    a:
      provider: x
edges: []
`))
	if err == nil {
		t.Fatal("expected validation error")
	}
}
