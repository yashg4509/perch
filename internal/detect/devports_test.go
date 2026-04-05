package detect

import (
	"testing"
)

func TestParsePortListeners_table(t *testing.T) {
	text := `
TCP *:3000 (LISTEN) node
TCP *:4000 (LISTEN) node
TCP *:99999 (LISTEN) other
`
	ls, err := ParsePortListeners(text)
	if err != nil {
		t.Fatal(err)
	}
	if len(ls) != 3 {
		t.Fatal(ls)
	}
}

func TestDevServiceHeuristic(t *testing.T) {
	cases := map[int]string{
		3000:  "next-frontend",
		4000:  "api",
		8080:  "api",
		5432:  "postgres",
		6379:  "redis",
		8000:  "python",
		54321: "supabase-local",
	}
	for port, want := range cases {
		if got, ok := DevServiceHeuristic(port); !ok || got != want {
			t.Fatalf("port %d: got %q ok=%v", port, got, ok)
		}
	}
	if _, ok := DevServiceHeuristic(1111); ok {
		t.Fatal("expected unknown port")
	}
}
