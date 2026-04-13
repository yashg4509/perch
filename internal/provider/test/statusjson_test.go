package provider_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yashg4509/perch/internal/provider"
	"github.com/yashg4509/perch/internal/testutil"
)

func statusFixture(t *testing.T, name string) []byte {
	t.Helper()
	root := testutil.RepoRoot(t)
	p := filepath.Join(root, "internal", "provider", "testdata", "status", name)
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestParseStatusJSON_renderFixture(t *testing.T) {
	st, err := provider.ParseStatusJSON(statusFixture(t, "render_backend.json"))
	if err != nil {
		t.Fatal(err)
	}
	if st.Healthy {
		t.Fatal("expected unhealthy")
	}
	if st.ErrorRate == nil || *st.ErrorRate != 0.123 {
		t.Fatalf("error_rate %+v", st.ErrorRate)
	}
	if st.LastDeploy == nil || st.LastDeploy.SHA != "d9e1f3a" || st.LastDeploy.Ago != "28m" {
		t.Fatalf("last_deploy %+v", st.LastDeploy)
	}
	if len(st.RecentErrors) != 1 {
		t.Fatal(st.RecentErrors)
	}
}

func TestParseStatusJSON_openaiFixture(t *testing.T) {
	st, err := provider.ParseStatusJSON(statusFixture(t, "openai_llm.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !st.Healthy {
		t.Fatal(st)
	}
	if st.DailyTokens == nil || *st.DailyTokens != 142000 {
		t.Fatalf("tokens %+v", st.DailyTokens)
	}
	if st.DailyCostUSD == nil || *st.DailyCostUSD != 0.43 {
		t.Fatalf("cost %+v", st.DailyCostUSD)
	}
	if st.ErrorRate == nil || *st.ErrorRate != 0.002 {
		t.Fatalf("error_rate %+v", st.ErrorRate)
	}
}
