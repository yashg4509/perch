package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yashg4509/perch/internal/scaffold"
)

func newInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Generate perch.yaml from detected project signals",
		RunE:  runInit,
	}
	cmd.Flags().String("name", "", "Application name (default: current directory base name)")
	return cmd
}

func runInit(cmd *cobra.Command, args []string) error {
	_ = args
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}
	env, err := cmd.Flags().GetString("env")
	if err != nil {
		return err
	}
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return err
	}

	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	if strings.TrimSpace(name) == "" {
		name = filepath.Base(wd)
		if name == "." || name == "/" || name == "" {
			return fmt.Errorf("init: cannot infer app name from directory; pass --name")
		}
	}

	reg, err := loadRegistryForProject(wd)
	if err != nil {
		return fmt.Errorf("init: load providers: %w", err)
	}

	opt := scaffold.Options{AppName: strings.TrimSpace(name), Env: env, Registry: reg}
	written, inf, err := scaffold.WriteIfChanged(wd, opt)
	if err != nil {
		return err
	}

	out := cmd.OutOrStdout()
	errOut := cmd.ErrOrStderr()
	for _, msg := range inf.NeedsPrompt {
		_, _ = fmt.Fprintf(errOut, "perch init: note: %s\n", msg)
	}

	if jsonOut {
		res := struct {
			Written     bool     `json:"written"`
			Unchanged   bool     `json:"unchanged"`
			NeedsPrompt []string `json:"needs_prompt,omitempty"`
		}{
			Written:     written,
			Unchanged:   !written,
			NeedsPrompt: inf.NeedsPrompt,
		}
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		return enc.Encode(res)
	}

	if written {
		_, _ = fmt.Fprintln(out, "Wrote perch.yaml")
	} else {
		_, _ = fmt.Fprintln(out, "perch.yaml unchanged")
	}
	return nil
}
