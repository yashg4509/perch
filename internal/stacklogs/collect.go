// Package stacklogs resolves log lines for deployable provider nodes using automatic
// credential discovery (auth files, environment tokens, credential store, CLI auto-setup)
// before falling back to setup hints.
package stacklogs

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/yashg4509/perch/internal/config"
	"github.com/yashg4509/perch/internal/credentials"
	"github.com/yashg4509/perch/internal/provider"
)

// MaxLines is the maximum number of log lines returned per resolution.
const MaxLines = 4000

// StrategyResult records one attempt in the log resolution chain.
type StrategyResult struct {
	Name   string `json:"name"`   // "auth_file", "env_token", "credentials_store", "auto_setup"
	Result string `json:"result"` // "success", "token_expired", "not_found", "install_failed", "auth_failed", "not_set", "skipped"
}

// LogResult is the outcome of [Resolve].
type LogResult struct {
	Lines           []string
	Source          string // "auth_file" | "env_token" | "credentials_store" | "custom" | "none"
	Provider        string
	Truncated       bool
	SetupHint       string // only set when Source == "none"
	StrategiesTried []StrategyResult
}

// platform hooks for tests (filesystem, env, subprocess, credential store).
var platformHooks = platform{
	getenv:           os.Getenv,
	readFile:         os.ReadFile,
	userHomeDir:      os.UserHomeDir,
	httpClient:       provider.HTTPClientForAPI(),
	credentialsStore: func() *credentials.Store { return credentials.NewStore() },
}

type platform struct {
	getenv           func(string) string
	readFile         func(string) ([]byte, error)
	userHomeDir      func() (string, error)
	httpClient       *http.Client
	credentialsStore func() *credentials.Store
}

// Resolve tries auth-file, env-token, credential-store, then auto-setup (last resort) in order.
func Resolve(ctx context.Context, nodeName string, n config.Node, reg *provider.Registry) (LogResult, error) {
	return resolveWithFlags(ctx, nodeName, n, reg, false)
}

func resolveWithFlags(ctx context.Context, nodeName string, n config.Node, reg *provider.Registry, autoSetupAttempted bool) (LogResult, error) {
	if reg == nil {
		return LogResult{}, fmt.Errorf("stacklogs: nil registry")
	}
	prov := strings.TrimSpace(n.Provider)
	out := LogResult{Provider: prov}

	switch prov {
	case "vercel":
		return resolveVercel(ctx, nodeName, n, reg, autoSetupAttempted)
	// TODO: add supabase, render
	default:
		out.Source = "none"
		out.SetupHint = fmt.Sprintf("Logs for provider %q are not yet supported", prov)
		return out, nil
	}
}

func nodeFields(n config.Node) map[string]string {
	m := make(map[string]string)
	if n.Project != "" {
		m["project"] = n.Project
	}
	if n.Service != "" {
		m["service"] = n.Service
	}
	return m
}

func tokenFromEnv(spec *provider.Spec) string {
	if spec == nil {
		return ""
	}
	ev := strings.TrimSpace(spec.Credentials.EnvVar)
	if ev == "" {
		return ""
	}
	return strings.TrimSpace(platformHooks.getenv(ev))
}

func tokenFromCredentialsStore(spec *provider.Spec) (string, bool) {
	if spec == nil || strings.TrimSpace(spec.Credentials.Key) == "" {
		return "", false
	}
	store := platformHooks.credentialsStore()
	v, ok, err := store.Get(spec.Credentials.Key)
	if err != nil || !ok {
		return "", false
	}
	v = strings.TrimSpace(v)
	return v, v != ""
}

func truncateLines(lines []string) ([]string, bool) {
	if len(lines) <= MaxLines {
		return lines, false
	}
	return lines[len(lines)-MaxLines:], true
}

func homePath(parts ...string) string {
	dir, err := platformHooks.userHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(append([]string{dir}, parts...)...)
}

func vercelAuthFilePath() string {
	switch runtime.GOOS {
	case "darwin":
		return homePath("Library", "Application Support", "com.vercel.cli", "auth.json")
	case "linux":
		return homePath(".local", "share", "com.vercel.cli", "auth.json")
	default:
		return ""
	}
}
