package provider_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yashg4509/perch/internal/provider"
)

func TestPusherSignedGET(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.URL.Query().Get("auth_key")
		if r.URL.Query().Get("auth_signature") == "" {
			t.Fatal("missing signature")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	code, err := provider.PusherSignedGET(context.Background(), srv.Client(), srv.URL, "/apps/1/channels", "app-key", "app-secret")
	if err != nil {
		t.Fatal(err)
	}
	if code != http.StatusOK {
		t.Fatalf("code=%d", code)
	}
	if gotAuth != "app-key" {
		t.Fatalf("auth_key=%q", gotAuth)
	}
}
