package exec

import (
	"bytes"
	"context"
	"fmt"
	osexec "os/exec"
)

// Result is the outcome of a finished subprocess (exit code and captured I/O).
type Result struct {
	Stdout   []byte
	Stderr   []byte
	ExitCode int
}

// Options configures environment and working directory for [Run].
type Options struct {
	Env []string
	Dir string
}

// Run starts a subprocess with context cancellation / deadline and captures stdout and stderr.
// If the process exits non-zero, err is nil and Result.ExitCode is set.
// If the context is cancelled or times out before the process exits, err is non-nil.
func Run(ctx context.Context, name string, args []string, opts Options) (*Result, error) {
	// #nosec G204 — subprocess is the primitive for CLI integrations; callers must not pass
	// untrusted shell strings as argv (user-defined commands use explicit sh -c in customstatus).
	cmd := osexec.CommandContext(ctx, name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Dir = opts.Dir
	if len(opts.Env) > 0 {
		cmd.Env = opts.Env
	}
	err := cmd.Run()
	res := &Result{
		Stdout: stdout.Bytes(),
		Stderr: stderr.Bytes(),
	}
	if err != nil {
		if ctx.Err() != nil {
			return res, fmt.Errorf("exec: %w", ctx.Err())
		}
		if ee, ok := err.(*osexec.ExitError); ok {
			res.ExitCode = ee.ExitCode()
			return res, nil
		}
		return res, fmt.Errorf("exec: %w", err)
	}
	return res, nil
}
