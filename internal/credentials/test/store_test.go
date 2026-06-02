package credentials_test

import (
	"path/filepath"
	"testing"

	"github.com/yashg4509/perch/internal/credentials"
)

func TestStore_roundTrip(t *testing.T) {
	home := t.TempDir()
	p := filepath.Join(home, ".perch", "credentials")
	s := credentials.NewStoreAt(p)
	if err := s.Set("vercel_token", "secret"); err != nil {
		t.Fatal(err)
	}
	v, ok, err := s.Get("vercel_token")
	if err != nil {
		t.Fatal(err)
	}
	if !ok || v != "secret" {
		t.Fatalf("got %q ok=%v", v, ok)
	}
}

func TestStore_persistsAcrossNewStore(t *testing.T) {
	home := t.TempDir()
	p := filepath.Join(home, ".perch", "credentials")
	if err := credentials.NewStoreAt(p).Set("k", "v"); err != nil {
		t.Fatal(err)
	}
	v, ok, err := credentials.NewStoreAt(p).Get("k")
	if err != nil || !ok || v != "v" {
		t.Fatalf("%q ok=%v err=%v", v, ok, err)
	}
}
