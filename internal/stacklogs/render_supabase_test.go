package stacklogs

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/yashg4509/perch/internal/config"
	"github.com/yashg4509/perch/internal/provider"
	"github.com/yashg4509/perch/internal/testutil"
)

func TestRenderAuthFilePath(t *testing.T) {
	restore := swapPlatformHooks(t)
	defer restore()
	tmp := t.TempDir()
	platformHooks.userHomeDir = func() (string, error) { return tmp, nil }
	got := renderAuthFilePath()
	want := filepath.Join(tmp, ".config", "render", "config.yaml")
	if got != want {
		t.Fatalf("path=%q want=%q", got, want)
	}
}

func TestReadRenderAuthToken(t *testing.T) {
	restore := swapPlatformHooks(t)
	defer restore()

	tmp := t.TempDir()
	authPath := filepath.Join(tmp, ".config", "render", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(authPath), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(authPath, []byte("api_key: render_tok\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	platformHooks.userHomeDir = func() (string, error) { return tmp, nil }
	platformHooks.readFile = os.ReadFile

	tok, ok := readRenderAuthToken()
	if !ok || tok != "render_tok" {
		t.Fatalf("token=%q ok=%v", tok, ok)
	}
}

func TestSupabaseAuthFilePath(t *testing.T) {
	restore := swapPlatformHooks(t)
	defer restore()
	tmp := t.TempDir()
	platformHooks.userHomeDir = func() (string, error) { return tmp, nil }
	got := supabaseAuthFilePath()
	want := filepath.Join(tmp, ".supabase", "access-token")
	if got != want {
		t.Fatalf("path=%q want=%q", got, want)
	}
}

func TestReadSupabaseAuthToken(t *testing.T) {
	restore := swapPlatformHooks(t)
	defer restore()

	tmp := t.TempDir()
	authPath := filepath.Join(tmp, ".supabase", "access-token")
	if err := os.MkdirAll(filepath.Dir(authPath), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(authPath, []byte("sbp_test_token\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	platformHooks.userHomeDir = func() (string, error) { return tmp, nil }
	platformHooks.readFile = os.ReadFile

	tok, ok := readSupabaseAuthToken()
	if !ok || tok != "sbp_test_token" {
		t.Fatalf("token=%q ok=%v", tok, ok)
	}
}

func TestResolve_Render_AuthFile(t *testing.T) {
	restore := swapPlatformHooks(t)
	defer restore()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer render_tok" {
			t.Fatalf("auth %q", r.Header.Get("Authorization"))
		}
		switch {
		case strings.HasSuffix(r.URL.Path, "/services/srv-test"):
			_, _ = io.WriteString(w, `{"id":"srv-test","ownerId":"own-test"}`)
		case r.URL.Path == "/logs" || strings.HasSuffix(r.URL.Path, "/logs"):
			if r.URL.Query().Get("ownerId") != "own-test" {
				t.Fatalf("ownerId=%q", r.URL.Query().Get("ownerId"))
			}
			if r.URL.Query().Get("resource") != "srv-test" {
				t.Fatalf("resource=%q", r.URL.Query().Get("resource"))
			}
			if r.URL.Query().Get("limit") != "100" {
				t.Fatalf("limit=%q", r.URL.Query().Get("limit"))
			}
			_, _ = io.WriteString(w, `{"logs":[{"message":"render log line"}]}`)
		default:
			t.Fatalf("path %s", r.URL.Path)
		}
	}))
	t.Cleanup(srv.Close)

	tmp := t.TempDir()
	authPath := filepath.Join(tmp, ".config", "render", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(authPath), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(authPath, []byte("api_key: render_tok\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	platformHooks.userHomeDir = func() (string, error) { return tmp, nil }
	platformHooks.readFile = os.ReadFile
	platformHooks.getenv = func(string) string { return "" }

	reg := renderRegistry(t, srv.URL)
	got, err := Resolve(context.Background(), "api", config.Node{Provider: "render", Service: "srv-test"}, reg)
	if err != nil {
		t.Fatal(err)
	}
	if got.Source != "auth_file" {
		t.Fatalf("source=%q", got.Source)
	}
	if len(got.Lines) != 1 || got.Lines[0] != "render log line" {
		t.Fatalf("lines=%v", got.Lines)
	}
}

func TestResolve_Supabase_AuthFile(t *testing.T) {
	restore := swapPlatformHooks(t)
	defer restore()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer sbp_test" {
			t.Fatalf("auth %q", r.Header.Get("Authorization"))
		}
		if !strings.HasSuffix(r.URL.Path, "/projects/ref-test/logs") {
			t.Fatalf("path %s", r.URL.Path)
		}
		if r.URL.Query().Get("sql") != "SELECT * FROM postgres_logs LIMIT 100" {
			t.Fatalf("sql=%q", r.URL.Query().Get("sql"))
		}
		_, _ = io.WriteString(w, `{"result":[{"event_message":"postgres says hi"}]}`)
	}))
	t.Cleanup(srv.Close)

	tmp := t.TempDir()
	authPath := filepath.Join(tmp, ".supabase", "access-token")
	if err := os.MkdirAll(filepath.Dir(authPath), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(authPath, []byte("sbp_test"), 0o600); err != nil {
		t.Fatal(err)
	}
	platformHooks.userHomeDir = func() (string, error) { return tmp, nil }
	platformHooks.readFile = os.ReadFile
	platformHooks.getenv = func(string) string { return "" }

	reg := supabaseRegistry(t, srv.URL)
	got, err := Resolve(context.Background(), "db", config.Node{Provider: "supabase", Project: "ref-test"}, reg)
	if err != nil {
		t.Fatal(err)
	}
	if got.Source != "auth_file" {
		t.Fatalf("source=%q", got.Source)
	}
	if len(got.Lines) != 1 || got.Lines[0] != "postgres says hi" {
		t.Fatalf("lines=%v", got.Lines)
	}
}

func TestResolve_Render_None_setupHint(t *testing.T) {
	restore := swapPlatformHooks(t)
	defer restore()

	platformHooks.readFile = func(string) ([]byte, error) { return nil, os.ErrNotExist }
	platformHooks.getenv = func(string) string { return "" }
	reg := renderRegistry(t, "http://127.0.0.1:1")
	got, err := Resolve(context.Background(), "api", config.Node{Provider: "render", Service: "srv-test"}, reg)
	if err != nil {
		t.Fatal(err)
	}
	if got.Source != "none" {
		t.Fatalf("source=%q", got.Source)
	}
	spec := reg.ByName["render"]
	want := buildSetupHint(spec)
	if got.SetupHint != want {
		t.Fatalf("hint=%q want=%q", got.SetupHint, want)
	}
	if !strings.Contains(got.SetupHint, "dashboard.render.com") {
		t.Fatalf("missing dashboard url in hint: %q", got.SetupHint)
	}
}

func TestRenderAuthFilePath_linuxConfig(t *testing.T) {
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		t.Skip("home-based auth paths on linux/darwin only")
	}
	restore := swapPlatformHooks(t)
	defer restore()
	tmp := t.TempDir()
	platformHooks.userHomeDir = func() (string, error) { return tmp, nil }
	got := renderAuthFilePath()
	want := filepath.Join(tmp, ".config", "render", "config.yaml")
	if got != want {
		t.Fatalf("path=%q want=%q", got, want)
	}
}

func renderRegistry(t *testing.T, baseURL string) *provider.Registry {
	t.Helper()
	root := testutil.RepoRoot(t)
	raw, err := os.ReadFile(filepath.Join(root, "providers", "hosting", "render.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	spec, err := provider.ParseProviderYAML(raw)
	if err != nil {
		t.Fatal(err)
	}
	spec.API.BaseURL = baseURL
	return &provider.Registry{ByName: map[string]*provider.Spec{"render": spec}}
}

func supabaseRegistry(t *testing.T, baseURL string) *provider.Registry {
	t.Helper()
	root := testutil.RepoRoot(t)
	raw, err := os.ReadFile(filepath.Join(root, "providers", "hosting", "supabase.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	spec, err := provider.ParseProviderYAML(raw)
	if err != nil {
		t.Fatal(err)
	}
	spec.API.BaseURL = baseURL
	return &provider.Registry{ByName: map[string]*provider.Spec{"supabase": spec}}
}
