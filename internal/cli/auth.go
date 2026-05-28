package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yashg4509/perch/internal/config"
	"github.com/yashg4509/perch/internal/credentials"
	"github.com/yashg4509/perch/internal/provider"
)

func newAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage Perch credentials (~/.perch/credentials)",
	}
	cmd.AddCommand(newAuthSyncEnvCmd())
	return cmd
}

func newAuthSyncEnvCmd() *cobra.Command {
	var envFile string
	var credPath string
	var overwrite bool
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "sync-env",
		Short: "Import secrets from a project .env into ~/.perch/credentials",
		Long: `Reads a local .env (never committed) and copies mapped secrets into
~/.perch/credentials using provider env_aliases. One-way import: Perch does not
write back to .env. Existing credential keys are skipped unless --overwrite.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = args
			reg, err := loadProviderRegistry()
			if err != nil {
				return err
			}
			env, err := cmd.Flags().GetString("env")
			if err != nil {
				return err
			}
			var specs []provider.CredentialsSpec
			wd, _ := os.Getwd()
			if perchPath, err := config.FindPerchYAML(wd); err == nil {
				if raw, err := os.ReadFile(perchPath); err == nil {
					if cfg, err := config.Load(raw); err == nil {
						specs = credentialSpecsForEnv(cfg, env, reg)
					}
				}
			}
			if len(specs) == 0 {
				specs = credentials.UniqueCredentialSpecs(reg)
			}

			path := credPath
			if path == "" {
				path, err = credentials.DefaultPath()
				if err != nil {
					return err
				}
			}
			store := credentials.NewStore(path)

			res, err := credentials.ImportEnvFile(store, envFile, specs, credentials.ImportOptions{
				Overwrite: overwrite,
				DryRun:    dryRun,
			})
			if err != nil {
				return err
			}

			jsonOut, _ := cmd.Root().PersistentFlags().GetBool("json")
			if jsonOut {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(res)
			}

			if dryRun {
				fmt.Fprintln(cmd.OutOrStdout(), "dry run — no credentials written")
			}
			printKeyList(cmd, "imported", res.Imported)
			printKeyList(cmd, "skipped (already set)", res.Skipped)
			if len(res.Missing) > 0 {
				printKeyList(cmd, "missing in .env", res.Missing)
			}
			if len(res.Imported) == 0 && len(res.Skipped) == 0 {
				fmt.Fprintln(cmd.ErrOrStderr(), "no credentials imported; fill .env or use --overwrite")
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&envFile, "env-file", ".env", "Path to project .env file")
	cmd.Flags().StringVar(&credPath, "credentials", "", "Credentials file (default ~/.perch/credentials)")
	cmd.Flags().BoolVar(&overwrite, "overwrite", false, "Replace existing credential keys")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be imported without writing")
	return cmd
}

func printKeyList(cmd *cobra.Command, label string, keys []string) {
	if len(keys) == 0 {
		return
	}
	fmt.Fprintf(cmd.OutOrStdout(), "%s (%d):\n", label, len(keys))
	for _, k := range keys {
		fmt.Fprintf(cmd.OutOrStdout(), "  - %s\n", k)
	}
}
