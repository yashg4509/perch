package config

import "strings"

// IsPlaceholder reports whether a perch.yaml project/service value is an unfilled template.
func IsPlaceholder(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return true
	}
	upper := strings.ToUpper(s)
	if strings.HasPrefix(upper, "CHANGE_ME") {
		return true
	}
	if strings.HasPrefix(upper, "YOUR_") {
		return true
	}
	return IsDevSentinelProject(s)
}

// IsDevSentinelProject marks perch.yaml filler ids from perch config sync-env defaults.
// They are not real vendor resource ids and must not imply the integration is set up.
func IsDevSentinelProject(s string) bool {
	switch strings.TrimSpace(strings.ToLower(s)) {
	case "local-dev", "local-neon":
		return true
	default:
		return false
	}
}
