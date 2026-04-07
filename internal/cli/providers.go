package cli

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/yashg4509/perch/internal/provider"
	rootproviders "github.com/yashg4509/perch/providers"
)

// loadRegistryForProject loads YAML from PERCH_PROVIDERS_DIR, else providers/ under projectRoot
// when it contains non-underscore *.yaml (any depth), else the bundled definitions shipped with the binary.
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
	var found bool
	_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		n := d.Name()
		if !strings.HasSuffix(n, ".yaml") && !strings.HasSuffix(n, ".yml") {
			return nil
		}
		base := strings.TrimSuffix(strings.TrimSuffix(n, ".yaml"), ".yml")
		if strings.HasPrefix(base, "_") {
			return nil
		}
		found = true
		return filepath.SkipAll
	})
	return found
}
