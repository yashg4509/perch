package credentials

import (
	"fmt"
	"os"
	"sort"

	"github.com/yashg4509/perch/internal/provider"
)

// ImportOptions controls env → credentials import behavior.
type ImportOptions struct {
	Overwrite bool
	DryRun    bool
}

// ImportResult reports which credential keys were imported, skipped, or missing.
type ImportResult struct {
	Imported []string `json:"imported"`
	Skipped  []string `json:"skipped"`
	Missing  []string `json:"missing"`
}

// ImportEnvFile reads envPath and writes matching secrets into store.
// Secrets are never logged; only credential keys appear in ImportResult.
func ImportEnvFile(store *Store, envPath string, specs []provider.CredentialsSpec, opts ImportOptions) (ImportResult, error) {
	if store == nil {
		return ImportResult{}, fmt.Errorf("credentials: nil store")
	}
	raw, err := os.ReadFile(envPath)
	if err != nil {
		return ImportResult{}, err
	}
	env, err := ParseDotenv(string(raw))
	if err != nil {
		return ImportResult{}, err
	}
	resolved := ResolveFromEnv(env, specs)

	var imported, skipped, missing []string
	for _, spec := range specs {
		if spec.Key == "" {
			continue
		}
		if _, ok := resolved[spec.Key]; !ok {
			missing = append(missing, spec.Key)
		}
	}

	keys := make([]string, 0, len(resolved))
	for k := range resolved {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		val := resolved[key]
		if val == "" {
			continue
		}
		if !opts.Overwrite {
			if existing, has, err := store.Get(key); err != nil {
				return ImportResult{}, err
			} else if has && existing != "" {
				skipped = append(skipped, key)
				continue
			}
		}
		if opts.DryRun {
			imported = append(imported, key)
			continue
		}
		if err := store.Set(key, val); err != nil {
			return ImportResult{}, err
		}
		imported = append(imported, key)
	}

	sort.Strings(imported)
	sort.Strings(skipped)
	sort.Strings(missing)
	return ImportResult{Imported: imported, Skipped: skipped, Missing: missing}, nil
}
