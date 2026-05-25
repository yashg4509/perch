package stacklogs

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

)

var (
	githubHTTPClient = &http.Client{}
	githubAPIBase    = "https://api.github.com"
)

func execLookPath(file string) (string, error) {
	return exec.LookPath(file)
}

func fetchGitHubActionsLogs(ctx context.Context, prov, project string) (LogResult, bool) {
	tok, ok := githubTokenFromEnvOrHosts()
	if !ok {
		return LogResult{}, false
	}
	owner, repo, ok := platformHooks.repoSlug()
	if !ok {
		return LogResult{}, false
	}

	if _, err := platformHooks.ghLookPath("gh"); err == nil {
		if lines, ok := ghCLIRunLogs(ctx, owner, repo); ok {
			lines = filterProviderLines(lines, prov, project)
			if len(lines) == 0 {
				return LogResult{}, false
			}
			lines, truncated := truncateLines(lines)
			return LogResult{
				Lines:     lines,
				Source:    "github_actions",
				Provider:  prov,
				Truncated: truncated,
			}, true
		}
	}

	lines, err := githubAPIDeployLogs(ctx, tok, owner, repo)
	if err != nil || len(lines) == 0 {
		return LogResult{}, false
	}
	lines = filterProviderLines(lines, prov, project)
	if len(lines) == 0 {
		return LogResult{}, false
	}
	lines, truncated := truncateLines(lines)
	return LogResult{
		Lines:     lines,
		Source:    "github_actions",
		Provider:  prov,
		Truncated: truncated,
	}, true
}

func ghCLIRunLogs(ctx context.Context, owner, repo string) ([]string, bool) {
	list := exec.CommandContext(ctx, "gh", "run", "list",
		"--repo", owner+"/"+repo,
		"--limit", "5",
		"--json", "databaseId,status,conclusion",
	)
	list.Env = append(os.Environ(), ghEnvToken()...)
	out, err := list.Output()
	if err != nil {
		return nil, false
	}
	var runs []struct {
		DatabaseID int64  `json:"databaseId"`
		Status     string `json:"status"`
		Conclusion string `json:"conclusion"`
	}
	if err := json.Unmarshal(out, &runs); err != nil || len(runs) == 0 {
		return nil, false
	}
	var runID int64
	for _, r := range runs {
		if r.DatabaseID == 0 {
			continue
		}
		runID = r.DatabaseID
		break
	}
	if runID == 0 {
		return nil, false
	}
	view := exec.CommandContext(ctx, "gh", "run", "view",
		fmt.Sprintf("%d", runID),
		"--repo", owner+"/"+repo,
		"--log",
	)
	view.Env = append(os.Environ(), ghEnvToken()...)
	logOut, err := view.Output()
	if err != nil {
		return nil, false
	}
	return splitLogLines(string(logOut)), true
}

func ghEnvToken() []string {
	var extra []string
	if t := strings.TrimSpace(platformHooks.getenv("GH_TOKEN")); t != "" {
		extra = append(extra, "GH_TOKEN="+t)
	}
	if t := strings.TrimSpace(platformHooks.getenv("GITHUB_TOKEN")); t != "" {
		extra = append(extra, "GITHUB_TOKEN="+t)
	}
	return extra
}

func githubAPIDeployLogs(ctx context.Context, token, owner, repo string) ([]string, error) {
	runsURL := fmt.Sprintf("%s/repos/%s/%s/actions/runs?per_page=5",
		strings.TrimSuffix(githubAPIBase, "/"), owner, repo)
	body, err := githubGET(ctx, token, runsURL)
	if err != nil {
		return nil, err
	}
	var runsResp struct {
		WorkflowRuns []struct {
			ID int64 `json:"id"`
		} `json:"workflow_runs"`
	}
	if err := json.Unmarshal(body, &runsResp); err != nil {
		return nil, err
	}
	if len(runsResp.WorkflowRuns) == 0 {
		return nil, fmt.Errorf("stacklogs: no workflow runs")
	}
	runID := runsResp.WorkflowRuns[0].ID
	jobsURL := fmt.Sprintf("%s/repos/%s/%s/actions/runs/%d/jobs?per_page=20",
		strings.TrimSuffix(githubAPIBase, "/"), owner, repo, runID)
	body, err = githubGET(ctx, token, jobsURL)
	if err != nil {
		return nil, err
	}
	var jobsResp struct {
		Jobs []struct {
			ID   int64  `json:"id"`
			Name string `json:"name"`
		} `json:"jobs"`
	}
	if err := json.Unmarshal(body, &jobsResp); err != nil {
		return nil, err
	}
	if len(jobsResp.Jobs) == 0 {
		return nil, fmt.Errorf("stacklogs: no workflow jobs")
	}
	jobID := jobsResp.Jobs[0].ID
	for _, j := range jobsResp.Jobs {
		name := strings.ToLower(j.Name)
		if strings.Contains(name, "deploy") || strings.Contains(name, "vercel") {
			jobID = j.ID
			break
		}
	}
	logsURL := fmt.Sprintf("%s/repos/%s/%s/actions/jobs/%d/logs",
		strings.TrimSuffix(githubAPIBase, "/"), owner, repo, jobID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, logsURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := githubHTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	logBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("stacklogs: github logs http %s", resp.Status)
	}
	return splitLogLines(string(logBody)), nil
}

func githubGET(ctx context.Context, token, rawURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := githubHTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("stacklogs: github http %s", resp.Status)
	}
	return body, nil
}

func filterProviderLines(lines []string, prov, project string) []string {
	needle := strings.ToLower(prov)
	proj := strings.ToLower(strings.TrimSpace(project))
	var filtered []string
	for _, line := range lines {
		low := strings.ToLower(line)
		if strings.Contains(low, needle) || (proj != "" && strings.Contains(low, proj)) {
			filtered = append(filtered, line)
		}
	}
	if len(filtered) > 0 {
		return filtered
	}
	return lines
}

func splitLogLines(s string) []string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	raw := strings.Split(s, "\n")
	if len(raw) > 0 && raw[len(raw)-1] == "" {
		raw = raw[:len(raw)-1]
	}
	return raw
}

var (
	gitRemoteSSH = regexp.MustCompile(`git@([^:]+):([^/]+)/(.+?)(?:\.git)?$`)
	gitRemoteURL = regexp.MustCompile(`https?://([^/]+)/([^/]+)/(.+?)(?:\.git)?$`)
)

func detectRepoSlug() (owner, repo string, ok bool) {
	wd, err := os.Getwd()
	if err != nil {
		return "", "", false
	}
	dir, err := filepath.Abs(wd)
	if err != nil {
		return "", "", false
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", "", false
		}
		dir = parent
	}
	remote, err := os.ReadFile(filepath.Join(dir, ".git", "config"))
	if err != nil {
		return parseGitRemoteFromCmd(dir)
	}
	return parseGitConfigRemotes(string(remote))
}

func parseGitConfigRemotes(config string) (owner, repo string, ok bool) {
	var url string
	for _, line := range strings.Split(config, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "url = ") {
			url = strings.TrimSpace(strings.TrimPrefix(line, "url = "))
		}
	}
	if url == "" {
		return "", "", false
	}
	return parseRemoteURL(url)
}

func parseGitRemoteFromCmd(dir string) (owner, repo string, ok bool) {
	out, err := exec.Command("git", "-C", dir, "remote", "get-url", "origin").Output()
	if err != nil {
		return "", "", false
	}
	return parseRemoteURL(strings.TrimSpace(string(out)))
}

func parseRemoteURL(raw string) (owner, repo string, ok bool) {
	raw = strings.TrimSpace(raw)
	if m := gitRemoteSSH.FindStringSubmatch(raw); len(m) == 4 {
		return m[2], strings.TrimSuffix(m[3], ".git"), true
	}
	if m := gitRemoteURL.FindStringSubmatch(raw); len(m) == 4 {
		return m[2], strings.TrimSuffix(m[3], ".git"), true
	}
	return "", "", false
}
