package stacklogs

import (
	"context"
	"os/exec"
	"strings"
	"time"

	"github.com/yashg4509/perch/internal/provider"
)

const autoSetupAuthTimeout = 120 * time.Second

var pkgManagerOrder = []string{"npm", "brew", "apt-get"}

// autoSetup installs the provider CLI when needed, runs auth_cmd, and persists the token.
// It is provider-agnostic and driven entirely by spec.CLI and spec.Credentials.
func autoSetup(ctx context.Context, spec *provider.Spec) (bool, string) {
	if spec == nil || spec.CLI == nil || len(spec.CLI.Install) == 0 {
		return false, "not_found"
	}
	binary := strings.TrimSpace(spec.CLI.Binary)
	if binary == "" {
		return false, "not_found"
	}

	if _, err := setupHooks.lookPath(binary); err != nil {
		cmd, pm := installCommand(spec)
		if cmd == "" {
			return false, "install_failed"
		}
		if err := setupHooks.runShell(ctx, cmd, 5*time.Minute); err != nil {
			return false, "install_failed"
		}
		_ = pm // detected package manager name (for future logging)
	}

	authCmd := strings.TrimSpace(spec.CLI.AuthCmd)
	if authCmd == "" {
		return false, "auth_failed"
	}
	if err := setupHooks.runShell(ctx, authCmd, autoSetupAuthTimeout); err != nil {
		return false, "auth_failed"
	}

	tok, ok := readAuthFileToken(spec)
	if !ok || tok == "" {
		return false, "auth_failed"
	}
	store := platformHooks.credentialsStore()
	if err := store.Set(spec.Credentials.Key, tok); err != nil {
		return false, "auth_failed"
	}
	return true, "success"
}

func installCommand(spec *provider.Spec) (cmd, pkgManager string) {
	for _, pm := range pkgManagerOrder {
		if _, err := setupHooks.lookPath(pm); err != nil {
			continue
		}
		c := strings.TrimSpace(spec.CLI.Install[pm])
		if c == "" {
			continue
		}
		return c, pm
	}
	return "", ""
}

func readAuthFileToken(spec *provider.Spec) (string, bool) {
	if spec == nil {
		return "", false
	}
	switch spec.Name {
	case "vercel":
		return readVercelAuthToken()
	default:
		return "", false
	}
}

var setupHooks = struct {
	lookPath func(string) (string, error)
	runShell func(context.Context, string, time.Duration) error
}{
	lookPath: exec.LookPath,
	runShell: runShellCommand,
}

func runShellCommand(ctx context.Context, cmdline string, timeout time.Duration) error {
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}
	// #nosec G204 -- cmdline comes from trusted provider YAML, not end-user input.
	c := exec.CommandContext(ctx, "sh", "-c", cmdline)
	return c.Run()
}
