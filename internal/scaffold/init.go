package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/yashg4509/perch/internal/config"
	"github.com/yashg4509/perch/internal/detect"
	"github.com/yashg4509/perch/internal/provider"
)

// Options controls generated perch.yaml content.
type Options struct {
	AppName  string
	Env      string
	Registry *provider.Registry
}

// Generate scans root for init signals and returns perch.yaml bytes plus edge inference metadata.
func Generate(root string, opt Options) ([]byte, detect.Inference, error) {
	if opt.AppName == "" {
		return nil, detect.Inference{}, fmt.Errorf("scaffold: app name is required")
	}
	if opt.Env == "" {
		return nil, detect.Inference{}, fmt.Errorf("scaffold: environment name is required")
	}
	if opt.Registry == nil {
		return nil, detect.Inference{}, fmt.Errorf("scaffold: registry is required")
	}

	sigs, err := collectSignals(root)
	if err != nil {
		return nil, detect.Inference{}, err
	}
	uniq := detect.UniqueProviders(sigs)
	nodes := AssignNodeNames(uniq)

	var depKeys []string
	pj := filepath.Join(root, "package.json")
	if _, err := os.Stat(pj); err == nil {
		depKeys, err = detect.NPMDependencyKeys(pj)
		if err != nil {
			return nil, detect.Inference{}, err
		}
	}
	inf := detect.InferEdges(nodes, depKeys)

	raw, err := renderPerchYAML(opt.AppName, opt.Env, nodes, inf.Edges, opt.Registry)
	if err != nil {
		return nil, inf, err
	}
	if _, err := config.Load(raw); err != nil {
		return nil, inf, fmt.Errorf("scaffold: generated invalid perch.yaml: %w", err)
	}
	return raw, inf, nil
}

// WriteIfChanged writes perch.yaml when missing or when content differs (idempotent).
func WriteIfChanged(root string, opt Options) (written bool, inf detect.Inference, err error) {
	raw, inf, err := Generate(root, opt)
	if err != nil {
		return false, inf, err
	}
	out := filepath.Join(root, "perch.yaml")
	if prev, err := os.ReadFile(out); err == nil && string(prev) == string(raw) {
		return false, inf, nil
	} else if err != nil && !os.IsNotExist(err) {
		return false, inf, err
	}
	// #nosec G306 — perch.yaml is non-secret stack metadata meant to be committed (see spec).
	if err := os.WriteFile(out, raw, 0o644); err != nil {
		return false, inf, err
	}
	return true, inf, nil
}

func collectSignals(root string) ([]detect.Signal, error) {
	a, err := detect.ConfigFileSignals(root)
	if err != nil {
		return nil, err
	}
	pj := filepath.Join(root, "package.json")
	if _, err := os.Stat(pj); err != nil {
		return a, nil
	}
	b, err := detect.PackageJSONSignals(pj)
	if err != nil {
		return nil, err
	}
	return append(a, b...), nil
}

func renderPerchYAML(app, env string, nodes map[string]string, edges []detect.EdgePair, reg *provider.Registry) ([]byte, error) {
	names := sortedNames(nodes)
	var b strings.Builder
	fmt.Fprintf(&b, "name: %s\n", app)
	fmt.Fprintf(&b, "environments:\n")
	fmt.Fprintf(&b, "  %s:\n", env)
	for _, name := range names {
		pid := nodes[name]
		for _, line := range nodeYAML(name, pid, reg) {
			fmt.Fprintf(&b, "%s\n", line)
		}
	}

	es := append([]detect.EdgePair(nil), edges...)
	sort.Slice(es, func(i, j int) bool {
		if es[i].From != es[j].From {
			return es[i].From < es[j].From
		}
		return es[i].To < es[j].To
	})
	if len(es) == 0 {
		fmt.Fprintf(&b, "edges: []\n")
	} else {
		fmt.Fprintf(&b, "edges:\n")
		for _, e := range es {
			fmt.Fprintf(&b, "  - %s -> %s\n", e.From, e.To)
		}
	}
	return []byte(b.String()), nil
}

func sortedNames(nodes map[string]string) []string {
	ks := make([]string, 0, len(nodes))
	for k := range nodes {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	return ks
}

func nodeYAML(name, pid string, reg *provider.Registry) []string {
	var lines []string
	lines = append(lines, fmt.Sprintf("    %s:", name))
	lines = append(lines, fmt.Sprintf("      provider: %s", pid))
	if pid == "custom" {
		lines = append(lines, `      status: "exit 0"`)
		return lines
	}
	sp, ok := reg.ByName[pid]
	if !ok {
		return lines
	}
	if !sp.Deployable {
		return lines
	}
	switch pid {
	case "vercel", "netlify", "cloudflare", "firebase":
		lines = append(lines, "      project: CHANGE_ME")
	case "render", "railway", "fly":
		lines = append(lines, "      service: CHANGE_ME")
	default:
		lines = append(lines, "      project: CHANGE_ME")
	}
	return lines
}
