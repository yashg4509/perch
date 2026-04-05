package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/yashg4509/perch/internal/config"
	"github.com/yashg4509/perch/internal/graph"
	"github.com/yashg4509/perch/internal/stackcontext"
	"github.com/yashg4509/perch/internal/stackstatus"
)

func newContextCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "context",
		Short: "Print merged stack topology + status for agents or JSON consumers",
		RunE:  runContext,
	}
	cmd.Flags().Bool("for-agent", false, "Emit plain text optimized for LLM context injection")
	return cmd
}

func runContext(cmd *cobra.Command, args []string) error {
	_ = args
	env, err := cmd.Flags().GetString("env")
	if err != nil {
		return err
	}
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return err
	}
	forAgent, err := cmd.Flags().GetBool("for-agent")
	if err != nil {
		return err
	}
	noColor, err := cmd.Flags().GetBool("no-color")
	if err != nil {
		return err
	}
	if forAgent && jsonOut {
		return fmt.Errorf("context: use only one of --json or --for-agent")
	}

	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	perchPath, err := config.FindPerchYAML(wd)
	if err != nil {
		return err
	}
	raw, err := os.ReadFile(perchPath)
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}
	cfg, err := config.Load(raw)
	if err != nil {
		return err
	}
	root := filepath.Dir(perchPath)
	reg, err := loadRegistryForProject(root)
	if err != nil {
		return err
	}

	ctx := context.Background()
	g, err := graph.Build(cfg, reg, env)
	if err != nil {
		return err
	}
	rep, err := stackstatus.Collect(ctx, cfg, env, reg)
	if err != nil {
		return err
	}

	at := time.Now()
	r := stackcontext.Build(at, g, rep)

	out := cmd.OutOrStdout()
	if forAgent {
		_ = noColor
		return writeContextForAgent(out, r)
	}
	_ = noColor
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(r)
}

func writeContextForAgent(w io.Writer, r *stackcontext.Report) error {
	var b strings.Builder
	_, _ = b.WriteString("Stack: ")
	b.WriteString(r.Stack)
	b.WriteString("\nEnvironment: ")
	b.WriteString(r.Environment)
	b.WriteString("\nGenerated: ")
	b.WriteString(r.GeneratedAt)
	b.WriteString("\n\nNodes:\n")
	for _, n := range r.Nodes {
		st := "healthy"
		if !n.Healthy {
			st = "unhealthy"
		}
		kind := "read-only"
		if n.Deployable {
			kind = "deployable"
		}
		_, _ = fmt.Fprintf(&b, "- %s (%s, %s): %s\n", n.Name, n.Provider, kind, st)
	}
	if r.Summary != "" {
		_, _ = fmt.Fprintf(&b, "\nSummary: %s\n", r.Summary)
	}
	_, err := w.Write([]byte(b.String()))
	return err
}
