package provider

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// Spec is the typed provider definition loaded from providers/*.yaml.
type Spec struct {
	Name        string          `yaml:"name"`
	Category    string          `yaml:"category"`
	Deployable  bool            `yaml:"deployable"`
	LLM         bool            `yaml:"llm"`
	CLI         *CLISpec        `yaml:"cli"`
	API         APISpec         `yaml:"api"`
	Credentials CredentialsSpec `yaml:"credentials"`
}

// CLISpec is the CLI section (required when Deployable is true).
type CLISpec struct {
	Binary   string            `yaml:"binary"`
	Commands map[string]string `yaml:"commands"`
}

// APISpec is the REST section (always required).
type APISpec struct {
	BaseURL    string            `yaml:"base_url"`
	AuthHeader string            `yaml:"auth_header"`
	Endpoints  map[string]string `yaml:"endpoints"`
}

// CredentialsSpec describes stored credential metadata (not the secret value).
type CredentialsSpec struct {
	Key    string `yaml:"key"`
	Prompt string `yaml:"prompt"`
}

// ParseProviderYAML decodes and validates one provider document.
func ParseProviderYAML(raw []byte) (*Spec, error) {
	var s Spec
	if err := yaml.Unmarshal(raw, &s); err != nil {
		return nil, fmt.Errorf("provider yaml: %w", err)
	}
	if err := validateSpec(&s); err != nil {
		return nil, err
	}
	return &s, nil
}

func validateSpec(s *Spec) error {
	if strings.TrimSpace(s.Name) == "" {
		return fmt.Errorf("provider: missing name")
	}
	if strings.TrimSpace(s.Category) == "" {
		return fmt.Errorf("provider %q: missing category", s.Name)
	}
	if strings.TrimSpace(s.API.BaseURL) == "" {
		return fmt.Errorf("provider %q: api.base_url is required", s.Name)
	}
	if strings.TrimSpace(s.API.AuthHeader) == "" {
		return fmt.Errorf("provider %q: api.auth_header is required", s.Name)
	}
	if len(s.API.Endpoints) == 0 {
		return fmt.Errorf("provider %q: api.endpoints is required", s.Name)
	}
	st := strings.TrimSpace(s.API.Endpoints["status"])
	if st == "" {
		return fmt.Errorf("provider %q: api.endpoints.status is required", s.Name)
	}
	if strings.TrimSpace(s.Credentials.Key) == "" {
		return fmt.Errorf("provider %q: credentials.key is required", s.Name)
	}
	if strings.TrimSpace(s.Credentials.Prompt) == "" {
		return fmt.Errorf("provider %q: credentials.prompt is required", s.Name)
	}

	if s.Deployable {
		if s.CLI == nil {
			return fmt.Errorf("provider %q: cli is required when deployable is true", s.Name)
		}
		if strings.TrimSpace(s.CLI.Binary) == "" {
			return fmt.Errorf("provider %q: cli.binary is required when deployable is true", s.Name)
		}
		if len(s.CLI.Commands) == 0 {
			return fmt.Errorf("provider %q: cli.commands is required when deployable is true", s.Name)
		}
		if strings.TrimSpace(s.CLI.Commands["status"]) == "" {
			return fmt.Errorf("provider %q: cli.commands.status is required when deployable is true", s.Name)
		}
	}
	return nil
}
