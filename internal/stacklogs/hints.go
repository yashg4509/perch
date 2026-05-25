package stacklogs

import (
	"strings"

	"github.com/yashg4509/perch/internal/provider"
)

func buildSetupHint(spec *provider.Spec) string {
	if spec == nil {
		return ""
	}
	var hint strings.Builder
	if u := strings.TrimSpace(spec.Credentials.DashboardURL); u != "" {
		hint.WriteString("Get token from here: ")
		hint.WriteString(u)
		hint.WriteString("\n")
	}
	if ev := strings.TrimSpace(spec.Credentials.EnvVar); ev != "" {
		hint.WriteString("Paste token in here: export ")
		hint.WriteString(ev)
		hint.WriteString("=your_token_here")
	}
	if hint.Len() == 0 {
		hint.WriteString(strings.TrimSpace(spec.Credentials.Prompt))
	}
	return hint.String()
}
