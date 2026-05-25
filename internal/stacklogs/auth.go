package stacklogs

import (
	"encoding/json"
	"strings"

	"gopkg.in/yaml.v3"
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

func readGitHubTokenFromHosts() (string, bool) {
	path := ghHostsPath()
	if path == "" {
		return "", false
	}
	b, err := platformHooks.readFile(path)
	if err != nil {
		return "", false
	}
	var root map[string]any
	if err := yaml.Unmarshal(b, &root); err != nil {
		return "", false
	}
	host, _ := root["github.com"].(map[string]any)
	if host == nil {
		for k, v := range root {
			if strings.Contains(k, "github") {
				host, _ = v.(map[string]any)
				break
			}
		}
	}
	if host == nil {
		return "", false
	}
	tok, _ := host["oauth_token"].(string)
	tok = strings.TrimSpace(tok)
	return tok, tok != ""
}

func githubTokenFromEnvOrHosts() (string, bool) {
	for _, k := range []string{"GITHUB_TOKEN", "GH_TOKEN"} {
		if t := strings.TrimSpace(platformHooks.getenv(k)); t != "" {
			return t, true
		}
	}
	return readGitHubTokenFromHosts()
}
