package provider

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestJoinBaseURL_rejectsSchemeRelative(t *testing.T) {
	_, err := joinBaseURL("https://api.example.com", "//evil.example/hook")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestJoinBaseURL_rejectsAbsolutePathWithHost(t *testing.T) {
	_, err := joinBaseURL("https://api.example.com", "https://evil.example/x")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestJoinBaseURL_ok(t *testing.T) {
	u, err := joinBaseURL("https://api.example.com", "/v1/foo")
	if err != nil {
		t.Fatal(err)
	}
	if u.String() != "https://api.example.com/v1/foo" {
		t.Fatalf("got %s", u.String())
	}
}

func TestJoinBaseURL_appendsToVersionedBase(t *testing.T) {
	tests := []struct {
		base, path, want string
	}{
		{"https://api.anthropic.com/v1", "/models", "https://api.anthropic.com/v1/models"},
		{"https://api.clerk.com/v1", "/users?limit=1", "https://api.clerk.com/v1/users?limit=1"},
	}
	for _, tc := range tests {
		u, err := joinBaseURL(tc.base, tc.path)
		if err != nil {
			t.Fatalf("%s + %s: %v", tc.base, tc.path, err)
		}
		if u.String() != tc.want {
			t.Fatalf("%s + %s: got %s want %s", tc.base, tc.path, u.String(), tc.want)
		}
	}
}

func TestSameHostRedirect_blocksCrossHost(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://other.example/", http.StatusFound)
	}))
	t.Cleanup(srv.Close)

	req, err := http.NewRequest(http.MethodGet, srv.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	client := &http.Client{CheckRedirect: sameHostRedirect}
	_, err = client.Do(req)
	if err == nil {
		t.Fatal("expected redirect error")
	}
}
