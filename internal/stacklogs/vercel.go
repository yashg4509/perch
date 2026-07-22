package stacklogs

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/yashg4509/perch/internal/config"
	"github.com/yashg4509/perch/internal/provider"
)

// TODO: github_actions strategy for deploy log history (v2)

func resolveVercel(ctx context.Context, nodeName string, n config.Node, reg *provider.Registry, autoSetupAttempted bool) (LogResult, error) {
	project := strings.TrimSpace(n.Project)
	if project == "" {
		return LogResult{
			Source:    "none",
			Provider:  "vercel",
			SetupHint: fmt.Sprintf("node %q needs a project field for Vercel logs", nodeName),
		}, nil
	}
	spec := reg.ByName["vercel"]
	if spec == nil {
		return LogResult{}, fmt.Errorf("stacklogs: vercel provider spec missing from registry")
	}

	var tried []StrategyResult

	// 1. auth_file
	if tok, hasTok := readVercelAuthToken(); hasTok {
		sr := StrategyResult{Name: "auth_file"}
		res, fetched, authFailed := tryFetchVercelLogs(ctx, spec, tok, project, "auth_file")
		if fetched {
			sr.Result = "success"
			res.StrategiesTried = append(tried, sr)
			return res, nil
		}
		if authFailed {
			sr.Result = "token_expired"
		} else {
			sr.Result = "not_found"
		}
		tried = append(tried, sr)
	} else {
		tried = append(tried, StrategyResult{Name: "auth_file", Result: "not_found"})
	}

	// 2. env_token
	envSR := StrategyResult{Name: "env_token"}
	if tok := tokenFromEnv(spec); tok != "" {
		res, fetched, authFailed := tryFetchVercelLogs(ctx, spec, tok, project, "env_token")
		if fetched {
			envSR.Result = "success"
			res.StrategiesTried = append(tried, envSR)
			return res, nil
		}
		if authFailed {
			envSR.Result = "token_expired"
		} else {
			envSR.Result = "not_found"
		}
	} else {
		envSR.Result = "not_set"
	}
	tried = append(tried, envSR)

	// 3. credentials_store
	storeSR := StrategyResult{Name: "credentials_store"}
	if tok, hasStore := tokenFromCredentialsStore(spec); hasStore {
		res, fetched, authFailed := tryFetchVercelLogs(ctx, spec, tok, project, "credentials_store")
		if fetched {
			storeSR.Result = "success"
			res.StrategiesTried = append(tried, storeSR)
			return res, nil
		}
		if authFailed {
			storeSR.Result = "token_expired"
		} else {
			storeSR.Result = "not_found"
		}
	} else {
		storeSR.Result = "not_set"
	}
	tried = append(tried, storeSR)

	// 4. auto_setup (last resort)
	if !autoSetupAttempted {
		setupOK, result := autoSetup(ctx, spec)
		tried = append(tried, StrategyResult{Name: "auto_setup", Result: result})
		if setupOK {
			inner, err := resolveVercel(ctx, nodeName, n, reg, true)
			inner.StrategiesTried = append(tried, inner.StrategiesTried...)
			return inner, err
		}
	}

	// Fallback: setup_hint
	return LogResult{
		Source:          "none",
		Provider:        "vercel",
		SetupHint:       buildSetupHint(spec),
		StrategiesTried: tried,
	}, nil
}

func tryFetchVercelLogs(ctx context.Context, spec *provider.Spec, token, project, source string) (LogResult, bool, bool) {
	res, ok, authFailed := fetchVercelLogs(ctx, spec, token, project, source)
	return res, ok, authFailed
}

func fetchVercelLogs(ctx context.Context, spec *provider.Spec, token, project, source string) (LogResult, bool, bool) {
	deploymentID, err := vercelLatestDeploymentID(ctx, spec, token, project)
	if isVercelAuthError(err) {
		logAuthStrategyFailed(source)
		return LogResult{Provider: "vercel", Source: source}, false, true
	}
	if err != nil || deploymentID == "" {
		return LogResult{}, false, false
	}
	lines, err := vercelDeploymentEvents(ctx, spec, token, deploymentID)
	if isVercelAuthError(err) {
		logAuthStrategyFailed(source)
		return LogResult{Provider: "vercel", Source: source}, false, true
	}
	if err != nil || len(lines) == 0 {
		return LogResult{}, false, false
	}
	lines, truncated := truncateLines(lines)
	return LogResult{
		Lines:     lines,
		Source:    source,
		Provider:  "vercel",
		Truncated: truncated,
	}, true, false
}

func vercelLatestDeploymentID(ctx context.Context, spec *provider.Spec, token, project string) (string, error) {
	q := url.Values{}
	q.Set("projectId", project)
	q.Set("limit", "1")
	body, err := vercelGET(ctx, spec, token, "/v6/deployments", q)
	if err != nil {
		return "", err
	}
	var resp struct {
		Deployments []struct {
			UID string `json:"uid"`
			ID  string `json:"id"`
		} `json:"deployments"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", err
	}
	if len(resp.Deployments) == 0 {
		return "", fmt.Errorf("stacklogs: no deployments")
	}
	d := resp.Deployments[0]
	if d.UID != "" {
		return d.UID, nil
	}
	return d.ID, nil
}

func vercelDeploymentEvents(ctx context.Context, spec *provider.Spec, token, deploymentID string) ([]string, error) {
	rawEp := spec.API.Endpoints["logs"]
	path, err := providerGETPath(rawEp)
	if err != nil {
		return nil, err
	}
	vars := map[string]string{
		"token":         token,
		"deployment_id": deploymentID,
		"deploymentId":  deploymentID,
	}
	path = provider.SubstitutePlaceholders(path, vars)
	q := url.Values{}
	q.Set("limit", "200")
	body, err := vercelGET(ctx, spec, token, path, q)
	if err != nil {
		return nil, err
	}
	return parseVercelEvents(body)
}

func parseVercelEvents(body []byte) ([]string, error) {
	body = bytes.TrimSpace(body)
	if len(body) == 0 {
		return nil, nil
	}
	// Live/historical responses are newline-delimited JSON (application/stream+json).
	if body[0] != '[' {
		return scanVercelEventsNDJSON(body)
	}
	var events []map[string]any
	if err := json.Unmarshal(body, &events); err != nil {
		if lines, err := scanVercelEventsNDJSON(body); err == nil && len(lines) > 0 {
			return lines, nil
		}
		return nil, err
	}
	return linesFromVercelEvents(events), nil
}

func scanVercelEventsNDJSON(body []byte) ([]string, error) {
	var out []string
	sc := bufio.NewScanner(bytes.NewReader(body))
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := bytes.TrimSpace(sc.Bytes())
		if len(line) == 0 {
			continue
		}
		var ev map[string]any
		if err := json.Unmarshal(line, &ev); err != nil {
			continue
		}
		out = append(out, linesFromVercelEvent(ev)...)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("stacklogs: no vercel events in stream")
	}
	return out, nil
}

func linesFromVercelEvents(events []map[string]any) []string {
	var lines []string
	for _, ev := range events {
		lines = append(lines, linesFromVercelEvent(ev)...)
	}
	return lines
}

func linesFromVercelEvent(ev map[string]any) []string {
	if ev == nil {
		return nil
	}
	var lines []string
	typ, _ := ev["type"].(string)
	switch typ {
	case "stdout", "stderr", "command", "fatal":
		if text, _ := ev["text"].(string); strings.TrimSpace(text) != "" {
			lines = append(lines, text)
		}
	}
	if payload, ok := ev["payload"].(map[string]any); ok {
		if text, _ := payload["text"].(string); strings.TrimSpace(text) != "" {
			lines = append(lines, text)
		}
	}
	return lines
}

func vercelGET(ctx context.Context, spec *provider.Spec, token, path string, query url.Values) ([]byte, error) {
	baseStr := strings.TrimSuffix(spec.API.BaseURL, "/")
	full, err := joinAPIPath(baseStr, path)
	if err != nil {
		return nil, err
	}
	if len(query) > 0 {
		full.RawQuery = query.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, full.String(), nil)
	if err != nil {
		return nil, err
	}
	if err := provider.ApplyAuthHeader(req, spec.API.AuthHeader, map[string]string{"token": token}); err != nil {
		return nil, err
	}
	resp, err := platformHooks.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return nil, &vercelAuthError{status: resp.Status}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("stacklogs: vercel http %s", resp.Status)
	}
	return body, nil
}

type vercelAuthError struct {
	status string
}

func (e *vercelAuthError) Error() string {
	return fmt.Sprintf("stacklogs: vercel auth %s", e.status)
}

func isVercelAuthError(err error) bool {
	var authErr *vercelAuthError
	return errors.As(err, &authErr)
}

func logAuthStrategyFailed(source string) {
	switch source {
	case "auth_file":
		log.Printf("stacklogs: auth_file token invalid or expired, trying next strategy")
	case "env_token", "credentials_store":
		log.Printf("stacklogs: %s token invalid or expired", source)
	}
}

func providerGETPath(endpoint string) (string, error) {
	endpoint = strings.TrimSpace(endpoint)
	method, rest, ok := strings.Cut(endpoint, " ")
	if !ok || !strings.EqualFold(method, "GET") {
		return "", fmt.Errorf("stacklogs: endpoint must be GET")
	}
	path := strings.TrimSpace(rest)
	if path == "" {
		return "", fmt.Errorf("stacklogs: empty GET path")
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return path, nil
}

func joinAPIPath(baseURL, path string) (*url.URL, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	rel, err := url.Parse(path)
	if err != nil {
		return nil, err
	}
	// ResolveReference treats "/segment" as replacing the entire base path
	// (e.g. https://api.render.com/v1 + /services/x → https://api.render.com/services/x).
	// When base_url includes a version prefix, append the endpoint path instead.
	// Same fix as provider.joinBaseURL.
	if strings.HasPrefix(path, "/") && base.Path != "" && base.Path != "/" {
		merged := strings.TrimSuffix(base.Path, "/") + rel.Path
		return &url.URL{
			Scheme:   base.Scheme,
			Host:     base.Host,
			Path:     merged,
			RawQuery: rel.RawQuery,
		}, nil
	}
	return base.ResolveReference(rel), nil
}
