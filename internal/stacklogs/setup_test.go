package stacklogs

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yashg4509/perch/internal/credentials"
	"github.com/yashg4509/perch/internal/provider"
)

func TestAutoSetup_skipsInstallWhenBinaryOnPath(t *testing.T) {
	restore := swapSetupHooks(t)
	defer restore()
	restorePlatform := swapPlatformHooks(t)
	defer restorePlatform()

	tmp := t.TempDir()
	platformHooks.userHomeDir = func() (string, error) { return tmp, nil }
	authPath := vercelAuthFilePath()
	if authPath == "" {
		t.Skip("no vercel auth path on this OS")
	}
	if err := os.MkdirAll(filepath.Dir(authPath), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(authPath, []byte(`{"token":"after_login"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	platformHooks.readFile = os.ReadFile
	credPath := filepath.Join(tmp, ".perch", "credentials")
	platformHooks.credentialsStore = func() *credentials.Store {
		return credentials.NewStoreAt(credPath)
	}

	var ran []string
	setupHooks.lookPath = func(name string) (string, error) {
		if name == "vercel" {
			return "/usr/bin/vercel", nil
		}
		return "", errNotFound
	}
	setupHooks.runShell = func(_ context.Context, cmd string, _ time.Duration) error {
		ran = append(ran, cmd)
		return nil
	}

	spec := &provider.Spec{
		Name: "vercel",
		CLI: &provider.CLISpec{
			Binary: "vercel",
			Install: map[string]string{
				"npm": "npm install -g vercel",
			},
			AuthCmd: "vercel login",
		},
		Credentials: provider.CredentialsSpec{Key: "vercel_token"},
	}

	ok, result := autoSetup(context.Background(), spec)
	if !ok || result != "success" {
		t.Fatalf("ok=%v result=%q", ok, result)
	}
	for _, cmd := range ran {
		if cmd == "npm install -g vercel" {
			t.Fatalf("install should be skipped when binary on PATH, ran=%v", ran)
		}
	}
	if len(ran) != 1 || ran[0] != "vercel login" {
		t.Fatalf("ran=%v", ran)
	}
	tok, ok, err := credentials.NewStoreAt(credPath).Get("vercel_token")
	if err != nil || !ok || tok != "after_login" {
		t.Fatalf("store token=%q ok=%v err=%v", tok, ok, err)
	}
}

func TestAutoSetup_returnsFalseWhenInstallEmpty(t *testing.T) {
	restore := swapSetupHooks(t)
	defer restore()

	spec := &provider.Spec{
		Name: "vercel",
		CLI: &provider.CLISpec{
			Binary:  "vercel",
			AuthCmd: "vercel login",
		},
		Credentials: provider.CredentialsSpec{Key: "vercel_token"},
	}
	ok, result := autoSetup(context.Background(), spec)
	if ok {
		t.Fatal("expected false")
	}
	if result != "not_found" {
		t.Fatalf("result=%q", result)
	}
}

func swapSetupHooks(t *testing.T) func() {
	t.Helper()
	prev := setupHooks
	return func() { setupHooks = prev }
}

var errNotFound = errPath("not found")

type errPath string

func (e errPath) Error() string { return string(e) }
