package provider

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"
)

// Registry maps provider id (YAML name) to parsed spec.
type Registry struct {
	ByName map[string]*Spec
}

// LoadRegistry reads every *.yaml / *.yml under dir (recursive), skipping files whose
// basename starts with '_' (e.g. _template.yaml).
func LoadRegistry(dir string) (*Registry, error) {
	return LoadRegistryFS(os.DirFS(dir))
}

// LoadRegistryFS reads every *.yaml / *.yml under fsys (recursive), skipping basenames
// that start with '_' (e.g. _template.yaml). Same semantics as [LoadRegistry].
func LoadRegistryFS(fsys fs.FS) (*Registry, error) {
	if fsys == nil {
		return nil, fmt.Errorf("provider: nil fs")
	}
	reg := &Registry{ByName: make(map[string]*Spec)}
	var errs []error
	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		name := d.Name()
		if !strings.HasSuffix(name, ".yaml") && !strings.HasSuffix(name, ".yml") {
			return nil
		}
		base := strings.TrimSuffix(strings.TrimSuffix(name, ".yaml"), ".yml")
		if strings.HasPrefix(base, "_") {
			return nil
		}
		raw, err := fs.ReadFile(fsys, path)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", path, err))
			return nil
		}
		spec, err := ParseProviderYAML(raw)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", path, err))
			return nil
		}
		if _, dup := reg.ByName[spec.Name]; dup {
			errs = append(errs, fmt.Errorf("%s: duplicate provider name %q", path, spec.Name))
			return nil
		}
		reg.ByName[spec.Name] = spec
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}
	return reg, nil
}
