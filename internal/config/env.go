package config

import "fmt"

// EnvironmentNodes returns the node map for env, or an error if env is missing or empty.
func (c *Config) EnvironmentNodes(env string) (map[string]Node, error) {
	if c == nil {
		return nil, fmt.Errorf("config: nil config")
	}
	if env == "" {
		return nil, fmt.Errorf("config: environment name is required")
	}
	m, ok := c.Environments[env]
	if !ok {
		return nil, fmt.Errorf("config: unknown environment %q", env)
	}
	return m, nil
}
