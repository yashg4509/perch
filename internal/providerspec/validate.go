package providerspec

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ValidateProviderYAML checks required fields for a single provider YAML document.
func ValidateProviderYAML(data []byte) error {
	var root map[string]any
	if err := yaml.Unmarshal(data, &root); err != nil {
		return fmt.Errorf("yaml: %w", err)
	}
	name, _ := root["name"].(string)
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("provider: missing or empty name")
	}
	cat, _ := root["category"].(string)
	if strings.TrimSpace(cat) == "" {
		return fmt.Errorf("provider %q: missing category", name)
	}
	deployable, ok := root["deployable"].(bool)
	if !ok {
		return fmt.Errorf("provider %q: deployable must be a boolean", name)
	}
	if deployable {
		if err := requireMap(root, "cli", name); err != nil {
			return err
		}
		if err := requireMap(root, "api", name); err != nil {
			return err
		}
		if err := requireMap(root, "credentials", name); err != nil {
			return err
		}
	} else {
		if err := requireMap(root, "api", name); err != nil {
			return err
		}
		if err := requireMap(root, "credentials", name); err != nil {
			return err
		}
	}
	return nil
}

func requireMap(root map[string]any, key, providerName string) error {
	v, ok := root[key]
	if !ok || v == nil {
		return fmt.Errorf("provider %q: missing %s", providerName, key)
	}
	if _, ok := v.(map[string]any); !ok {
		return fmt.Errorf("provider %q: %s must be a mapping", providerName, key)
	}
	return nil
}

// ValidateProviderYAMLDir validates every *.yaml file in dir (non-recursive).
func ValidateProviderYAMLDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".yaml") && !strings.HasSuffix(name, ".yml") {
			continue
		}
		path := filepath.Join(dir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := ValidateProviderYAML(data); err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
	}
	return nil
}
