package trace_test

import (
	"testing"
	"time"

	"github.com/yashg4509/perch/internal/trace"
)

func TestFilterWindow_includesNeighbors(t *testing.T) {
	center := time.Date(2026, 4, 4, 14, 31, 2, 0, time.UTC)
	lines := []trace.Line{
		{At: center.Add(-5 * time.Second), Text: "before"},
		{At: center, Text: "hit"},
		{At: center.Add(3 * time.Second), Text: "after"},
		{At: center.Add(10 * time.Second), Text: "too late"},
	}
	// ±5s includes the line at center-5s (boundary-inclusive).
	got := trace.FilterWindow(lines, center, 5*time.Second)
	if len(got) != 3 {
		t.Fatalf("%d %+v", len(got), got)
	}
}

func TestFilterWindow_empty(t *testing.T) {
	if len(trace.FilterWindow(nil, time.Now(), time.Second)) != 0 {
		t.Fatal()
	}
}
