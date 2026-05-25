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
	platformHooks.repoSlug = func() (string, string, bool) { return "", "", false }

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
	platformHooks.repoSlug = func() (string, string, bool) { return "", "", false }

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
	platformHooks.repoSlug = func() (string, string, bool) { return "", "", false }

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

func TestResolve_Vercel_GitHubActions(t *testing.T) {
	restore := swapPlatformHooks(t)
	defer restore()

	ghSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/acme/widget/actions/runs":
			_, _ = io.WriteString(w, `{"workflow_runs":[{"id":99}]}`)
		case "/repos/acme/widget/actions/runs/99/jobs":
			_, _ = io.WriteString(w, `{"jobs":[{"id":7,"name":"Deploy to Vercel"}]}`)
		case "/repos/acme/widget/actions/jobs/7/logs":
			_, _ = io.WriteString(w, "step 1\nDeploying with vercel\nstep 2\n")
		default:
			t.Fatalf("github path %s", r.URL.Path)
		}
	}))
	t.Cleanup(ghSrv.Close)

	oldClient := githubHTTPClient
	oldBase := githubAPIBase
	githubHTTPClient = ghSrv.Client()
	githubAPIBase = ghSrv.URL
	t.Cleanup(func() {
		githubHTTPClient = oldClient
		githubAPIBase = oldBase
	})

	platformHooks.readFile = func(path string) ([]byte, error) {
		if strings.HasSuffix(path, "hosts.yml") {
			return []byte("github.com:\n  oauth_token: gh_test\n"), nil
		}
		return nil, os.ErrNotExist
	}
	platformHooks.userHomeDir = func() (string, error) { return t.TempDir(), nil }
	platformHooks.getenv = func(string) string { return "" }
	platformHooks.repoSlug = func() (string, string, bool) { return "acme", "widget", true }
	platformHooks.ghLookPath = func(string) (string, error) { return "", os.ErrNotExist }

	reg := vercelRegistry(t, "http://127.0.0.1:1") // unreachable; should not be called
	got, err := Resolve(context.Background(), "web", config.Node{Provider: "vercel", Project: "demo"}, reg)
	if err != nil {
		t.Fatal(err)
	}
	if got.Source != "github_actions" {
		t.Fatalf("source=%q", got.Source)
	}
	if len(got.Lines) == 0 {
		t.Fatal("expected filtered github log lines")
	}
	found := false
	for _, l := range got.Lines {
		if strings.Contains(strings.ToLower(l), "vercel") {
			found = true
		}
	}
	if !found {
		t.Fatalf("lines=%v", got.Lines)
	}
}

func TestResolve_Vercel_None(t *testing.T) {
	restore := swapPlatformHooks(t)
	defer restore()

	platformHooks.readFile = func(string) ([]byte, error) { return nil, os.ErrNotExist }
	platformHooks.userHomeDir = func() (string, error) { return t.TempDir(), nil }
	platformHooks.getenv = func(string) string { return "" }
	platformHooks.repoSlug = func() (string, string, bool) { return "", "", false }

	reg := vercelRegistry(t, "http://127.0.0.1:1")
	got, err := Resolve(context.Background(), "web", config.Node{Provider: "vercel", Project: "demo"}, reg)
	if err != nil {
		t.Fatal(err)
	}
	if got.Source != "none" {
		t.Fatalf("source=%q", got.Source)
	}
	if got.SetupHint != vercelSetupHint() {
		t.Fatalf("hint=%q", got.SetupHint)
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

func TestReadGitHubTokenFromHosts(t *testing.T) {
	restore := swapPlatformHooks(t)
	defer restore()

	tmp := t.TempDir()
	hosts := filepath.Join(tmp, "gh", "hosts.yml")
	if err := os.MkdirAll(filepath.Dir(hosts), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(hosts, []byte("github.com:\n  oauth_token: tok_from_hosts\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	platformHooks.userHomeDir = func() (string, error) { return tmp, nil }
	platformHooks.getenv = func(k string) string {
		if k == "XDG_CONFIG_HOME" {
			return filepath.Join(tmp, "xdg")
		}
		return ""
	}
	// Override gh path via XDG
	xdgHosts := filepath.Join(tmp, "xdg", "gh", "hosts.yml")
	if err := os.MkdirAll(filepath.Dir(xdgHosts), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(xdgHosts, []byte("github.com:\n  oauth_token: xdg_tok\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	platformHooks.readFile = os.ReadFile

	tok, ok := readGitHubTokenFromHosts()
	if !ok || tok != "xdg_tok" {
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

func TestDetectRepoSlug_WalksUpFromSubdir(t *testing.T) {
	root := t.TempDir()
	repo := filepath.Join(root, "my-app")
	sub := filepath.Join(repo, "packages", "web")
	for _, d := range []string{repo, sub} {
		if err := os.MkdirAll(d, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	gitDir := filepath.Join(repo, ".git")
	if err := os.Mkdir(gitDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(gitDir, "config"), []byte(`
[remote "origin"]
	url = https://github.com/acme/widget.git
`), 0o600); err != nil {
		t.Fatal(err)
	}

	prev, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(sub); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(prev) })

	owner, name, ok := detectRepoSlug()
	if !ok {
		t.Fatal("expected repo detection from subdirectory")
	}
	if owner != "acme" || name != "widget" {
		t.Fatalf("got %s/%s", owner, name)
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
	platformHooks = platform{
		getenv:      os.Getenv,
		readFile:    os.ReadFile,
		userHomeDir: os.UserHomeDir,
		httpClient:  provider.HTTPClientForAPI(),
		ghLookPath:  execLookPath,
		repoSlug:    detectRepoSlug,
	}
	return func() { platformHooks = prev }
}
