package providers

import "embed"

// Bundled provider specs live in category subfolders; _template stays at repo root for copy-paste.
//
//go:embed _template.yaml
//go:embed github.yaml
//go:embed hosting/*.yaml
//go:embed data/*.yaml
//go:embed saas/*.yaml
//go:embed workflows/*.yaml
//go:embed ai/*.yaml
//go:embed observability/*.yaml
var bundledYAML embed.FS

// Files returns the embedded provider YAML tree (recursive load via [provider.LoadRegistryFS]).
func Files() embed.FS {
	return bundledYAML
}
