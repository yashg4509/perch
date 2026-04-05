package provider

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
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

	spec := &Spec{
		Name: "openai",
		API: APISpec{
			BaseURL:    srv.URL,
			AuthHeader: "Authorization: Bearer {token}",
			Endpoints:  map[string]string{"status": "GET /v1/models"},
		},
	}
	ctx := context.Background()
	var out map[string]any
	if err := DoGETJSON(ctx, http.DefaultClient, spec, "status", map[string]string{"token": "tok"}, &out); err != nil {
		t.Fatal(err)
	}
	if out["ok"] != true {
		t.Fatalf("%v", out)
	}
}

func TestDecodeJSON(t *testing.T) {
	var ns NodeStatus
	if err := DecodeJSON([]byte(`{"healthy":true}`), &ns); err != nil {
		t.Fatal(err)
	}
	if !ns.Healthy {
		t.Fatal(ns)
	}
}

func TestReadOnlyStub_Status(t *testing.T) {
	root := repoRoot(t)
	reg, err := LoadRegistry(filepath.Join(root, "providers"))
	if err != nil {
		t.Fatal(err)
	}
	spec := reg.ByName["openai"]
	stub := NewReadOnlyStub(spec, ReadOnlyStubOptions{
		StatusBody: []byte(`{"healthy":true,"daily_tokens":142}`),
	})
	ctx := context.Background()
	ns, err := stub.Status(ctx, Node{Name: "llm", Provider: "openai", Fields: map[string]string{}})
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
	root := repoRoot(t)
	reg, err := LoadRegistry(filepath.Join(root, "providers"))
	if err != nil {
		t.Fatal(err)
	}
	stub := NewReadOnlyStub(reg.ByName["openai"], ReadOnlyStubOptions{})
	ctx := context.Background()
	_, err = stub.Deploy(ctx, Node{})
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrUnsupported) {
		t.Fatalf("err = %v", err)
	}
}
