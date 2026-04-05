package credentials

import (
	"path/filepath"
	"testing"
)

func TestStore_roundTrip(t *testing.T) {
	home := t.TempDir()
	p := filepath.Join(home, ".perch", "credentials")
	s := NewStore(p)
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
	if err := NewStore(p).Set("k", "v"); err != nil {
		t.Fatal(err)
	}
	v, ok, err := NewStore(p).Get("k")
	if err != nil || !ok || v != "v" {
		t.Fatalf("%q ok=%v err=%v", v, ok, err)
	}
}
