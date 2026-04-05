package customstatus

import (
	"context"
	"fmt"
	"os"
	"runtime"

	"github.com/yashg4509/perch/internal/provider"
	"github.com/yashg4509/perch/pkg/exec"
)

// Run executes a custom status shell line (perch.yaml `status:` for provider custom).
// Exit code 0 maps to healthy; non-zero or context errors map to unhealthy.
func Run(ctx context.Context, cmdline string) (provider.NodeStatus, error) {
	if cmdline == "" {
		return provider.NodeStatus{}, fmt.Errorf("customstatus: empty status command")
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
	if err != nil {
		return provider.NodeStatus{Healthy: false}, nil
	}
	return provider.NodeStatus{Healthy: res.ExitCode == 0}, nil
}
