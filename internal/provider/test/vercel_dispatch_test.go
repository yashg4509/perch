package provider_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/yashg4509/perch/internal/provider"
	"github.com/yashg4509/perch/internal/testutil"
)

// T8-002: providers/hosting/vercel.yaml drives real DoGETJSON path (httptest only, no network).
func TestVercelYAML_statusEndpoint_dispatch(t *testing.T) {
	root := testutil.RepoRoot(t)
	raw, err := os.ReadFile(filepath.Join(root, "providers", "hosting", "vercel.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	spec, err := provider.ParseProviderYAML(raw)
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
	err = provider.DoGETJSON(ctx, http.DefaultClient, spec, "status", map[string]string{
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
