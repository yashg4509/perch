package cli

import (
	"io/fs"
	"testing"

	webdist "github.com/yashg4509/perch/web"
)

func TestEmbeddedUIDistHasIndexHTML(t *testing.T) {
	t.Parallel()
	sub, err := fs.Sub(webdist.Dist, "dist")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fs.Stat(sub, "index.html"); err != nil {
		t.Fatalf("embedded UI must include index.html under dist/ (run: cd web && npm run build:embed): %v", err)
	}
}
