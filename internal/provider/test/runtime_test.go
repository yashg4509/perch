package provider_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yashg4509/perch/internal/provider"
	"github.com/yashg4509/perch/internal/testutil"
)

func TestDoGETJSON_roundTrip(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ah := r.Header.Get("Authorization"); ah != "Bearer tok" {
			t.Errorf("Authorization = %q", ah)
		}
		if !strings.HasSuffix(r.URL.Path, "/v1/models") {
			t.Errorf("path = %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"ok":true,"id":"m1"}`)
	}))
	t.Cleanup(srv.Close)

	spec := &provider.Spec{
		Name: "openai",
		API: provider.APISpec{
			BaseURL:    srv.URL,
			AuthHeader: "Authorization: Bearer {token}",
			Endpoints:  map[string]string{"status": "GET /v1/models"},
		},
	}
	ctx := context.Background()
	var out map[string]any
	if err := provider.DoGETJSON(ctx, http.DefaultClient, spec, "status", map[string]string{"token": "tok"}, &out); err != nil {
		t.Fatal(err)
	}
	if out["ok"] != true {
		t.Fatalf("%v", out)
	}
}

func TestDecodeJSON(t *testing.T) {
	var ns provider.NodeStatus
	if err := provider.DecodeJSON([]byte(`{"healthy":true}`), &ns); err != nil {
		t.Fatal(err)
	}
	if !ns.Healthy {
		t.Fatal(ns)
	}
}

func TestReadOnlyStub_Status(t *testing.T) {
	root := testutil.RepoRoot(t)
	reg, err := provider.LoadRegistry(filepath.Join(root, "providers"))
	if err != nil {
		t.Fatal(err)
	}
	spec := reg.ByName["openai"]
	stub := provider.NewReadOnlyStub(spec, provider.ReadOnlyStubOptions{
		StatusBody: []byte(`{"healthy":true,"daily_tokens":142}`),
	})
	ctx := context.Background()
	ns, err := stub.Status(ctx, provider.Node{Name: "llm", Provider: "openai", Fields: map[string]string{}})
	if err != nil {
		t.Fatal(err)
	}
	if !ns.Healthy {
		t.Fatal(ns)
	}
	if ns.DailyTokens == nil || *ns.DailyTokens != 142 {
		t.Fatalf("daily_tokens %+v", ns.DailyTokens)
	}
}

func TestReadOnlyStub_DeployUnsupported(t *testing.T) {
	root := testutil.RepoRoot(t)
	reg, err := provider.LoadRegistry(filepath.Join(root, "providers"))
	if err != nil {
		t.Fatal(err)
	}
	stub := provider.NewReadOnlyStub(reg.ByName["openai"], provider.ReadOnlyStubOptions{})
	ctx := context.Background()
	_, err = stub.Deploy(ctx, provider.Node{})
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, provider.ErrUnsupported) {
		t.Fatalf("err = %v", err)
	}
}
