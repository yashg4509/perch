package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/yashg4509/perch/internal/config"
	"github.com/yashg4509/perch/internal/graph"
	"github.com/yashg4509/perch/internal/provider"
	"github.com/yashg4509/perch/internal/stackstatus"
	"github.com/yashg4509/perch/internal/tui"
)

func runRootTUI(cmd *cobra.Command) error {
	env, err := cmd.Flags().GetString("env")
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

	var g *graph.Graph
	var loadErr error
	var cfg *config.Config
	var reg *provider.Registry

	perchPath, ferr := config.FindPerchYAML(wd)
	if ferr != nil {
		loadErr = fmt.Errorf("find perch.yaml: %w", ferr)
	} else {
		raw, rerr := os.ReadFile(perchPath)
		if rerr != nil {
			loadErr = fmt.Errorf("read config: %w", rerr)
		} else {
			var lerr error
			cfg, lerr = config.Load(raw)
			if lerr != nil {
				loadErr = lerr
			} else {
				r, rr := loadRegistryForProject(filepath.Dir(perchPath))
				if rr != nil {
					loadErr = rr
				} else {
					reg = r
					g, loadErr = graph.Build(cfg, reg, env)
				}
			}
		}
	}

	var envSw *tui.EnvSwitcher
	if cfg != nil && reg != nil && len(cfg.Environments) > 0 {
		names := make([]string, 0, len(cfg.Environments))
		for k := range cfg.Environments {
			names = append(names, k)
		}
		sort.Strings(names)
		idx := 0
		for i, n := range names {
			if n == env {
				idx = i
				break
			}
		}
		envSw = &tui.EnvSwitcher{
			Names: names,
			Index: idx,
			Build: func(e string) (*graph.Graph, error) {
				return graph.Build(cfg, reg, e)
			},
		}
	}

	var fetch tui.StatusFetcher
	if cfg != nil && reg != nil {
		cfgRef := cfg
		regRef := reg
		fetch = func(ctx context.Context, env string) (*stackstatus.EnvReport, error) {
			return stackstatus.Collect(ctx, cfgRef, env, regRef)
		}
	}
	m := tui.NewStackModelWithEnvsAndFetch(g, loadErr, noColor, envSw, fetch)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err = p.Run()
	if err != nil {
		return fmt.Errorf("tui: %w", err)
	}
	return nil
}
