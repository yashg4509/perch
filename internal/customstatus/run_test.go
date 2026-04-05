package customstatus

import (
	"context"
	"runtime"
	"testing"
)

func TestRun_trueIsHealthy(t *testing.T) {
	ctx := context.Background()
	cmd := "true"
	if runtime.GOOS == "windows" {
		cmd = "exit /b 0"
	}
	st, err := Run(ctx, cmd)
	if err != nil {
		t.Fatal(err)
	}
	if !st.Healthy {
		t.Fatal(st)
	}
}

func TestRun_falseIsUnhealthy(t *testing.T) {
	ctx := context.Background()
	cmd := "false"
	if runtime.GOOS == "windows" {
		cmd = "exit /b 1"
	}
	st, err := Run(ctx, cmd)
	if err != nil {
		t.Fatal(err)
	}
	if st.Healthy {
		t.Fatal(st)
	}
}

func TestRun_scriptInTempDir(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses /bin/sh heredoc")
	}
	ctx := context.Background()
	st, err := Run(ctx, "echo -n ok | grep -q ok")
	if err != nil {
		t.Fatal(err)
	}
	if !st.Healthy {
		t.Fatal(st)
	}
}
