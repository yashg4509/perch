package stacklogs

import (
	"encoding/json"
	"strings"

	"github.com/yashg4509/perch/internal/provider"
)

// AuthFileToken returns a token from the provider CLI auth file when supported.
func AuthFileToken(spec *provider.Spec) (string, bool) {
	if spec == nil {
		return "", false
	}
	switch spec.Name {
	case "vercel":
		return readVercelAuthToken()
	case "render":
		return readRenderAuthToken()
	case "supabase":
		return readSupabaseAuthToken()
	default:
		return "", false
	}
}

func readVercelAuthToken() (string, bool) {
	path := vercelAuthFilePath()
	if path == "" {
		return "", false
	}
	b, err := platformHooks.readFile(path)
	if err != nil {
		return "", false
	}
	var cfg struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(b, &cfg); err != nil {
		return "", false
	}
	tok := strings.TrimSpace(cfg.Token)
	return tok, tok != ""
}
