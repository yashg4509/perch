package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/yashg4509/perch/internal/config"
)

func newEdgeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edge",
		Short: "List, add, or remove stack edges in perch.yaml",
		Long: `Edit directed edges between node names (same names as under environments.*).

Writes a canonical perch.yaml: key order is sorted and inline comments are not preserved.`,
	}
	cmd.AddCommand(newEdgeListCmd())
	cmd.AddCommand(newEdgeAddCmd())
	cmd.AddCommand(newEdgeRemoveCmd())
	return cmd
}

func newEdgeListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "Print edges defined in perch.yaml",
		RunE:  runEdgeList,
	}
}

func newEdgeAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add FROM TO",
		Short: "Append an edge FROM -> TO if missing",
		Args:  cobra.ExactArgs(2),
		RunE:  runEdgeAdd,
	}
	cmd.Flags().Bool("dry-run", false, "Print the resulting YAML to stdout instead of writing")
	return cmd
}

func newEdgeRemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "remove FROM TO",
		Aliases: []string{"rm", "delete", "del"},
		Short:   "Remove every edge FROM -> TO",
		Args:    cobra.ExactArgs(2),
		RunE:    runEdgeRemove,
	}
	cmd.Flags().Bool("dry-run", false, "Print the resulting YAML to stdout instead of writing")
	return cmd
}

func runEdgeList(cmd *cobra.Command, args []string) error {
	_ = args
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return err
	}
	_, cfg, err := loadPerchFromCWD()
	if err != nil {
		return err
	}
	out := cmd.OutOrStdout()
	if jsonOut {
		type row struct {
			From string `json:"from"`
			To   string `json:"to"`
		}
		edges := make([]row, 0, len(cfg.Edges))
		for _, e := range cfg.Edges {
			edges = append(edges, row{From: e.From, To: e.To})
		}
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		return enc.Encode(map[string]any{"edges": edges})
	}
	if len(cfg.Edges) == 0 {
		_, _ = fmt.Fprintln(out, "(no edges)")
		return nil
	}
	for _, e := range cfg.Edges {
		_, _ = fmt.Fprintf(out, "%s -> %s\n", e.From, e.To)
	}
	return nil
}

func runEdgeAdd(cmd *cobra.Command, args []string) error {
	from, to := args[0], args[1]
	dry, err := cmd.Flags().GetBool("dry-run")
	if err != nil {
		return err
	}
	perchPath, cfg, err := loadPerchFromCWD()
	if err != nil {
		return err
	}
	if err := config.AddEdge(cfg, from, to); err != nil {
		return err
	}
	note := fmt.Sprintf("added edge %s -> %s", from, to)
	return writePerchYAML(cmd, perchPath, cfg, dry, note)
}

func runEdgeRemove(cmd *cobra.Command, args []string) error {
	from, to := args[0], args[1]
	dry, err := cmd.Flags().GetBool("dry-run")
	if err != nil {
		return err
	}
	perchPath, cfg, err := loadPerchFromCWD()
	if err != nil {
		return err
	}
	n := config.RemoveEdge(cfg, from, to)
	if err := config.Validate(cfg); err != nil {
		return err
	}
	if n == 0 && !dry {
		return fmt.Errorf("edge: no edge %q -> %q", from, to)
	}
	note := ""
	if n == 0 {
		note = "dry-run: no matching edge (output is canonical perch.yaml)"
	} else {
		note = fmt.Sprintf("removed %d edge(s) %q -> %q", n, from, to)
	}
	return writePerchYAML(cmd, perchPath, cfg, dry, note)
}

func loadPerchFromCWD() (perchPath string, cfg *config.Config, err error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", nil, err
	}
	perchPath, err = config.FindPerchYAML(wd)
	if err != nil {
		return "", nil, err
	}
	raw, err := os.ReadFile(perchPath)
	if err != nil {
		return "", nil, fmt.Errorf("read config: %w", err)
	}
	cfg, err = config.Load(raw)
	if err != nil {
		return "", nil, err
	}
	return perchPath, cfg, nil
}

func writePerchYAML(cmd *cobra.Command, perchPath string, cfg *config.Config, dry bool, stderrNote string) error {
	data, err := config.FormatYAML(cfg)
	if err != nil {
		return err
	}
	out := cmd.OutOrStdout()
	errOut := cmd.ErrOrStderr()
	if dry {
		_, _ = fmt.Fprint(out, string(data))
		if stderrNote != "" {
			_, _ = fmt.Fprintf(errOut, "perch edge: %s\n", stderrNote)
		}
		return nil
	}
	if err := os.WriteFile(perchPath, data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", perchPath, err)
	}
	if stderrNote != "" {
		_, _ = fmt.Fprintf(errOut, "perch edge: %s\n", stderrNote)
	}
	_, _ = fmt.Fprintf(errOut, "Updated %s\n", filepath.Clean(perchPath))
	return nil
}
