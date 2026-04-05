package providers

import "embed"

//go:embed *.yaml
var bundledYAML embed.FS

// Files returns the embedded provider YAML tree (non-recursive; skips underscore files when loaded via [provider.LoadRegistryFS]).
func Files() embed.FS {
	return bundledYAML
}
