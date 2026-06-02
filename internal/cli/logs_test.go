package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestLogsCmd_registered(t *testing.T) {
	root := NewRootCmd()
	logs, _, err := root.Find([]string{"logs"})
	if err != nil {
		t.Fatalf("logs command: %v", err)
	}
	if logs == nil {
		t.Fatal("logs command not found")
	}

	setup, _, err := logs.Find([]string{"setup"})
	if err != nil {
		t.Fatalf("logs setup command: %v", err)
	}
	if setup == nil {
		t.Fatal("logs setup command not found")
	}
}

func TestLogsCmd_help(t *testing.T) {
	root := NewRootCmd()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{"logs", "--help"})
	buf.Reset()
	if err := root.Execute(); err != nil {
		t.Fatalf("logs --help: %v", err)
	}
	if !strings.Contains(buf.String(), "setup") {
		t.Fatalf("logs --help missing setup subcommand:\n%s", buf.String())
	}
}

func TestLogsSetupCmd_help(t *testing.T) {
	root := NewRootCmd()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{"logs", "setup", "--help"})
	if err := root.Execute(); err != nil {
		t.Fatalf("logs setup --help: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "perch viz") {
		t.Fatalf("expected setup description in help:\n%s", out)
	}
}
