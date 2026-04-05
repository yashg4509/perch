package detect

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectConfigFiles_table(t *testing.T) {
	cases := []struct {
		name     string
		files    map[string]string   // relPath -> content (empty file ok)
		wantProv map[string]struct{} // provider ids
	}{
		{
			name:     "vercel.json",
			files:    map[string]string{"vercel.json": "{}"},
			wantProv: map[string]struct{}{"vercel": {}},
		},
		{
			name:     "netlify.toml",
			files:    map[string]string{"netlify.toml": ""},
			wantProv: map[string]struct{}{"netlify": {}},
		},
		{
			name:     "fly.toml",
			files:    map[string]string{"fly.toml": ""},
			wantProv: map[string]struct{}{"fly": {}},
		},
		{
			name:     "railway.toml",
			files:    map[string]string{"railway.toml": ""},
			wantProv: map[string]struct{}{"railway": {}},
		},
		{
			name:     "railway.json",
			files:    map[string]string{"railway.json": "{}"},
			wantProv: map[string]struct{}{"railway": {}},
		},
		{
			name:     "wrangler.toml",
			files:    map[string]string{"wrangler.toml": ""},
			wantProv: map[string]struct{}{"cloudflare": {}},
		},
		{
			name:     "render.yaml",
			files:    map[string]string{"render.yaml": ""},
			wantProv: map[string]struct{}{"render": {}},
		},
		{
			name:     "firebase.json",
			files:    map[string]string{"firebase.json": "{}"},
			wantProv: map[string]struct{}{"firebase": {}},
		},
		{
			name: "prisma_postgres",
			files: map[string]string{
				"prisma/schema.prisma": `datasource db {
  provider = "postgresql"
  url      = env("DATABASE_URL")
}`,
			},
			wantProv: map[string]struct{}{"postgres": {}},
		},
		{
			name: "prisma_mysql",
			files: map[string]string{
				"prisma/schema.prisma": `datasource db {
  provider = "mysql"
}`,
			},
			wantProv: map[string]struct{}{"mysql": {}},
		},
		{
			name: "prisma_mongodb",
			files: map[string]string{
				"prisma/schema.prisma": `datasource db {
  provider = "mongodb"
}`,
			},
			wantProv: map[string]struct{}{"mongodb": {}},
		},
		{
			name:     "docker-compose.yml",
			files:    map[string]string{"docker-compose.yml": "services:\n  api:\n"},
			wantProv: map[string]struct{}{"custom": {}},
		},
		{
			name:     "docker-compose.yaml",
			files:    map[string]string{"docker-compose.yaml": "version: '3'\n"},
			wantProv: map[string]struct{}{"custom": {}},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			for rel, body := range tc.files {
				p := filepath.Join(root, rel)
				if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			sigs, err := ConfigFileSignals(root)
			if err != nil {
				t.Fatal(err)
			}
			got := map[string]struct{}{}
			for _, s := range sigs {
				for _, p := range s.Providers {
					got[p] = struct{}{}
				}
			}
			if len(got) != len(tc.wantProv) {
				t.Fatalf("got %v want %v", got, tc.wantProv)
			}
			for p := range tc.wantProv {
				if _, ok := got[p]; !ok {
					t.Fatalf("missing provider %q, got %v", p, got)
				}
			}
		})
	}
}

func TestDetectConfigFiles_emptyRoot(t *testing.T) {
	sigs, err := ConfigFileSignals(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if len(sigs) != 0 {
		t.Fatalf("%+v", sigs)
	}
}

func TestDetectConfigFiles_mergeSameScan(t *testing.T) {
	root := t.TempDir()
	_ = os.WriteFile(filepath.Join(root, "vercel.json"), []byte("{}"), 0o644)
	_ = os.WriteFile(filepath.Join(root, "fly.toml"), []byte(""), 0o644)
	sigs, err := ConfigFileSignals(root)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]struct{}{}
	for _, s := range sigs {
		for _, p := range s.Providers {
			got[p] = struct{}{}
		}
	}
	if len(got) != 2 {
		t.Fatalf("%v", got)
	}
}
