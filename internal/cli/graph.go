package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/yashg4509/perch/internal/config"
	"github.com/yashg4509/perch/internal/graph"
)

func newGraphCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "graph",
		Short: "Print stack topology for the selected environment",
		RunE:  runGraph,
	}
}

func runGraph(cmd *cobra.Command, args []string) error {
	_ = args
	env, err := cmd.Flags().GetString("env")
	if err != nil {
		return err
	}
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return err
	}
	noColor, err := cmd.Flags().GetBool("no-color")
	if err != nil {
		return err
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

	g, err := graph.Build(cfg, reg, env)
	if err != nil {
		return err
	}
	rep := graph.NewJSONReport(g)

	out := cmd.OutOrStdout()
	if jsonOut {
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		return enc.Encode(rep)
	}

	_ = noColor
	_, _ = fmt.Fprintf(out, "%s (%s)\n", rep.AppName, rep.Environment)
	for _, n := range rep.Nodes {
		kind := "read-only"
		if n.Deployable {
			kind = "deployable"
		}
		_, _ = fmt.Fprintf(out, "  %s  %s  (%s)\n", n.Name, n.Provider, kind)
	}
	for _, e := range rep.Edges {
		_, _ = fmt.Fprintf(out, "  %s -> %s\n", e.From, e.To)
	}
	return nil
}
