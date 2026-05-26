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
	if hint.Len() == 0 {
		hint.WriteString(strings.TrimSpace(spec.Credentials.Prompt))
	}
	return hint.String()
}
