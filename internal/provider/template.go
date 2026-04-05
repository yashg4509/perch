package provider

import (
	"fmt"
	"net/http"
	"strings"
)

// SubstitutePlaceholders replaces `{key}` segments using vars (one code path for CLI/API templates).
func SubstitutePlaceholders(s string, vars map[string]string) string {
	out := s
	for k, v := range vars {
		out = strings.ReplaceAll(out, "{"+k+"}", v)
	}
	return out
}

// ApplyAuthHeader sets request headers from an auth_header template like "Authorization: Bearer {token}".
func ApplyAuthHeader(req *http.Request, authHeaderTemplate string, vars map[string]string) error {
	line := SubstitutePlaceholders(authHeaderTemplate, vars)
	key, val, ok := strings.Cut(line, ": ")
	if !ok || strings.TrimSpace(key) == "" {
		return fmt.Errorf("provider: auth_header must look like 'Name: value', got %q", authHeaderTemplate)
	}
	req.Header.Set(strings.TrimSpace(key), strings.TrimSpace(val))
	return nil
}
