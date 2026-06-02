package credentials_test

import (
	"testing"

	"github.com/yashg4509/perch/internal/credentials"
	"github.com/yashg4509/perch/internal/provider"
)

func TestParseDotenv_basic(t *testing.T) {
	raw := `# comment
OPENAI_API_KEY=sk-test
export CLERK_SECRET_KEY="sk_clerk"
EMPTY=
`
	m, err := credentials.ParseDotenv(raw)
	if err != nil {
		t.Fatal(err)
	}
	if m["OPENAI_API_KEY"] != "sk-test" {
		t.Fatalf("openai: %q", m["OPENAI_API_KEY"])
	}
	if m["CLERK_SECRET_KEY"] != "sk_clerk" {
		t.Fatalf("clerk: %q", m["CLERK_SECRET_KEY"])
	}
	if _, ok := m["EMPTY"]; ok {
		t.Fatal("empty values should be omitted")
	}
}

func TestResolveFromEnv_directAlias(t *testing.T) {
	specs := []provider.CredentialsSpec{
		{Key: "openai_api_key", EnvAliases: []string{"OPENAI_API_KEY"}},
		{Key: "clerk_secret_key", EnvAliases: []string{"CLERK_SECRET_KEY"}},
	}
	env := map[string]string{
		"OPENAI_API_KEY":   "sk-o",
		"CLERK_SECRET_KEY": "sk-c",
	}
	got := credentials.ResolveFromEnv(env, specs)
	if got["openai_api_key"] != "sk-o" {
		t.Fatalf("openai_api_key=%q", got["openai_api_key"])
	}
	if got["clerk_secret_key"] != "sk-c" {
		t.Fatalf("clerk_secret_key=%q", got["clerk_secret_key"])
	}
}

func TestResolveFromEnv_compositeCloudinary(t *testing.T) {
	env := map[string]string{
		"CLOUDINARY_API_KEY":    "123",
		"CLOUDINARY_API_SECRET": "sec",
	}
	got := credentials.ResolveFromEnv(env, nil)
	basic, ok := got["cloudinary_api_basic"]
	if !ok || basic == "" {
		t.Fatalf("cloudinary_api_basic missing: %v", got)
	}
	// base64("123:sec")
	if basic != "MTIzOnNlYw==" {
		t.Fatalf("cloudinary_api_basic=%q", basic)
	}
}

func TestResolveFromEnv_skipsUnset(t *testing.T) {
	specs := []provider.CredentialsSpec{
		{Key: "neon_api_key", EnvAliases: []string{"NEON_API_KEY"}},
	}
	got := credentials.ResolveFromEnv(map[string]string{}, specs)
	if len(got) != 0 {
		t.Fatalf("expected empty, got %v", got)
	}
}
