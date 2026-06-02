package stacklogs

import (
	"strings"
	"testing"

	"github.com/yashg4509/perch/internal/provider"
)

func TestBuildSetupHint_dashboardURL(t *testing.T) {
	spec := &provider.Spec{
		Credentials: provider.CredentialsSpec{
			Prompt:       "fallback prompt",
			DashboardURL: "https://vercel.com/account/tokens",
			EnvVar:       "VERCEL_TOKEN",
		},
	}
	got := buildSetupHint(spec)
	if !strings.Contains(got, "Get token from here: https://vercel.com/account/tokens") {
		t.Fatalf("missing dashboard url: %q", got)
	}
	if strings.Contains(got, "export ") {
		t.Fatalf("should not include shell export hint: %q", got)
	}
}

func TestBuildSetupHint_promptFallback(t *testing.T) {
	spec := &provider.Spec{
		Credentials: provider.CredentialsSpec{Prompt: "Enter your API token"},
	}
	if got := buildSetupHint(spec); got != "Enter your API token" {
		t.Fatalf("got %q", got)
	}
}
