// Package customlogs runs user-defined shell one-liners from perch.yaml for provider "custom"
// logs capture. Same trust model as [github.com/yashg4509/perch/internal/customstatus]: only run
// stacks you trust.
package customlogs

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/yashg4509/perch/pkg/exec"
)

// MaxCapturedBytes is the upper bound on combined stdout+stderr kept in memory after a run.
const MaxCapturedBytes = 512 * 1024

// MaxLines is the maximum number of lines returned from each of stdout and stderr.
const MaxLines = 4000

// Result is bounded log capture from a single shell invocation.
type Result struct {
	StdoutLines []string
	StderrLines []string
	ExitCode    int
	Truncated   bool
	TimedOut    bool
	RunError    string // non-empty when the process could not be started or context failed before exit
}

// Run executes cmdline via the system shell, captures stdout/stderr, and splits into lines.
// ctx should carry a deadline; on timeout partial output may be present with TimedOut set.
func Run(ctx context.Context, cmdline string) (*Result, error) {
	cmdline = strings.TrimSpace(cmdline)
	if cmdline == "" {
		return nil, fmt.Errorf("customlogs: empty logs command")
	}
	var name string
	var args []string
	if runtime.GOOS == "windows" {
		name = "cmd"
		args = []string{"/C", cmdline}
	} else {
		name = "/bin/sh"
		args = []string{"-c", cmdline}
	}
	res, err := exec.Run(ctx, name, args, exec.Options{Env: os.Environ()})
	out := &Result{}
	if err != nil {
		if ctx.Err() != nil {
			out.TimedOut = ctx.Err() == context.DeadlineExceeded
			out.RunError = err.Error()
			// Still return stdout/stderr fragments if any were captured.
			out.StdoutLines, out.StderrLines, out.Truncated = splitAndTruncate(res.Stdout, res.Stderr)
			return out, nil
		}
		return nil, err
	}
	out.ExitCode = res.ExitCode
	out.StdoutLines, out.StderrLines, out.Truncated = splitAndTruncate(res.Stdout, res.Stderr)
	return out, nil
}

func splitAndTruncate(stdout, stderr []byte) (outLines, errLines []string, truncated bool) {
	stdout, stderr, truncated = truncateBuffers(stdout, stderr)
	outLines = splitLines(stdout, MaxLines)
	errLines = splitLines(stderr, MaxLines)
	return outLines, errLines, truncated
}

func truncateBuffers(stdout, stderr []byte) ([]byte, []byte, bool) {
	truncated := false
	if len(stdout)+len(stderr) <= MaxCapturedBytes {
		return stdout, stderr, false
	}
	truncated = true
	// Prefer stdout; give stderr a small tail slice of the budget.
	const stderrBudget = 16 * 1024
	if len(stderr) > stderrBudget {
		stderr = stderr[len(stderr)-stderrBudget:]
	}
	remain := MaxCapturedBytes - len(stderr)
	if remain < 0 {
		remain = 0
	}
	if len(stdout) > remain {
		stdout = stdout[len(stdout)-remain:]
	}
	return stdout, stderr, truncated
}

func splitLines(b []byte, max int) []string {
	if len(b) == 0 {
		return nil
	}
	s := strings.ReplaceAll(string(b), "\r\n", "\n")
	raw := strings.Split(s, "\n")
	if len(raw) > 0 && raw[len(raw)-1] == "" {
		raw = raw[:len(raw)-1]
	}
	if len(raw) > max {
		return raw[len(raw)-max:]
	}
	return raw
}
