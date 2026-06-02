package config

import (
	"net/url"
	"strings"

	"github.com/yashg4509/perch/internal/provider"
)

// InferProject returns a non-secret project id from dotenv using provider conventions.
func InferProject(providerID string, spec *provider.Spec, node Node, env map[string]string) string {
	if env == nil {
		return ""
	}
	if key := strings.TrimSpace(node.EnvProject); key != "" {
		if v := strings.TrimSpace(env[key]); v != "" {
			return v
		}
	}
	if spec != nil {
		for _, alias := range spec.ProjectEnvAliases {
			if v := strings.TrimSpace(env[alias]); v != "" {
				return v
			}
		}
	}
	switch providerID {
	case "neon":
		return inferNeonProject(env)
	case "sentry":
		return inferSentryProject(env)
	case "posthog":
		return inferPosthogProject(env)
	case "github":
		return inferGitHubProject(env)
	default:
		return ""
	}
}

// InferService returns a service id from dotenv (deployable hosts).
func InferService(providerID string, spec *provider.Spec, node Node, env map[string]string) string {
	if env == nil {
		return ""
	}
	if key := strings.TrimSpace(node.EnvService); key != "" {
		if v := strings.TrimSpace(env[key]); v != "" {
			return v
		}
	}
	if spec != nil {
		for _, alias := range spec.ServiceEnvAliases {
			if v := strings.TrimSpace(env[alias]); v != "" {
				return v
			}
		}
	}
	return ""
}

// EffectiveProject uses perch.yaml project or infers from .env when still a placeholder.
func EffectiveProject(n Node, spec *provider.Spec, env map[string]string) string {
	if !IsPlaceholder(n.Project) {
		return n.Project
	}
	providerID := n.Provider
	if spec != nil {
		providerID = spec.Name
	}
	if v := InferProject(providerID, spec, n, env); v != "" {
		return v
	}
	return n.Project
}

// EffectiveService uses perch.yaml service or infers from .env when still a placeholder.
func EffectiveService(n Node, spec *provider.Spec, env map[string]string) string {
	if !IsPlaceholder(n.Service) {
		return n.Service
	}
	providerID := n.Provider
	if spec != nil {
		providerID = spec.Name
	}
	if v := InferService(providerID, spec, n, env); v != "" {
		return v
	}
	return n.Service
}

// ProviderConfigured reports whether a node is wired for status, using .env inference.
func ProviderConfigured(n Node, spec *provider.Spec, env map[string]string, hasCredential func(key string) (bool, error)) (bool, string) {
	if spec == nil {
		return false, "unknown provider"
	}
	if spec.Deployable {
		proj := EffectiveProject(n, spec, env)
		svc := EffectiveService(n, spec, env)
		if IsPlaceholder(proj) && IsPlaceholder(svc) {
			return false, "set project or service in perch.yaml, or add matching vars to .env"
		}
		if svc == "" && IsPlaceholder(proj) {
			return false, "set project in perch.yaml or .env"
		}
		return true, ""
	}
	if spec.Name == "inngest" && env["INNGEST_DEV"] == "1" {
		return true, ""
	}
	if spec.Name == "neon" {
		db := strings.TrimSpace(env["DATABASE_URL"])
		if db != "" && strings.Contains(db, "neon") {
			return true, ""
		}
	}
	proj := EffectiveProject(n, spec, env)
	if proj != "" && IsPlaceholder(proj) {
		return false, "set project in perch.yaml or add provider env vars to .env"
	}
	if spec.Credentials.Key != "" {
		if hasCredential == nil {
			return false, "missing credential (run perch auth sync-env)"
		}
		ok, err := hasCredential(spec.Credentials.Key)
		if err != nil {
			return false, "credentials store error"
		}
		if !ok {
			return false, "missing credential (run perch auth sync-env)"
		}
	}
	if proj == "" && needsProjectID(spec.Name) {
		return false, "set project in perch.yaml or .env"
	}
	return true, ""
}

func needsProjectID(providerID string) bool {
	switch providerID {
	case "openai", "anthropic", "inngest":
		return false
	default:
		return true
	}
}

// ApplyEnvInference fills placeholder project/service on cfg from dotenv and provider metadata.
func ApplyEnvInference(cfg *Config, envName string, dotenv map[string]string, reg *provider.Registry) []string {
	if cfg == nil || reg == nil {
		return nil
	}
	nodes, ok := cfg.Environments[envName]
	if !ok {
		return nil
	}
	var applied []string
	for name, n := range nodes {
		spec := reg.ByName[n.Provider]
		if spec == nil {
			continue
		}
		changed := false
		if IsPlaceholder(n.Project) {
			if v := InferProject(spec.Name, spec, n, dotenv); v != "" {
				n.Project = v
				applied = append(applied, name+".project")
				changed = true
			}
		}
		if IsPlaceholder(n.Service) {
			if v := InferService(spec.Name, spec, n, dotenv); v != "" {
				n.Service = v
				applied = append(applied, name+".service")
				changed = true
			}
		}
		if changed {
			nodes[name] = n
		}
	}
	cfg.Environments[envName] = nodes
	return applied
}

func inferNeonProject(env map[string]string) string {
	if id := strings.TrimSpace(env["NEON_PROJECT_ID"]); id != "" {
		return id
	}
	db := strings.TrimSpace(env["DATABASE_URL"])
	if db == "" || !strings.Contains(db, "neon") {
		return ""
	}
	u, err := url.Parse(db)
	if err != nil {
		return ""
	}
	host := u.Hostname()
	if idx := strings.Index(host, "."); idx > 0 {
		ep := host[:idx]
		if strings.HasPrefix(ep, "ep-") {
			return ep
		}
	}
	return ""
}

func inferSentryProject(env map[string]string) string {
	for _, key := range []string{"SENTRY_DSN", "NEXT_PUBLIC_SENTRY_DSN"} {
		if v := parseSentryDSNProject(env[key]); v != "" {
			return v
		}
	}
	return ""
}

func parseSentryDSNProject(dsn string) string {
	dsn = strings.TrimSpace(dsn)
	if dsn == "" {
		return ""
	}
	u, err := url.Parse(dsn)
	if err != nil {
		return ""
	}
	path := strings.Trim(u.Path, "/")
	if path == "" {
		return ""
	}
	// https://key@host/PROJECT_ID
	if i := strings.LastIndex(path, "/"); i >= 0 {
		path = path[i+1:]
	}
	return path
}

func inferPosthogProject(env map[string]string) string {
	for _, key := range []string{"POSTHOG_PROJECT_ID", "NEXT_PUBLIC_POSTHOG_PROJECT_ID"} {
		if v := strings.TrimSpace(env[key]); v != "" {
			return v
		}
	}
	return ""
}

func inferGitHubProject(env map[string]string) string {
	for _, key := range []string{"GITHUB_REPOSITORY", "GITHUB_REPO"} {
		if v := strings.TrimSpace(env[key]); v != "" {
			return v
		}
	}
	return ""
}
