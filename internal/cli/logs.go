package cli

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yashg4509/perch/internal/credentials"
	"github.com/yashg4509/perch/internal/graph"
	"github.com/yashg4509/perch/internal/provider"
	"github.com/yashg4509/perch/internal/stacklogs"
)

var logsPkgManagerOrder = []string{"npm", "brew", "apt-get"}

func newLogsCmd() *cobra.Command {
	logs := &cobra.Command{
		Use:   "logs",
		Short: "Configure provider credentials for stack log viewing",
	}
	logs.AddCommand(newLogsSetupCmd())
	return logs
}

func newLogsSetupCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "setup",
		Short: "Install provider CLIs and authenticate for log access",
		Long:  "Runs interactive provider login in the foreground and saves tokens to ~/.perch/credentials before starting perch viz.",
		RunE:  runLogsSetup,
	}
}

func runLogsSetup(cmd *cobra.Command, args []string) error {
	_ = args
	env, err := cmd.Flags().GetString("env")
	if err != nil {
		return err
	}

	cfg, reg, _, err := loadStackFromWD()
	if err != nil {
		return err
	}

	g, err := graph.Build(cfg, reg, env)
	if err != nil {
		return err
	}

	out := cmd.OutOrStdout()
	store := credentials.NewStore()

	for _, node := range g.Nodes {
		if !node.Deployable {
			continue
		}
		if node.Provider == "custom" {
			continue
		}

		spec := reg.ByName[node.Provider]
		if spec == nil {
			_, _ = fmt.Fprintf(out, "Setting up logs for %s (%s)...\n", node.Name, node.Provider)
			_, _ = fmt.Fprintf(out, "✗ unknown provider %q\n\n", node.Provider)
			continue
		}

		if err := setupLogsForNode(out, node.Name, spec, store); err != nil {
			_, _ = fmt.Fprintf(out, "✗ %s\n\n", err)
		}
	}

	_, _ = fmt.Fprintln(out, "Setup complete. Run perch viz to see logs in the UI.")
	return nil
}

func setupLogsForNode(out io.Writer, nodeName string, spec *provider.Spec, store *credentials.Store) error {
	prov := spec.Name
	_, _ = fmt.Fprintf(out, "Setting up logs for %s (%s)...\n", nodeName, prov)

	if logsAlreadyConfigured(spec, store) {
		_, _ = fmt.Fprintln(out, "✓ already configured")
		_, _ = fmt.Fprintln(out)
		return nil
	}

	if spec.CLI == nil || len(spec.CLI.Install) == 0 {
		_, _ = fmt.Fprintf(out, "⚠ %s log setup not yet supported\n\n", prov)
		return nil
	}

	binary := strings.TrimSpace(spec.CLI.Binary)
	if binary == "" {
		_, _ = fmt.Fprintf(out, "⚠ %s log setup not yet supported\n\n", prov)
		return nil
	}

	if _, err := exec.LookPath(binary); err != nil {
		installCmd, pm := logsInstallCommand(spec)
		if installCmd == "" {
			_, _ = fmt.Fprintf(out, "✗ could not install %s CLI (no supported package manager)\n\n", prov)
			return nil
		}
		_, _ = fmt.Fprintf(out, "→ installing %s CLI via %s...\n", prov, pm)
		// #nosec G204 -- installCmd comes from trusted provider YAML.
		if err := exec.Command("sh", "-c", installCmd).Run(); err != nil {
			_, _ = fmt.Fprintf(out, "✗ install failed: %v\n\n", err)
			return nil
		}
	} else {
		_, _ = fmt.Fprintf(out, "→ %s CLI already installed\n", prov)
	}

	authCmd := strings.TrimSpace(spec.CLI.AuthCmd)
	if authCmd == "" {
		_, _ = fmt.Fprintf(out, "⚠ %s log setup not yet supported\n\n", prov)
		return nil
	}

	_, _ = fmt.Fprintf(out, "→ opening browser for %s login...\n", prov)
	// #nosec G204 -- authCmd comes from trusted provider YAML.
	auth := exec.Command("sh", "-c", authCmd)
	auth.Stdin = os.Stdin
	auth.Stdout = os.Stdout
	auth.Stderr = os.Stderr
	if err := auth.Run(); err != nil {
		_, _ = fmt.Fprintf(out, "✗ auth failed: %v\n\n", err)
		return nil
	}

	stacklogs.PersistAutoSetupToken(spec)
	if _, ok, _ := store.Get(spec.Credentials.Key); ok {
		_, _ = fmt.Fprintln(out, "✓ token saved to ~/.perch/credentials")
	}
	_, _ = fmt.Fprintf(out, "✓ logs ready for %s\n\n", nodeName)
	return nil
}

func logsAlreadyConfigured(spec *provider.Spec, store *credentials.Store) bool {
	if tok, ok := stacklogs.AuthFileToken(spec); ok && strings.TrimSpace(tok) != "" {
		return true
	}
	key := strings.TrimSpace(spec.Credentials.Key)
	if key == "" {
		return false
	}
	v, ok, err := store.Get(key)
	return err == nil && ok && strings.TrimSpace(v) != ""
}

func logsInstallCommand(spec *provider.Spec) (cmd, pkgManager string) {
	if spec == nil || spec.CLI == nil {
		return "", ""
	}
	for _, pm := range logsPkgManagerOrder {
		if _, err := exec.LookPath(pm); err != nil {
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
