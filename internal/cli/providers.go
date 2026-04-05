package cli

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/yashg4509/perch/internal/provider"
	rootproviders "github.com/yashg4509/perch/providers"
)

// loadRegistryForProject loads YAML from PERCH_PROVIDERS_DIR, else providers/ under projectRoot
// when it contains non-underscore *.yaml, else the bundled definitions shipped with the binary.
func loadRegistryForProject(projectRoot string) (*provider.Registry, error) {
	if d := strings.TrimSpace(os.Getenv("PERCH_PROVIDERS_DIR")); d != "" {
		return provider.LoadRegistry(d)
	}
	local := strings.TrimSpace(projectRoot)
	if local != "" {
		p := filepath.Join(local, "providers")
		if hasProviderYAML(p) {
			return provider.LoadRegistry(p)
		}
	}
	return provider.LoadRegistryFS(rootproviders.Files())
}

func hasProviderYAML(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		n := e.Name()
		if !strings.HasSuffix(n, ".yaml") && !strings.HasSuffix(n, ".yml") {
			continue
		}
		base := strings.TrimSuffix(strings.TrimSuffix(n, ".yaml"), ".yml")
		if strings.HasPrefix(base, "_") {
			continue
		}
		return true
	}
	return false
}
