package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const vizAPITestYAML = `name: viz-api-test
environments:
  dev:
    web:
      provider: custom
      status: "echo ok"
      logs: "printf 'line-1\nline-2\n'"
    worker:
      provider: custom
      status: "echo ok"
    api:
      provider: vercel
      project: sample
edges:
  - web -> api
`

func TestServeLogsJSON_CustomNode(t *testing.T) {
	t.Chdir(writeVizTestStack(t))

	req := httptest.NewRequest(http.MethodGet, "/api/logs?env=dev&node=web", nil)
	rr := httptest.NewRecorder()

	serveLogsJSON(rr, req, "dev")
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}

	var got struct {
		StdoutLines []string `json:"stdout_lines"`
		StderrLines []string `json:"stderr_lines"`
		ExitCode    int      `json:"exit_code"`
		Truncated   bool     `json:"truncated"`
		TimedOut    bool     `json:"timed_out"`
		RunError    string   `json:"run_error"`
		Source      string   `json:"source"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("invalid json: %v; body=%s", err, rr.Body.String())
	}
	if got.ExitCode != 0 {
		t.Fatalf("exit_code=%d", got.ExitCode)
	}
	if len(got.StdoutLines) != 2 || got.StdoutLines[0] != "line-1" || got.StdoutLines[1] != "line-2" {
		t.Fatalf("stdout_lines=%v", got.StdoutLines)
	}
	if len(got.StderrLines) != 0 || got.RunError != "" || got.TimedOut || got.Truncated {
		t.Fatalf("unexpected payload: %+v", got)
	}
	if got.Source != "custom" {
		t.Fatalf("source=%q want custom", got.Source)
	}
}

func TestServeLogsJSON_ValidationErrors(t *testing.T) {
	t.Chdir(writeVizTestStack(t))

	tests := []struct {
		name   string
		url    string
		status int
	}{
		{
			name:   "missing node",
			url:    "/api/logs?env=dev",
			status: http.StatusBadRequest,
		},
		{
			name:   "unknown env",
			url:    "/api/logs?env=prod&node=web",
			status: http.StatusBadRequest,
		},
		{
			name:   "unknown node",
			url:    "/api/logs?env=dev&node=missing",
			status: http.StatusBadRequest,
		},
		{
			name:   "vercel provider without credentials",
			url:    "/api/logs?env=dev&node=api",
			status: http.StatusOK,
		},
		{
			name:   "custom without logs command",
			url:    "/api/logs?env=dev&node=worker",
			status: http.StatusBadRequest,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.url, nil)
			rr := httptest.NewRecorder()
			serveLogsJSON(rr, req, "dev")
			if rr.Code != tc.status {
				t.Fatalf("status=%d want=%d body=%s", rr.Code, tc.status, rr.Body.String())
			}
			var got map[string]any
			if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
				t.Fatalf("invalid json: %v; body=%s", err, rr.Body.String())
			}
			if tc.status >= 400 {
				if got["error"] == "" {
					t.Fatalf("expected error message in %v", got)
				}
				return
			}
			if tc.name == "vercel provider without credentials" {
				if got["source"] != "none" {
					t.Fatalf("source=%v", got["source"])
				}
				if got["setup_hint"] == "" {
					t.Fatalf("expected setup_hint in %v", got)
				}
			}
		})
	}
}

func TestServeCredentialsPost(t *testing.T) {
	dir := writeVizTestStack(t)
	t.Chdir(dir)
	t.Setenv("HOME", t.TempDir())

	body := `{"key":"vercel_token","token":"vca_test_secret"}`
	req := httptest.NewRequest(http.MethodPost, "/api/credentials", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	serveCredentialsPost(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	var got map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got["ok"] != true {
		t.Fatalf("got=%v", got)
	}
}

func TestServeCredentialsPost_validation(t *testing.T) {
	dir := writeVizTestStack(t)
	t.Chdir(dir)
	t.Setenv("HOME", t.TempDir())

	tests := []struct {
		name   string
		body   string
		status int
	}{
		{name: "unknown key", body: `{"key":"bogus","token":"x"}`, status: http.StatusBadRequest},
		{name: "empty key", body: `{"key":"","token":"x"}`, status: http.StatusBadRequest},
		{name: "empty token", body: `{"key":"vercel_token","token":""}`, status: http.StatusBadRequest},
		{name: "invalid json", body: `{`, status: http.StatusBadRequest},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/credentials", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			serveCredentialsPost(rr, req)
			if rr.Code != tc.status {
				t.Fatalf("status=%d want=%d body=%s", rr.Code, tc.status, rr.Body.String())
			}
		})
	}
}

func writeVizTestStack(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "perch.yaml"), []byte(vizAPITestYAML), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}
