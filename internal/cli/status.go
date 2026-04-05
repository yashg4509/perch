package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/yashg4509/perch/internal/config"
	"github.com/yashg4509/perch/internal/stackstatus"
)

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show health for all nodes in the selected environment",
		RunE:  runStatus,
	}
}

func runStatus(cmd *cobra.Command, args []string) error {
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

	ctx := context.Background()
	rep, err := stackstatus.Collect(ctx, cfg, env, reg)
	if err != nil {
		return err
	}

	out := cmd.OutOrStdout()
	if jsonOut {
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		return enc.Encode(rep)
	}

	_ = noColor
	for _, n := range rep.Nodes {
		st := "unhealthy"
		if n.Healthy {
			st = "healthy"
		}
		_, _ = fmt.Fprintf(out, "%s (%s): %s\n", n.Name, n.Provider, st)
	}
	return nil
}
