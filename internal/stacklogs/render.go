package stacklogs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/yashg4509/perch/internal/config"
	"github.com/yashg4509/perch/internal/provider"
	"gopkg.in/yaml.v3"
)

// TODO: v2 pagination via hasMore and nextStartTime on Render logs API.

func resolveRenderLogs(ctx context.Context, nodeName string, n config.Node, reg *provider.Registry, autoSetupAttempted bool) (LogResult, error) {
	service := strings.TrimSpace(n.Service)
	if service == "" {
		return LogResult{
			Source:    "none",
			Provider:  "render",
			SetupHint: fmt.Sprintf("node %q needs a service field for Render logs", nodeName),
		}, nil
	}
	spec := reg.ByName["render"]
	if spec == nil {
		return LogResult{}, fmt.Errorf("stacklogs: render provider spec missing from registry")
	}

	var tried []StrategyResult

	// 1. auth_file
	if tok, hasTok := readRenderAuthToken(); hasTok {
		sr := StrategyResult{Name: "auth_file"}
		res, fetched, authFailed := tryFetchRenderLogs(ctx, spec, tok, service, "auth_file")
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
		res, fetched, authFailed := tryFetchRenderLogs(ctx, spec, tok, service, "env_token")
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
		res, fetched, authFailed := tryFetchRenderLogs(ctx, spec, tok, service, "credentials_store")
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
			inner, err := resolveRenderLogs(ctx, nodeName, n, reg, true)
			inner.StrategiesTried = append(tried, inner.StrategiesTried...)
			return inner, err
		}
	}

	return LogResult{
		Source:          "none",
		Provider:        "render",
		SetupHint:       buildSetupHint(spec),
		StrategiesTried: tried,
	}, nil
}

func tryFetchRenderLogs(ctx context.Context, spec *provider.Spec, token, service, source string) (LogResult, bool, bool) {
	res, ok, authFailed := fetchRenderLogs(ctx, spec, token, service, source)
	return res, ok, authFailed
}

func fetchRenderLogs(ctx context.Context, spec *provider.Spec, token, service, source string) (LogResult, bool, bool) {
	lines, err := renderServiceLogs(ctx, spec, token, service)
	if isRenderAuthError(err) {
		logAuthStrategyFailed(source)
		return LogResult{Provider: "render", Source: source}, false, true
	}
	if err != nil || len(lines) == 0 {
		return LogResult{}, false, false
	}
	lines, truncated := truncateLines(lines)
	return LogResult{
		Lines:     lines,
		Source:    source,
		Provider:  "render",
		Truncated: truncated,
	}, true, false
}

func renderServiceLogs(ctx context.Context, spec *provider.Spec, token, service string) ([]string, error) {
	ownerID, err := renderServiceOwnerID(ctx, spec, token, service)
	if err != nil {
		return nil, err
	}
	q := url.Values{}
	q.Set("ownerId", ownerID)
	q.Add("resource", service)
	q.Set("limit", "100")
	body, err := renderGET(ctx, spec, token, "/logs", q)
	if err != nil {
		return nil, err
	}
	return parseRenderLogs(body)
}

func renderServiceOwnerID(ctx context.Context, spec *provider.Spec, token, service string) (string, error) {
	path := fmt.Sprintf("/services/%s", url.PathEscape(service))
	body, err := renderGET(ctx, spec, token, path, nil)
	if err != nil {
		return "", err
	}
	var svc struct {
		OwnerID string `json:"ownerId"`
	}
	if err := json.Unmarshal(body, &svc); err != nil {
		return "", err
	}
	ownerID := strings.TrimSpace(svc.OwnerID)
	if ownerID == "" {
		return "", fmt.Errorf("stacklogs: render service %q missing ownerId", service)
	}
	return ownerID, nil
}

func parseRenderLogs(body []byte) ([]string, error) {
	var resp struct {
		Logs []struct {
			Message string `json:"message"`
		} `json:"logs"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}
	var lines []string
	for _, entry := range resp.Logs {
		msg := strings.TrimSpace(entry.Message)
		if msg != "" {
			lines = append(lines, msg)
		}
	}
	if len(lines) == 0 {
		return nil, fmt.Errorf("stacklogs: no render log lines")
	}
	return lines, nil
}

func renderGET(ctx context.Context, spec *provider.Spec, token, path string, query url.Values) ([]byte, error) {
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
		return nil, &renderAuthError{status: resp.Status}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("stacklogs: render http %s", resp.Status)
	}
	return body, nil
}

type renderAuthError struct {
	status string
}

func (e *renderAuthError) Error() string {
	return fmt.Sprintf("stacklogs: render auth %s", e.status)
}

func isRenderAuthError(err error) bool {
	var authErr *renderAuthError
	return errors.As(err, &authErr)
}

func renderAuthFilePath() string {
	return homePath(".config", "render", "config.yaml")
}

func readRenderAuthToken() (string, bool) {
	path := renderAuthFilePath()
	if path == "" {
		return "", false
	}
	b, err := platformHooks.readFile(path)
	if err != nil {
		return "", false
	}
	var cfg struct {
		APIKey string `yaml:"api_key"`
	}
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return "", false
	}
	tok := strings.TrimSpace(cfg.APIKey)
	return tok, tok != ""
}
