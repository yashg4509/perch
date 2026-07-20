package credentials

import (
	"encoding/base64"
	"strings"

	"github.com/yashg4509/perch/internal/provider"
)

// ResolveFromEnv maps provider credential keys to secret values found in env.
// Direct aliases come from specs; composite rules fill keys like cloudinary_api_basic.
func ResolveFromEnv(env map[string]string, specs []provider.CredentialsSpec) map[string]string {
	out := make(map[string]string)
	seen := make(map[string]bool)

	for _, spec := range specs {
		if spec.Key == "" || seen[spec.Key] {
			continue
		}
		for _, alias := range spec.EnvAliases {
			if v, ok := env[alias]; ok && strings.TrimSpace(v) != "" {
				out[spec.Key] = strings.TrimSpace(v)
				seen[spec.Key] = true
				break
			}
		}
	}

	applyComposites(env, out, seen)
	return out
}

func applyComposites(env map[string]string, out map[string]string, seen map[string]bool) {
	if !seen["cloudinary_api_basic"] {
		k := strings.TrimSpace(env["CLOUDINARY_API_KEY"])
		s := strings.TrimSpace(env["CLOUDINARY_API_SECRET"])
		if k != "" && s != "" {
			out["cloudinary_api_basic"] = base64.StdEncoding.EncodeToString([]byte(k + ":" + s))
			seen["cloudinary_api_basic"] = true
		}
	}
	if !seen["upstash_management_basic"] {
		email := strings.TrimSpace(env["UPSTASH_EMAIL"])
		apiKey := strings.TrimSpace(env["UPSTASH_API_KEY"])
		if email != "" && apiKey != "" {
			out["upstash_management_basic"] = base64.StdEncoding.EncodeToString([]byte(email + ":" + apiKey))
			seen["upstash_management_basic"] = true
		}
	}
}

// UniqueCredentialSpecs deduplicates credentials by key from a provider registry.
func UniqueCredentialSpecs(reg *provider.Registry) []provider.CredentialsSpec {
	if reg == nil {
		return nil
	}
	seen := make(map[string]bool)
	var out []provider.CredentialsSpec
	for _, spec := range reg.ByName {
		k := spec.Credentials.Key
		if k == "" || seen[k] {
			continue
		}
		seen[k] = true
		out = append(out, spec.Credentials)
	}
	return out
}
