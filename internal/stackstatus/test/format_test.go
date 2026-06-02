package stackstatus_test

import (
	"strings"
	"testing"

	"github.com/yashg4509/perch/internal/stackstatus"
)

func TestFormatHuman_groups(t *testing.T) {
	rep := &stackstatus.EnvReport{
		Env: "dev",
		Nodes: []stackstatus.NodeReport{
			{Name: "web", Provider: "custom", Healthy: true, StatusSource: stackstatus.SourceShell, Configured: true},
			{Name: "api", Provider: "custom", Healthy: false, StatusSource: stackstatus.SourceShell, Configured: true, Detail: "custom health command failed"},
			{Name: "openai", Provider: "openai", Healthy: false, StatusSource: stackstatus.SourceUnchecked, Configured: true},
			{Name: "billing", Provider: "stripe", Healthy: false, StatusSource: stackstatus.SourceUnconfigured, Configured: false},
		},
	}
	out := stackstatus.FormatHuman("app", "dev", rep)
	if !strings.Contains(out, "1 up · 1 down · 1 wired · 1 not wired") {
		t.Fatalf("summary missing:\n%s", out)
	}
	if !strings.Contains(out, "UP\n  web") {
		t.Fatalf("up group missing:\n%s", out)
	}
	if !strings.Contains(out, "DOWN\n  api") {
		t.Fatalf("down group missing:\n%s", out)
	}
	if !strings.Contains(out, "UNKNOWN") || !strings.Contains(out, "openai") {
		t.Fatalf("unknown group missing:\n%s", out)
	}
	if !strings.Contains(out, "SKIPPED — in perch.yaml") || !strings.Contains(out, "billing") {
		t.Fatalf("skipped group missing:\n%s", out)
	}
}

func TestGroupFor_uncheckedIsWired(t *testing.T) {
	g := stackstatus.GroupFor(stackstatus.NodeReport{
		Healthy: false, StatusSource: stackstatus.SourceUnchecked, Configured: true,
	})
	if g != stackstatus.GroupWired {
		t.Fatalf("got %v", g)
	}
}
