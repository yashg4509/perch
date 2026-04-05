package provider

import (
	"errors"
	"fmt"
	"io/fs"
	"strings"
)

// LoadRegistryFS reads every *.yaml / *.yml from the root of fsys, skipping basenames
// that start with '_' (e.g. _template.yaml). Same semantics as [LoadRegistry].
func LoadRegistryFS(fsys fs.FS) (*Registry, error) {
	if fsys == nil {
		return nil, fmt.Errorf("provider: nil fs")
	}
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return nil, err
	}
	reg := &Registry{ByName: make(map[string]*Spec)}
	var errs []error
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		base := strings.TrimSuffix(strings.TrimSuffix(name, ".yaml"), ".yml")
		if !strings.HasSuffix(name, ".yaml") && !strings.HasSuffix(name, ".yml") {
			continue
		}
		if strings.HasPrefix(base, "_") {
			continue
		}
		raw, err := fs.ReadFile(fsys, name)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", name, err))
			continue
		}
		spec, err := ParseProviderYAML(raw)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", name, err))
			continue
		}
		if _, dup := reg.ByName[spec.Name]; dup {
			errs = append(errs, fmt.Errorf("%s: duplicate provider name %q", name, spec.Name))
			continue
		}
		reg.ByName[spec.Name] = spec
	}
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}
	return reg, nil
}
