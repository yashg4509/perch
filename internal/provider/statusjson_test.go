package provider

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func statusFixture(t *testing.T, name string) []byte {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	p := filepath.Join(filepath.Dir(file), "testdata", "status", name)
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestParseStatusJSON_renderFixture(t *testing.T) {
	st, err := ParseStatusJSON(statusFixture(t, "render_backend.json"))
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
	st, err := ParseStatusJSON(statusFixture(t, "openai_llm.json"))
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
