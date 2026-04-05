package scaffold

import (
	"fmt"
	"sort"
)

// AssignNodeNames maps stable logical node names to provider ids (deterministic for idempotent init).
func AssignNodeNames(providers []string) map[string]string {
	sorted := append([]string(nil), providers...)
	sort.Strings(sorted)
	out := make(map[string]string)
	used := map[string]bool{}
	providerPlaced := map[string]bool{}

	try := func(nodeName, prov string) {
		if providerPlaced[prov] || used[nodeName] {
			return
		}
		out[nodeName] = prov
		used[nodeName] = true
		providerPlaced[prov] = true
	}

	for _, p := range sorted {
		switch p {
		case "vercel", "netlify", "cloudflare":
			try("frontend", p)
		case "render", "railway", "fly":
			try("backend", p)
		case "supabase":
			try("db", p)
		case "stripe":
			try("payments", p)
		case "openai", "anthropic":
			try("llm", p)
		}
	}

	for _, p := range sorted {
		if providerPlaced[p] {
			continue
		}
		base := p
		for i := 0; ; i++ {
			name := base
			if i > 0 {
				name = fmt.Sprintf("%s_%d", base, i)
			}
			if !used[name] {
				out[name] = p
				used[name] = true
				providerPlaced[p] = true
				break
			}
		}
	}
	return out
}
