package provider

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Registry maps provider id (YAML name) to parsed spec.
type Registry struct {
	ByName map[string]*Spec
}

// LoadRegistry reads every *.yaml / *.yml in dir (non-recursive), skipping files whose
// basename starts with '_' (e.g. _template.yaml).
func LoadRegistry(dir string) (*Registry, error) {
	entries, err := os.ReadDir(dir)
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
		path := filepath.Join(dir, name)
		raw, err := os.ReadFile(path)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", path, err))
			continue
		}
		spec, err := ParseProviderYAML(raw)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", path, err))
			continue
		}
		if _, dup := reg.ByName[spec.Name]; dup {
			errs = append(errs, fmt.Errorf("%s: duplicate provider name %q", path, spec.Name))
			continue
		}
		reg.ByName[spec.Name] = spec
	}
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}
	return reg, nil
}
