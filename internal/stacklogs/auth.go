package stacklogs

import (
	"encoding/json"
	"strings"
)

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
