package provider

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// joinBaseURL resolves path against baseURL and enforces that the result stays on the same
// origin as base (scheme + host). This blocks scheme-relative URLs (//host/...) that
// [url.URL.ResolveReference] would otherwise treat as a new authority.
func joinBaseURL(baseURL, path string) (*url.URL, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("provider: base_url: %w", err)
	}
	if base.Scheme == "" || base.Host == "" {
		return nil, fmt.Errorf("provider: base_url must include scheme and host")
	}
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, fmt.Errorf("provider: empty path")
	}
	if strings.HasPrefix(path, "//") {
		return nil, fmt.Errorf("provider: path must not start with //")
	}
	rel, err := url.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("provider: path: %w", err)
	}
	if rel.Scheme != "" || rel.Host != "" {
		return nil, fmt.Errorf("provider: path must be host-relative only")
	}
	full := base.ResolveReference(rel)
	if !sameOrigin(full, base) {
		return nil, fmt.Errorf("provider: resolved URL left configured origin")
	}
	return full, nil
}

func sameOrigin(a, b *url.URL) bool {
	return strings.EqualFold(a.Scheme, b.Scheme) && strings.EqualFold(a.Host, b.Host)
}

// sameHostRedirect rejects redirects that change scheme or host (mitigates open redirects /
// SSRF via 3xx from the platform API).
func sameHostRedirect(req *http.Request, via []*http.Request) error {
	if len(via) >= 10 {
		return fmt.Errorf("provider http: stopped after 10 redirects")
	}
	if len(via) == 0 {
		return nil
	}
	orig := via[0].URL
	if !sameOrigin(req.URL, orig) {
		return fmt.Errorf("provider http: cross-origin redirect blocked")
	}
	return nil
}

// HTTPClientForAPI returns an [http.Client] that blocks cross-host redirects. Pass it to
// [DoGETJSON] instead of [http.DefaultClient] for outbound provider calls.
func HTTPClientForAPI() *http.Client {
	return &http.Client{CheckRedirect: sameHostRedirect}
}

// withSameHostRedirects returns a shallow copy of c with [sameHostRedirect] installed when c
// does not already define [http.Client.CheckRedirect], preserving Transport and Timeout.
func withSameHostRedirects(c *http.Client) *http.Client {
	if c == nil {
		return HTTPClientForAPI()
	}
	if c.CheckRedirect != nil {
		return c
	}
	cp := *c
	cp.CheckRedirect = sameHostRedirect
	return &cp
}
