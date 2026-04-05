package detect

import "sort"

// UniqueProviders returns sorted unique provider ids from signals.
func UniqueProviders(sigs []Signal) []string {
	seen := map[string]struct{}{}
	for _, s := range sigs {
		for _, p := range s.Providers {
			seen[p] = struct{}{}
		}
	}
	out := make([]string, 0, len(seen))
	for p := range seen {
		out = append(out, p)
	}
	sort.Strings(out)
	return out
}
