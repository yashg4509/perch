package detect

import "sort"

// EdgePair is a directed edge between logical node names in perch.yaml.
type EdgePair struct {
	From string
	To   string
}

// Inference collects inferred edges and cases that need a user prompt.
type Inference struct {
	Edges       []EdgePair
	NeedsPrompt []string
}

// InferEdges applies deterministic rules from package.json dependencies and detected nodes.
// npmPackages are dependency keys (e.g. "openai", "@supabase/supabase-js").
//
// Heuristics (all optional; empty edges is valid):
//   - @supabase/supabase-js → edge from the primary stack connector (API host if present, else Vercel/Netlify/etc.) to the supabase node
//   - openai → connector → openai node when the openai package is listed
//   - stripe → connector → stripe node when the stripe package is listed
//
// Edges are sorted by (From, To) for stable perch.yaml output.
func InferEdges(nodes map[string]string, npmPackages []string) Inference {
	var inf Inference
	seen := make(map[EdgePair]struct{})
	add := func(from, to string) {
		if from == "" || to == "" {
			return
		}
		e := EdgePair{From: from, To: to}
		if _, ok := seen[e]; ok {
			return
		}
		seen[e] = struct{}{}
		inf.Edges = append(inf.Edges, e)
	}

	if hasNPM(npmPackages, "@supabase/supabase-js") {
		if !hasProvider(nodes, "supabase") {
			inf.NeedsPrompt = append(inf.NeedsPrompt, "package.json lists @supabase/supabase-js but no supabase provider was detected")
		} else {
			conn := stackConnectorNode(nodes)
			sup := stableNodeForProvider(nodes, "supabase")
			if conn == "" {
				inf.NeedsPrompt = append(inf.NeedsPrompt, "@supabase/supabase-js present but no hosting node (vercel/netlify/render/…) was detected to attach the edge from")
			} else {
				add(conn, sup)
			}
		}
	}

	if hasNPM(npmPackages, "openai") && hasProvider(nodes, "openai") {
		conn := stackConnectorNode(nodes)
		llm := stableNodeForProvider(nodes, "openai")
		if conn != "" && llm != "" {
			add(conn, llm)
		}
	}

	if hasNPM(npmPackages, "stripe") && hasProvider(nodes, "stripe") {
		conn := stackConnectorNode(nodes)
		pay := stableNodeForProvider(nodes, "stripe")
		if conn != "" && pay != "" {
			add(conn, pay)
		}
	}

	sort.Slice(inf.Edges, func(i, j int) bool {
		if inf.Edges[i].From != inf.Edges[j].From {
			return inf.Edges[i].From < inf.Edges[j].From
		}
		return inf.Edges[i].To < inf.Edges[j].To
	})
	return inf
}

func hasNPM(pkgs []string, want string) bool {
	for _, p := range pkgs {
		if p == want {
			return true
		}
	}
	return false
}

func hasProvider(nodes map[string]string, provider string) bool {
	for _, p := range nodes {
		if p == provider {
			return true
		}
	}
	return false
}

// stackConnectorNode picks a single “app” node to hang integration edges from.
// Prefers API/worker hosts, then static/edge frontends — deterministic tie-break by node name.
func stackConnectorNode(nodes map[string]string) string {
	for _, prov := range []string{"render", "railway", "fly"} {
		if n := stableNodeForProvider(nodes, prov); n != "" {
			return n
		}
	}
	for _, prov := range []string{"vercel", "netlify", "cloudflare", "firebase"} {
		if n := stableNodeForProvider(nodes, prov); n != "" {
			return n
		}
	}
	return ""
}

func stableNodeForProvider(nodes map[string]string, providerID string) string {
	var names []string
	for n, p := range nodes {
		if p == providerID {
			names = append(names, n)
		}
	}
	if len(names) == 0 {
		return ""
	}
	sort.Strings(names)
	return names[0]
}
