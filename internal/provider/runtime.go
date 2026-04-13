package provider

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// DoGETJSON performs a GET using spec.API against the named endpoint key, after substituting vars
// into the endpoint path template and auth header. Only "GET ..." endpoint forms are supported here.
func DoGETJSON(ctx context.Context, client *http.Client, spec *Spec, endpointKey string, vars map[string]string, dest any) error {
	if client == nil {
		client = HTTPClientForAPI()
	} else {
		client = withSameHostRedirects(client)
	}
	rawEp, ok := spec.API.Endpoints[endpointKey]
	if !ok {
		return fmt.Errorf("provider: unknown endpoint %q", endpointKey)
	}
	path, err := parseGETPath(rawEp)
	if err != nil {
		return err
	}
	path = SubstitutePlaceholders(path, vars)
	baseStr := strings.TrimSuffix(spec.API.BaseURL, "/")
	full, err := joinBaseURL(baseStr, path)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, full.String(), nil)
	if err != nil {
		return fmt.Errorf("provider: request: %w", err)
	}
	if err := ApplyAuthHeader(req, spec.API.AuthHeader, vars); err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("provider: http: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("provider: read body: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("provider: http %s: %s", resp.Status, truncate(body, 200))
	}
	return DecodeJSON(body, dest)
}

func parseGETPath(endpoint string) (string, error) {
	endpoint = strings.TrimSpace(endpoint)
	method, rest, ok := strings.Cut(endpoint, " ")
	if !ok {
		return "", fmt.Errorf("provider: endpoint %q must start with HTTP method", endpoint)
	}
	if !strings.EqualFold(method, "GET") {
		return "", fmt.Errorf("provider: only GET endpoints supported in runtime helper, got %q", method)
	}
	path := strings.TrimSpace(rest)
	if path == "" {
		return "", fmt.Errorf("provider: empty GET path")
	}
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		u, err := url.Parse(path)
		if err != nil {
			return "", err
		}
		return u.RequestURI(), nil
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return path, nil
}

func truncate(b []byte, n int) string {
	s := string(b)
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
