package provider

import (
	"net/http"
	"testing"
)

func TestSubstitutePlaceholders(t *testing.T) {
	got := SubstitutePlaceholders("GET /v9/projects/{project}/env", map[string]string{
		"project": "my-app",
	})
	if got != "GET /v9/projects/my-app/env" {
		t.Fatalf("got %q", got)
	}
}

func TestApplyAuthHeader(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "https://example.com/x", nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := ApplyAuthHeader(req, "Authorization: Bearer {token}", map[string]string{"token": "abc"}); err != nil {
		t.Fatal(err)
	}
	if req.Header.Get("Authorization") != "Bearer abc" {
		t.Fatalf("got %q", req.Header.Get("Authorization"))
	}
}
