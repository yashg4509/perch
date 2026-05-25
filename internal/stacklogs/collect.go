// Package stacklogs resolves log lines for deployable provider nodes using automatic
// credential discovery (auth files, GitHub Actions, environment tokens) before falling
// back to setup hints.
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
	"github.com/yashg4509/perch/internal/provider"
)

// MaxLines is the maximum number of log lines returned per resolution.
const MaxLines = 4000

// LogResult is the outcome of [Resolve].
type LogResult struct {
	Lines     []string
	Source    string // "auth_file" | "github_actions" | "env_token" | "custom" | "none"
	Provider  string
	Truncated bool
	SetupHint string // only set when Source == "none"
}

// platform hooks for tests (filesystem, env, subprocess, repo detection).
var platformHooks = platform{
	getenv:      os.Getenv,
	readFile:    os.ReadFile,
	userHomeDir: os.UserHomeDir,
	httpClient:  provider.HTTPClientForAPI(),
	ghLookPath:  execLookPath,
	repoSlug:    detectRepoSlug,
}

type platform struct {
	getenv      func(string) string
	readFile    func(string) ([]byte, error)
	userHomeDir func() (string, error)
	httpClient  *http.Client
	ghLookPath  func(string) (string, error)
	repoSlug    func() (owner, repo string, ok bool)
}

// Resolve tries auth-file, GitHub Actions, and environment-token strategies in order.
func Resolve(ctx context.Context, nodeName string, n config.Node, reg *provider.Registry) (LogResult, error) {
	if reg == nil {
		return LogResult{}, fmt.Errorf("stacklogs: nil registry")
	}
	prov := strings.TrimSpace(n.Provider)
	out := LogResult{Provider: prov}

	switch prov {
	case "vercel":
		return resolveVercel(ctx, nodeName, n, reg)
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

func vercelSetupHint() string {
	return "Set VERCEL_TOKEN or run 'vercel login' to enable logs for this node"
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

func ghHostsPath() string {
	if xdg := strings.TrimSpace(platformHooks.getenv("XDG_CONFIG_HOME")); xdg != "" {
		return filepath.Join(xdg, "gh", "hosts.yml")
	}
	return homePath(".config", "gh", "hosts.yml")
}
