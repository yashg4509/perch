package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/yashg4509/perch/internal/config"
	"github.com/yashg4509/perch/internal/credentials"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage perch.yaml project fields (non-secret)",
	}
	cmd.AddCommand(newConfigSyncEnvCmd())
	return cmd
}

func newConfigSyncEnvCmd() *cobra.Command {
	var envFile string
	var envMapFile string
	var perchPath string

	cmd := &cobra.Command{
		Use:   "sync-env",
		Short: "Backfill perch.yaml project/service fields from .env (optional)",
		Long: `Best-effort update of perch.yaml (not secrets): infers project/service from .env
using provider project_env_aliases, then optional perch.envmap.yaml overrides, then dev
defaults. Status and graph already read .env at runtime — you do not need this for
local dev. Does not modify ~/.perch/credentials — use perch auth sync-env for API keys.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = args
			env, err := cmd.Flags().GetString("env")
			if err != nil {
				return err
			}
			wd, err := os.Getwd()
			if err != nil {
				return err
			}
			if perchPath == "" {
				perchPath, err = config.FindPerchYAML(wd)
				if err != nil {
					return err
				}
			}
			root := filepath.Dir(perchPath)
			if envMapFile == "" {
				envMapFile = filepath.Join(root, "perch.envmap.yaml")
			}
			raw, err := os.ReadFile(perchPath)
			if err != nil {
				return err
			}
			cfg, err := config.Load(raw)
			if err != nil {
				return err
			}
			dotenvRaw, err := os.ReadFile(envFile)
			if err != nil {
				return err
			}
			dotenv, err := credentials.ParseDotenv(string(dotenvRaw))
			if err != nil {
				return err
			}

			reg, _ := loadProviderRegistry()

			var applied []string
			if reg != nil {
				applied = append(applied, config.ApplyEnvInference(cfg, env, dotenv, reg)...)
			}
			if em, err := config.LoadEnvMap(envMapFile); err == nil {
				if mappings, ok := em[env]; ok {
					a, err := config.ApplyEnvMap(cfg, env, dotenv, mappings)
					if err != nil {
						return err
					}
					applied = append(applied, a...)
				}
			}
			if env == "dev" {
				applied = append(applied, config.ApplyDevDefaults(cfg, env)...)
			}

			out, err := config.Marshal(cfg)
			if err != nil {
				return err
			}
			if err := os.WriteFile(perchPath, out, 0o644); err != nil {
				return err
			}

			jsonOut, _ := cmd.Root().PersistentFlags().GetBool("json")
			if jsonOut {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(map[string]any{
					"env":     env,
					"applied": applied,
					"path":    perchPath,
				})
			}
			fmt.Fprintf(cmd.OutOrStdout(), "updated %s (%d fields)\n", perchPath, len(applied))
			for _, a := range applied {
				fmt.Fprintf(cmd.OutOrStdout(), "  - %s\n", a)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&envFile, "env-file", ".env", "Path to project .env")
	cmd.Flags().StringVar(&envMapFile, "envmap", "", "Path to perch.envmap.yaml (default beside perch.yaml)")
	cmd.Flags().StringVar(&perchPath, "config", "", "Path to perch.yaml")
	return cmd
}
