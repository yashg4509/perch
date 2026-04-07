package detect

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

// npmPackageToProvider maps package.json dependency keys to perch provider ids (spec: Init table).
var npmPackageToProvider = map[string]string{
	"@supabase/supabase-js":       "supabase",
	"@clerk/nextjs":               "clerk",
	"@auth0/auth0-react":          "auth0",
	"firebase":                    "firebase",
	"stripe":                      "stripe",
	"resend":                      "resend",
	"@sendgrid/mail":              "sendgrid",
	"twilio":                      "twilio",
	"@upstash/redis":              "upstash",
	"@upstash/qstash":             "upstash-qstash",
	"ioredis":                     "redis",
	"redis":                       "redis",
	"mongoose":                    "mongodb",
	"@neondatabase/serverless":    "neon",
	"@planetscale/database":       "planetscale",
	"pg":                          "postgres",
	"postgres":                    "postgres",
	"@aws-sdk/client-s3":          "aws-s3",
	"cloudinary":                  "cloudinary",
	"pusher":                      "pusher",
	"@trigger.dev/sdk":            "trigger",
	"inngest":                     "inngest",
	"openai":                      "openai",
	"@langchain/openai":           "openai",
	"@anthropic-ai/sdk":           "anthropic",
	"langchain":                   "langsmith",
	"@pinecone-database/pinecone": "pinecone",
	"@sentry/node":                "sentry",
	"posthog-js":                  "posthog",
	"@datadog/datadog-api-client": "datadog",
	"@logtail/node":               "logtail",
}

type packageLockJSON struct {
	Dependencies    map[string]any `json:"dependencies"`
	DevDependencies map[string]any `json:"devDependencies"`
}

// PackageJSONSignals returns provider hints from a package.json file path.
func PackageJSONSignals(packageJSONPath string) ([]Signal, error) {
	raw, err := os.ReadFile(packageJSONPath)
	if err != nil {
		return nil, err
	}
	var pj packageLockJSON
	if err := json.Unmarshal(raw, &pj); err != nil {
		return nil, fmt.Errorf("detect: package.json: %w", err)
	}
	seen := map[string]struct{}{}
	var provs []string
	add := func(pkg string) {
		p, ok := npmPackageToProvider[pkg]
		if !ok {
			return
		}
		if _, dup := seen[p]; dup {
			return
		}
		seen[p] = struct{}{}
		provs = append(provs, p)
	}
	for pkg := range pj.Dependencies {
		add(pkg)
	}
	for pkg := range pj.DevDependencies {
		add(pkg)
	}
	sort.Strings(provs)
	if len(provs) == 0 {
		return nil, nil
	}
	return []Signal{{RelPath: "package.json", Providers: provs}}, nil
}

// NPMDependencyKeys returns sorted dependency keys (dependencies + devDependencies) for edge inference.
func NPMDependencyKeys(packageJSONPath string) ([]string, error) {
	raw, err := os.ReadFile(packageJSONPath)
	if err != nil {
		return nil, err
	}
	var pj packageLockJSON
	if err := json.Unmarshal(raw, &pj); err != nil {
		return nil, fmt.Errorf("detect: package.json: %w", err)
	}
	seen := map[string]struct{}{}
	for pkg := range pj.Dependencies {
		seen[pkg] = struct{}{}
	}
	for pkg := range pj.DevDependencies {
		seen[pkg] = struct{}{}
	}
	var keys []string
	for k := range seen {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys, nil
}
