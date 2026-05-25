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
	"time"

	"github.com/yashg4509/perch/internal/config"
	"github.com/yashg4509/perch/internal/credentials"
	"github.com/yashg4509/perch/internal/provider"
	"github.com/yashg4509/perch/internal/testutil"
)

func TestResolve_Vercel_AuthFile(t *testing.T) {
	restore := swapPlatformHooks(t)
	defer restore()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/v6/deployments":
			_, _ = io.WriteString(w, `{"deployments":[{"uid":"dpl_test123"}]}`)
		case strings.HasPrefix(r.URL.Path, "/v2/deployments/dpl_test123/events"):
			_, _ = io.WriteString(w, `[{"type":"stdout","text":"hello from vercel"}]`)
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	t.Cleanup(srv.Close)

	tmp := t.TempDir()
	platformHooks.userHomeDir = func() (string, error) { return tmp, nil }
	authPath := vercelAuthFilePath()
	if authPath == "" {
		t.Skip("no vercel auth path on this OS")
	}
	if err := os.MkdirAll(filepath.Dir(authPath), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(authPath, []byte(`{"token":"vca_test"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	platformHooks.readFile = os.ReadFile
	platformHooks.getenv = func(string) string { return "" }
	reg := vercelRegistry(t, srv.URL)
	got, err := Resolve(context.Background(), "web", config.Node{Provider: "vercel", Project: "demo"}, reg)
	if err != nil {
		t.Fatal(err)
	}
	if got.Source != "auth_file" {
		t.Fatalf("source=%q", got.Source)
	}
	if len(got.Lines) != 1 || got.Lines[0] != "hello from vercel" {
		t.Fatalf("lines=%v", got.Lines)
	}
}

func TestResolve_Vercel_EnvToken(t *testing.T) {
	restore := swapPlatformHooks(t)
	defer restore()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer env_tok" {
			t.Fatalf("auth header %q", r.Header.Get("Authorization"))
		}
		switch r.URL.Path {
		case "/v6/deployments":
			_, _ = io.WriteString(w, `{"deployments":[{"uid":"dpl_env"}]}`)
		case "/v2/deployments/dpl_env/events":
			_, _ = io.WriteString(w, `[{"type":"stderr","text":"env log line"}]`)
		default:
			t.Fatalf("path %s", r.URL.Path)
		}
	}))
	t.Cleanup(srv.Close)

	platformHooks.readFile = func(string) ([]byte, error) { return nil, os.ErrNotExist }
	platformHooks.userHomeDir = func() (string, error) { return t.TempDir(), nil }
	platformHooks.getenv = func(k string) string {
		if k == "VERCEL_TOKEN" {
			return "env_tok"
		}
		return ""
	}
	reg := vercelRegistry(t, srv.URL)
	got, err := Resolve(context.Background(), "web", config.Node{Provider: "vercel", Project: "demo"}, reg)
	if err != nil {
		t.Fatal(err)
	}
	if got.Source != "env_token" {
		t.Fatalf("source=%q", got.Source)
	}
	if len(got.Lines) != 1 || got.Lines[0] != "env log line" {
		t.Fatalf("lines=%v", got.Lines)
	}
}

func TestResolve_Vercel_403Fallthrough(t *testing.T) {
	restore := swapPlatformHooks(t)
	defer restore()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		switch {
		case auth == "Bearer vca_bad":
			w.WriteHeader(http.StatusForbidden)
			return
		case auth == "Bearer env_tok":
			switch r.URL.Path {
			case "/v6/deployments":
				_, _ = io.WriteString(w, `{"deployments":[{"uid":"dpl_env"}]}`)
			case "/v2/deployments/dpl_env/events":
				_, _ = io.WriteString(w, `[{"type":"stdout","text":"env after 403 fallthrough"}]`)
			default:
				t.Fatalf("path %s", r.URL.Path)
			}
		default:
			t.Fatalf("unexpected auth %q path %s", auth, r.URL.Path)
		}
	}))
	t.Cleanup(srv.Close)

	tmp := t.TempDir()
	platformHooks.userHomeDir = func() (string, error) { return tmp, nil }
	authPath := vercelAuthFilePath()
	if authPath == "" {
		t.Skip("no vercel auth path on this OS")
	}
	if err := os.MkdirAll(filepath.Dir(authPath), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(authPath, []byte(`{"token":"vca_bad"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	platformHooks.readFile = os.ReadFile
	platformHooks.getenv = func(k string) string {
		if k == "VERCEL_TOKEN" {
			return "env_tok"
		}
		return ""
	}
	reg := vercelRegistry(t, srv.URL)
	got, err := Resolve(context.Background(), "web", config.Node{Provider: "vercel", Project: "demo"}, reg)
	if err != nil {
		t.Fatal(err)
	}
	if got.Source != "env_token" {
		t.Fatalf("source=%q want env_token", got.Source)
	}
	if len(got.Lines) != 1 || got.Lines[0] != "env after 403 fallthrough" {
		t.Fatalf("lines=%v", got.Lines)
	}
}

func TestResolve_Vercel_CredentialsStore(t *testing.T) {
	restore := swapPlatformHooks(t)
	defer restore()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer store_tok" {
			t.Fatalf("auth %q", r.Header.Get("Authorization"))
		}
		switch r.URL.Path {
		case "/v6/deployments":
			_, _ = io.WriteString(w, `{"deployments":[{"uid":"dpl_store"}]}`)
		case "/v2/deployments/dpl_store/events":
			_, _ = io.WriteString(w, `[{"type":"stdout","text":"from credentials store"}]`)
		default:
			t.Fatalf("path %s", r.URL.Path)
		}
	}))
	t.Cleanup(srv.Close)

	tmp := t.TempDir()
	credPath := filepath.Join(tmp, ".perch", "credentials")
	store := credentials.NewStoreAt(credPath)
	if err := store.Set("vercel_token", "store_tok"); err != nil {
		t.Fatal(err)
	}
	platformHooks.readFile = func(string) ([]byte, error) { return nil, os.ErrNotExist }
	platformHooks.userHomeDir = func() (string, error) { return tmp, nil }
	platformHooks.getenv = func(string) string { return "" }
	platformHooks.credentialsStore = func() *credentials.Store { return credentials.NewStoreAt(credPath) }

	reg := vercelRegistry(t, srv.URL)
	got, err := Resolve(context.Background(), "web", config.Node{Provider: "vercel", Project: "demo"}, reg)
	if err != nil {
		t.Fatal(err)
	}
	if got.Source != "credentials_store" {
		t.Fatalf("source=%q", got.Source)
	}
	if len(got.Lines) != 1 || got.Lines[0] != "from credentials store" {
		t.Fatalf("lines=%v", got.Lines)
	}
}

func TestResolve_Vercel_autoSetupAttempted_skipsAutoSetup(t *testing.T) {
	restore := swapPlatformHooks(t)
	defer restore()
	restoreSetup := swapSetupHooks(t)
	defer restoreSetup()

	var autoSetupCalls int
	setupHooks.lookPath = func(string) (string, error) { return "", errNotFound }
	setupHooks.runShell = func(context.Context, string, time.Duration) error {
		autoSetupCalls++
		return nil
	}

	reg := vercelRegistry(t, "http://127.0.0.1:1")
	got, err := resolveWithFlags(context.Background(), "web", config.Node{Provider: "vercel", Project: "demo"}, reg, true)
	if err != nil {
		t.Fatal(err)
	}
	if autoSetupCalls > 0 {
		t.Fatalf("auto_setup should be skipped, calls=%d", autoSetupCalls)
	}
	for _, s := range got.StrategiesTried {
		if s.Name == "auto_setup" {
			t.Fatalf("unexpected auto_setup in %v", got.StrategiesTried)
		}
	}
}

func TestResolve_Vercel_None(t *testing.T) {
	restore := swapPlatformHooks(t)
	defer restore()

	platformHooks.readFile = func(string) ([]byte, error) { return nil, os.ErrNotExist }
	platformHooks.userHomeDir = func() (string, error) { return t.TempDir(), nil }
	platformHooks.getenv = func(string) string { return "" }
	reg := vercelRegistry(t, "http://127.0.0.1:1")
	got, err := Resolve(context.Background(), "web", config.Node{Provider: "vercel", Project: "demo"}, reg)
	if err != nil {
		t.Fatal(err)
	}
	if got.Source != "none" {
		t.Fatalf("source=%q", got.Source)
	}
	spec := reg.ByName["vercel"]
	want := buildSetupHint(spec)
	if got.SetupHint != want {
		t.Fatalf("hint=%q want=%q", got.SetupHint, want)
	}
	if len(got.Lines) != 0 {
		t.Fatalf("lines=%v", got.Lines)
	}
}

func TestReadVercelAuthToken_OSPath(t *testing.T) {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		t.Skip("auth path test only on darwin/linux")
	}
	restore := swapPlatformHooks(t)
	defer restore()

	tmp := t.TempDir()
	authPath := vercelAuthFilePath()
	if authPath == "" {
		t.Fatal("expected auth path")
	}
	// Redirect home to temp but keep path structure for darwin/linux
	if runtime.GOOS == "darwin" {
		authPath = filepath.Join(tmp, "Library", "Application Support", "com.vercel.cli", "auth.json")
	} else {
		authPath = filepath.Join(tmp, ".local", "share", "com.vercel.cli", "auth.json")
	}
	if err := os.MkdirAll(filepath.Dir(authPath), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(authPath, []byte(`{"token":"from_file"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	platformHooks.userHomeDir = func() (string, error) { return tmp, nil }
	platformHooks.readFile = os.ReadFile

	tok, ok := readVercelAuthToken()
	if !ok || tok != "from_file" {
		t.Fatalf("token=%q ok=%v", tok, ok)
	}
}

func TestParseVercelEvents(t *testing.T) {
	body := []byte(`[
	  {"type":"stdout","text":"a"},
	  {"type":"stderr","payload":{"text":"b"}}
	]`)
	lines, err := parseVercelEvents(body)
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 2 || lines[0] != "a" || lines[1] != "b" {
		t.Fatalf("%v", lines)
	}
}

func TestParseVercelEvents_NDJSON(t *testing.T) {
	body := []byte("{\"type\":\"stdout\",\"text\":\"line-1\"}\n{\"type\":\"stderr\",\"text\":\"line-2\"}\n")
	lines, err := parseVercelEvents(body)
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 2 || lines[0] != "line-1" || lines[1] != "line-2" {
		t.Fatalf("%v", lines)
	}
}

func TestVercelAuthFilePath_MacSpaces(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("darwin only")
	}
	tmp := t.TempDir()
	platformHooks.userHomeDir = func() (string, error) { return tmp, nil }
	got := vercelAuthFilePath()
	want := filepath.Join(tmp, "Library", "Application Support", "com.vercel.cli", "auth.json")
	if got != want {
		t.Fatalf("path=%q want=%q", got, want)
	}
	if err := os.MkdirAll(filepath.Dir(got), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(got, []byte(`{"token":"spaced_path_ok"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	platformHooks.readFile = os.ReadFile
	tok, ok := readVercelAuthToken()
	if !ok || tok != "spaced_path_ok" {
		t.Fatalf("token=%q ok=%v", tok, ok)
	}
}

func vercelRegistry(t *testing.T, baseURL string) *provider.Registry {
	t.Helper()
	root := testutil.RepoRoot(t)
	raw, err := os.ReadFile(filepath.Join(root, "providers", "hosting", "vercel.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	spec, err := provider.ParseProviderYAML(raw)
	if err != nil {
		t.Fatal(err)
	}
	spec.API.BaseURL = baseURL
	return &provider.Registry{ByName: map[string]*provider.Spec{"vercel": spec}}
}

func swapPlatformHooks(t *testing.T) func() {
	t.Helper()
	prev := platformHooks
	prevSetup := setupHooks
	tmp := t.TempDir()
	credPath := filepath.Join(tmp, ".perch", "credentials")
	platformHooks = platform{
		getenv:      os.Getenv,
		readFile:    os.ReadFile,
		userHomeDir: func() (string, error) { return tmp, nil },
		httpClient:  provider.HTTPClientForAPI(),
		credentialsStore: func() *credentials.Store {
			return credentials.NewStoreAt(credPath)
		},
	}
	setupHooks.lookPath = func(string) (string, error) { return "", errNotFound }
	setupHooks.runShell = func(context.Context, string, time.Duration) error {
		return errPath("auto setup disabled in test")
	}
	return func() {
		platformHooks = prev
		setupHooks = prevSetup
	}
}
