// Package trace filters log lines by wall-clock window (spec: `perch trace --at`; parsing later).
package trace

import "time"

// Line is one parsed log line with an absolute timestamp (spec: trace --at waterfall).
type Line struct {
	At   time.Time
	Text string
}

// FilterWindow returns lines where At is within [center-radius, center+radius].
func FilterWindow(lines []Line, center time.Time, radius time.Duration) []Line {
	if len(lines) == 0 {
		return nil
	}
	minT := center.Add(-radius)
	maxT := center.Add(radius)
	var out []Line
	for _, ln := range lines {
		if !ln.At.Before(minT) && !ln.At.After(maxT) {
			out = append(out, ln)
		}
	}
	return out
}
