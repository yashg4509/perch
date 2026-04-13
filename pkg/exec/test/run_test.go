package exec_test

import (
	"bytes"
	"context"
	"os"
	osexec "os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	perchexec "github.com/yashg4509/perch/pkg/exec"
)

func TestRun_echoStdout(t *testing.T) {
	echo := mustEcho(t)
	ctx := context.Background()
	res, err := perchexec.Run(ctx, echo, []string{"hello", "perch"}, perchexec.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if res.ExitCode != 0 {
		t.Fatalf("exit %d", res.ExitCode)
	}
	if got := strings.TrimSpace(string(res.Stdout)); got != "hello perch" {
		t.Fatalf("stdout %q", got)
	}
}

func TestRun_stderrCaptured(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses /bin/sh -c")
	}
	ctx := context.Background()
	res, err := perchexec.Run(ctx, "/bin/sh", []string{"-c", `echo out >&1; echo err >&2`}, perchexec.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if res.ExitCode != 0 {
		t.Fatalf("exit %d stderr=%q", res.ExitCode, res.Stderr)
	}
	if !bytes.Contains(res.Stdout, []byte("out")) {
		t.Fatalf("stdout %q", res.Stdout)
	}
	if !bytes.Contains(res.Stderr, []byte("err")) {
		t.Fatalf("stderr %q", res.Stderr)
	}
}

func TestRun_nonZeroExit(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses /bin/sh -c")
	}
	ctx := context.Background()
	res, err := perchexec.Run(ctx, "/bin/sh", []string{"-c", "exit 7"}, perchexec.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if res.ExitCode != 7 {
		t.Fatalf("exit %d", res.ExitCode)
	}
}

func TestRun_timeout(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses /bin/sleep")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	_, err := perchexec.Run(ctx, "/bin/sleep", []string{"10"}, perchexec.Options{})
	if err == nil {
		t.Fatal("expected timeout")
	}
	if !strings.Contains(err.Error(), "context") && !strings.Contains(err.Error(), "deadline") {
		t.Fatalf("err = %v", err)
	}
}

func TestRun_extraEnv(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses /bin/sh -c")
	}
	ctx := context.Background()
	res, err := perchexec.Run(ctx, "/bin/sh", []string{"-c", `printf '%s' "$PERCH_TEST_VAR"`}, perchexec.Options{
		Env: append(append([]string{}, filteredEnv()...), "PERCH_TEST_VAR=from-exec"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if string(res.Stdout) != "from-exec" {
		t.Fatalf("got %q", res.Stdout)
	}
}

func TestRun_dir(t *testing.T) {
	tmp := t.TempDir()
	sub := filepath.Join(tmp, "sub")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS == "windows" {
		t.Skip("uses pwd via sh")
	}
	ctx := context.Background()
	res, err := perchexec.Run(ctx, "/bin/sh", []string{"-c", "basename $(pwd)"}, perchexec.Options{Dir: sub})
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(res.Stdout)) != "sub" {
		t.Fatalf("stdout %q", res.Stdout)
	}
}

func mustEcho(t *testing.T) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		p, err := osexec.LookPath("echo")
		if err != nil {
			t.Skip("no echo in PATH")
		}
		return p
	}
	return "/bin/echo"
}

func filteredEnv() []string {
	// Minimal inherit for subprocess tests that need HOME etc.
	var out []string
	for _, e := range os.Environ() {
		if strings.HasPrefix(e, "PATH=") || strings.HasPrefix(e, "HOME=") || strings.HasPrefix(e, "USER=") {
			out = append(out, e)
		}
	}
	return out
}
