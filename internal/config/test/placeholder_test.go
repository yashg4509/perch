package config_test

import (
	"testing"

	"github.com/yashg4509/perch/internal/config"
)

func TestIsPlaceholder(t *testing.T) {
	cases := map[string]bool{
		"":                        true,
		"CHANGE_ME":               true,
		"CHANGE_ME_FOO":           true,
		"YOUR_VERCEL_PROJECT":     true,
		"perch-brief":             false,
		"local-dev":               true,
		"local-neon":              true,
		"my-pinecone-index":       false,
	}
	for in, want := range cases {
		if got := config.IsPlaceholder(in); got != want {
			t.Fatalf("IsPlaceholder(%q)=%v want %v", in, got, want)
		}
	}
}
