package customlogs

import (
	"context"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestRun_echo(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell-specific test")
	}
	ctx := context.Background()
	res, err := Run(ctx, `printf 'a\nb\n'`)
	if err != nil {
		t.Fatal(err)
	}
	if res.RunError != "" {
		t.Fatalf("RunError: %s", res.RunError)
	}
	if res.ExitCode != 0 {
		t.Fatalf("exit: %d", res.ExitCode)
	}
	if strings.Join(res.StdoutLines, "|") != "a|b" {
		t.Fatalf("stdout lines: %#v", res.StdoutLines)
	}
}

func TestRun_emptyErrors(t *testing.T) {
	_, err := Run(context.Background(), "   ")
	if err == nil {
		t.Fatal("want error for empty command")
	}
}

func TestRun_timeout(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell-specific test")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	res, err := Run(ctx, "sleep 5")
	if err != nil {
		t.Fatal(err)
	}
	if !res.TimedOut {
		t.Fatalf("want timed out, got %#v", res)
	}
}
