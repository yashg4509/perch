package snapshot

import (
	"path/filepath"
	"testing"
	"time"
)

func TestStore_saveListRoundTrip(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(filepath.Join(dir, "snapshots.json"))

	rec := Record{
		Name:      "pre-release",
		CreatedAt: time.Date(2026, 4, 4, 12, 0, 0, 0, time.UTC),
		SHAs: map[string]string{
			"frontend": "aaa111",
			"backend":  "bbb222",
		},
	}
	if err := s.Save(rec); err != nil {
		t.Fatal(err)
	}
	list, err := s.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("%+v", list)
	}
	if list[0].Name != "pre-release" || list[0].SHAs["backend"] != "bbb222" {
		t.Fatalf("%+v", list[0])
	}
}

func TestStore_appendSecond(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(filepath.Join(dir, "s.json"))
	_ = s.Save(Record{Name: "a", SHAs: map[string]string{"n": "1"}})
	_ = s.Save(Record{Name: "b", SHAs: map[string]string{"n": "2"}})
	list, err := s.List()
	if err != nil || len(list) != 2 {
		t.Fatalf("%v %+v", err, list)
	}
}
