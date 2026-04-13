package detect

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPackageJSONSignals_table(t *testing.T) {
	cases := []struct {
		json string
		want map[string]struct{}
	}{
		{`{"dependencies":{"@supabase/supabase-js":"^2"}}`, map[string]struct{}{"supabase": {}}},
		{`{"devDependencies":{"stripe":"^14"}}`, map[string]struct{}{"stripe": {}}},
		{`{"dependencies":{"openai":"^4","pg":"^8"}}`, map[string]struct{}{"openai": {}, "postgres": {}}},
		{`{"dependencies":{"@langchain/openai":"0.1"}}`, map[string]struct{}{"openai": {}}},
		{`{"dependencies":{"@anthropic-ai/sdk":"^0.10"}}`, map[string]struct{}{"anthropic": {}}},
		{`{"dependencies":{"@pinecone-database/pinecone":"^2"}}`, map[string]struct{}{"pinecone": {}}},
		{`{"dependencies":{"@clerk/nextjs":"^5"}}`, map[string]struct{}{"clerk": {}}},
		{`{"dependencies":{"langchain":"^0.2"}}`, map[string]struct{}{"langsmith": {}}},
	}
	for i, tc := range cases {
		root := t.TempDir()
		p := filepath.Join(root, "package.json")
		if err := os.WriteFile(p, []byte(tc.json), 0o644); err != nil {
			t.Fatal(err)
		}
		sigs, err := PackageJSONSignals(filepath.Join(root, "package.json"))
		if err != nil {
			t.Fatalf("case %d: %v", i, err)
		}
		got := map[string]struct{}{}
		for _, s := range sigs {
			for _, pr := range s.Providers {
				got[pr] = struct{}{}
			}
		}
		if len(got) != len(tc.want) {
			t.Fatalf("case %d: got %v want %v", i, got, tc.want)
		}
		for w := range tc.want {
			if _, ok := got[w]; !ok {
				t.Fatalf("case %d: missing %q", i, w)
			}
		}
	}
}

func TestPackageJSONSignals_missingFile(t *testing.T) {
	_, err := PackageJSONSignals(filepath.Join(t.TempDir(), "package.json"))
	if err == nil {
		t.Fatal("expected error")
	}
}
