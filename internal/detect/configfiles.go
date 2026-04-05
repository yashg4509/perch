package detect

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Signal is one detection hit from the working tree (committed files on disk during init).
type Signal struct {
	RelPath   string
	Providers []string
}

var prismaProviderRe = regexp.MustCompile(`provider\s*=\s*"([^"]+)"`)

// ConfigFileSignals returns provider hints from well-known config filenames under root (non-.env scan).
func ConfigFileSignals(root string) ([]Signal, error) {
	root = filepath.Clean(root)
	var out []Signal
	add := func(rel string, provs []string) {
		out = append(out, Signal{RelPath: rel, Providers: append([]string(nil), provs...)})
	}

	checks := []struct {
		rel   string
		provs []string
	}{
		{"vercel.json", []string{"vercel"}},
		{"netlify.toml", []string{"netlify"}},
		{"fly.toml", []string{"fly"}},
		{"railway.toml", []string{"railway"}},
		{"railway.json", []string{"railway"}},
		{"wrangler.toml", []string{"cloudflare"}},
		{"render.yaml", []string{"render"}},
		{"firebase.json", []string{"firebase"}},
	}
	for _, c := range checks {
		if existsFile(filepath.Join(root, c.rel)) {
			add(c.rel, c.provs)
		}
	}

	prismaPath := filepath.Join(root, "prisma", "schema.prisma")
	if existsFile(prismaPath) {
		data, err := os.ReadFile(prismaPath)
		if err != nil {
			return nil, err
		}
		if p, ok := prismaMappedProvider(string(data)); ok {
			add("prisma/schema.prisma", []string{p})
		}
	}

	for _, name := range []string{"docker-compose.yml", "docker-compose.yaml"} {
		if existsFile(filepath.Join(root, name)) {
			add(name, []string{"custom"})
			break
		}
	}

	return out, nil
}

func existsFile(p string) bool {
	st, err := os.Stat(p)
	return err == nil && !st.IsDir()
}

func prismaMappedProvider(schema string) (string, bool) {
	m := prismaProviderRe.FindStringSubmatch(schema)
	if m == nil {
		return "", false
	}
	raw := strings.ToLower(strings.TrimSpace(m[1]))
	switch raw {
	case "postgresql", "postgres":
		return "postgres", true
	case "mysql":
		return "mysql", true
	case "mongodb":
		return "mongodb", true
	default:
		return "", false
	}
}
