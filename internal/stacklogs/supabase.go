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
	"time"

	"github.com/yashg4509/perch/internal/config"
	"github.com/yashg4509/perch/internal/provider"
	"github.com/zalando/go-keyring"
)

// TODO: v2 pagination and iso_timestamp_start/end windowing for Supabase logs API.

func resolveSupabaseLogs(ctx context.Context, nodeName string, n config.Node, reg *provider.Registry, autoSetupAttempted bool) (LogResult, error) {
	project := strings.TrimSpace(n.Project)
	if project == "" {
		return LogResult{
			Source:    "none",
			Provider:  "supabase",
			SetupHint: fmt.Sprintf("node %q needs a project field for Supabase logs", nodeName),
		}, nil
	}
	spec := reg.ByName["supabase"]
	if spec == nil {
		return LogResult{}, fmt.Errorf("stacklogs: supabase provider spec missing from registry")
	}

	var tried []StrategyResult

	// 1. auth_file
	if tok, hasTok := readSupabaseAuthToken(); hasTok {
		sr := StrategyResult{Name: "auth_file"}
		res, fetched, authFailed := tryFetchSupabaseLogs(ctx, spec, tok, project, "auth_file")
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
		res, fetched, authFailed := tryFetchSupabaseLogs(ctx, spec, tok, project, "env_token")
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
		res, fetched, authFailed := tryFetchSupabaseLogs(ctx, spec, tok, project, "credentials_store")
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
			inner, err := resolveSupabaseLogs(ctx, nodeName, n, reg, true)
			inner.StrategiesTried = append(tried, inner.StrategiesTried...)
			return inner, err
		}
	}

	return LogResult{
		Source:          "none",
		Provider:        "supabase",
		SetupHint:       buildSetupHint(spec),
		StrategiesTried: tried,
	}, nil
}

func tryFetchSupabaseLogs(ctx context.Context, spec *provider.Spec, token, project, source string) (LogResult, bool, bool) {
	res, ok, authFailed := fetchSupabaseLogs(ctx, spec, token, project, source)
	return res, ok, authFailed
}

func fetchSupabaseLogs(ctx context.Context, spec *provider.Spec, token, project, source string) (LogResult, bool, bool) {
	lines, err := supabaseProjectLogs(ctx, spec, token, project)
	if isSupabaseAuthError(err) {
		logAuthStrategyFailed(source)
		return LogResult{Provider: "supabase", Source: source}, false, true
	}
	if err != nil || len(lines) == 0 {
		return LogResult{}, false, false
	}
	lines, truncated := truncateLines(lines)
	return LogResult{
		Lines:     lines,
		Source:    source,
		Provider:  "supabase",
		Truncated: truncated,
	}, true, false
}

func supabaseProjectLogs(ctx context.Context, spec *provider.Spec, token, project string) ([]string, error) {
	// Management API analytics logs endpoint (not /projects/{ref}/logs).
	path := fmt.Sprintf("/projects/%s/analytics/endpoints/logs.all", url.PathEscape(project))
	q := url.Values{}
	q.Set("sql", "SELECT event_message FROM postgres_logs LIMIT 100")
	end := time.Now().UTC()
	start := end.Add(-24 * time.Hour)
	q.Set("iso_timestamp_start", start.Format(time.RFC3339))
	q.Set("iso_timestamp_end", end.Format(time.RFC3339))
	body, err := supabaseGET(ctx, spec, token, path, q)
	if err != nil {
		return nil, err
	}
	return parseSupabaseLogs(body)
}

func parseSupabaseLogs(body []byte) ([]string, error) {
	var resp struct {
		Result []json.RawMessage `json:"result"`
		Error  string            `json:"error"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}
	if errMsg := strings.TrimSpace(resp.Error); errMsg != "" {
		return nil, fmt.Errorf("stacklogs: supabase logs error: %s", errMsg)
	}
	var lines []string
	for _, raw := range resp.Result {
		if len(raw) == 0 || string(raw) == "null" {
			continue
		}
		for _, line := range linesFromSupabaseLogRow(raw) {
			if line != "" {
				lines = append(lines, line)
			}
		}
	}
	if len(lines) == 0 {
		return nil, fmt.Errorf("stacklogs: no supabase log lines")
	}
	return lines, nil
}

func linesFromSupabaseLogRow(raw json.RawMessage) []string {
	var row map[string]any
	if err := json.Unmarshal(raw, &row); err != nil {
		return []string{strings.TrimSpace(string(raw))}
	}
	for _, key := range []string{"event_message", "message", "msg"} {
		if v, ok := row[key].(string); ok {
			if s := strings.TrimSpace(v); s != "" {
				return []string{s}
			}
		}
	}
	b, err := json.Marshal(row)
	if err != nil {
		return nil
	}
	return []string{string(b)}
}

func supabaseGET(ctx context.Context, spec *provider.Spec, token, path string, query url.Values) ([]byte, error) {
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
		return nil, &supabaseAuthError{status: resp.Status}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("stacklogs: supabase http %s", resp.Status)
	}
	return body, nil
}

type supabaseAuthError struct {
	status string
}

func (e *supabaseAuthError) Error() string {
	return fmt.Sprintf("stacklogs: supabase auth %s", e.status)
}

func isSupabaseAuthError(err error) bool {
	var authErr *supabaseAuthError
	return errors.As(err, &authErr)
}

func supabaseAuthFilePath() string {
	return homePath(".supabase", "access-token")
}

// supabaseKeyringService matches the Supabase CLI credentials store namespace
// (zalando/go-keyring → macOS Keychain / Linux Secret Service / Windows Credential Manager).
const supabaseKeyringService = "Supabase CLI"

// supabaseKeyringGet is swapped in tests so CI/dev keychains do not leak into unit tests.
var supabaseKeyringGet = func(service, account string) (string, error) {
	return keyring.Get(service, account)
}

func readSupabaseAuthToken() (string, bool) {
	// Prefer the legacy plain-text file when present (tests + older CLI installs).
	if tok, ok := readSupabaseAuthTokenFile(); ok {
		return tok, true
	}
	if tok, ok := readSupabaseAuthTokenKeyring(); ok {
		return tok, true
	}
	return "", false
}

func readSupabaseAuthTokenFile() (string, bool) {
	path := supabaseAuthFilePath()
	if path == "" {
		return "", false
	}
	b, err := platformHooks.readFile(path)
	if err != nil {
		return "", false
	}
	tok := strings.TrimSpace(string(b))
	return tok, tok != ""
}

func readSupabaseAuthTokenKeyring() (string, bool) {
	for _, account := range supabaseKeyringAccounts() {
		tok, err := supabaseKeyringGet(supabaseKeyringService, account)
		if err != nil {
			continue
		}
		tok = strings.TrimSpace(tok)
		if tok != "" {
			return tok, true
		}
	}
	return "", false
}

func supabaseKeyringAccounts() []string {
	// Default profile name used by the Supabase CLI; profile file may override.
	accounts := []string{"supabase"}
	if b, err := platformHooks.readFile(homePath(".supabase", "profile")); err == nil {
		if name := strings.TrimSpace(string(b)); name != "" && name != "supabase" {
			accounts = append([]string{name}, accounts...)
		}
	}
	return accounts
}
