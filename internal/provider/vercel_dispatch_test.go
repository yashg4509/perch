package provider

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// T8-002: providers/vercel.yaml drives real DoGETJSON path (httptest only, no network).
func TestVercelYAML_statusEndpoint_dispatch(t *testing.T) {
	root := repoRoot(t)
	raw, err := os.ReadFile(filepath.Join(root, "providers", "vercel.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	spec, err := ParseProviderYAML(raw)
	if err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v9/projects/acme" {
			t.Fatalf("path %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"healthy":true,"name":"acme"}`)
	}))
	t.Cleanup(srv.Close)

	spec.API.BaseURL = srv.URL
	ctx := context.Background()
	var got map[string]any
	err = DoGETJSON(ctx, http.DefaultClient, spec, "status", map[string]string{
		"token":   "tok",
		"project": "acme",
	}, &got)
	if err != nil {
		t.Fatal(err)
	}
	if got["healthy"] != true {
		t.Fatalf("%v", got)
	}
}
