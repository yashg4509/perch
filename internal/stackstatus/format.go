package stackstatus

import (
	"fmt"
	"sort"
	"strings"
)

// DisplayGroup buckets nodes for compact human output.
type DisplayGroup int

const (
	GroupUp DisplayGroup = iota
	GroupDown
	GroupWired
	GroupNotWired
	GroupPending
)

// GroupFor classifies a status row for presentation (not the same as Healthy).
func GroupFor(n NodeReport) DisplayGroup {
	if n.Healthy {
		return GroupUp
	}
	switch n.StatusSource {
	case SourceUnchecked:
		return GroupWired
	case SourceUnconfigured:
		return GroupNotWired
	case SourcePlaceholder:
		return GroupPending
	case SourceShell, SourceAPI, SourceProbe:
		return GroupDown
	default:
		if !n.Configured {
			return GroupNotWired
		}
		return GroupDown
	}
}

// SummaryCounts tallies nodes by display group.
type SummaryCounts struct {
	Up       int
	Down     int
	Wired    int
	NotWired int
	Pending  int
	Total    int
}

// CountGroups returns display buckets for a report.
func CountGroups(rep *EnvReport) SummaryCounts {
	var c SummaryCounts
	if rep == nil {
		return c
	}
	c.Total = len(rep.Nodes)
	for _, n := range rep.Nodes {
		switch GroupFor(n) {
		case GroupUp:
			c.Up++
		case GroupDown:
			c.Down++
		case GroupWired:
			c.Wired++
		case GroupNotWired:
			c.NotWired++
		case GroupPending:
			c.Pending++
		}
	}
	return c
}

// FormatSummaryLine is one line for headers (e.g. "2 up · 5 wired · 12 not wired").
func FormatSummaryLine(c SummaryCounts) string {
	var parts []string
	if c.Up > 0 {
		parts = append(parts, fmt.Sprintf("%d up", c.Up))
	}
	if c.Down > 0 {
		parts = append(parts, fmt.Sprintf("%d down", c.Down))
	}
	if c.Wired > 0 {
		parts = append(parts, fmt.Sprintf("%d wired", c.Wired))
	}
	if c.NotWired > 0 {
		parts = append(parts, fmt.Sprintf("%d not wired", c.NotWired))
	}
	if c.Pending > 0 {
		parts = append(parts, fmt.Sprintf("%d pending", c.Pending))
	}
	if len(parts) == 0 {
		return "0 nodes"
	}
	return strings.Join(parts, " · ")
}

// FormatHuman renders grouped status for the terminal and agent context.
func FormatHuman(stackName, env string, rep *EnvReport) string {
	if rep == nil {
		return ""
	}
	c := CountGroups(rep)
	var b strings.Builder
	if stackName != "" && env != "" {
		fmt.Fprintf(&b, "%s · %s\n", stackName, env)
	} else if env != "" {
		fmt.Fprintf(&b, "%s\n", env)
	}
	fmt.Fprintf(&b, "%s\n", FormatSummaryLine(c))

	buckets := map[DisplayGroup][]NodeReport{
		GroupUp:       nil,
		GroupDown:     nil,
		GroupWired:    nil,
		GroupNotWired: nil,
		GroupPending:  nil,
	}
	for _, n := range rep.Nodes {
		g := GroupFor(n)
		buckets[g] = append(buckets[g], n)
	}
	for _, list := range buckets {
		sort.Slice(list, func(i, j int) bool { return list[i].Name < list[j].Name })
	}

	writeNames := func(title string, nodes []NodeReport) {
		if len(nodes) == 0 {
			return
		}
		b.WriteString("\n")
		b.WriteString(title)
		b.WriteByte('\n')
		names := make([]string, len(nodes))
		for i, n := range nodes {
			names[i] = n.Name
		}
		b.WriteString("  ")
		b.WriteString(strings.Join(names, "  "))
		b.WriteByte('\n')
	}

	writeDown := func(nodes []NodeReport) {
		if len(nodes) == 0 {
			return
		}
		b.WriteString("\nDOWN\n")
		for _, n := range nodes {
			line := fmt.Sprintf("  %s (%s)", n.Name, n.Provider)
			if d := shortDetail(n.Detail); d != "" {
				line += " — " + d
			}
			b.WriteString(line)
			b.WriteByte('\n')
		}
	}

	writeUp := func(nodes []NodeReport) {
		if len(nodes) == 0 {
			return
		}
		b.WriteString("\nUP\n")
		for _, n := range nodes {
			line := fmt.Sprintf("  %s (%s)", n.Name, n.Provider)
			if n.StatusSource == SourceAppEnv && n.Detail != "" {
				line += " — " + shortDetail(n.Detail)
			}
			b.WriteString(line)
			b.WriteByte('\n')
		}
	}
	writeUp(buckets[GroupUp])
	writeDown(buckets[GroupDown])
	writeNames("SKIPPED — in perch.yaml but no credential in .env (not used unless you add keys)", buckets[GroupNotWired])
	writeNames("PENDING — deploy/host check not implemented", buckets[GroupPending])
	writeNames("UNKNOWN — configured; HTTP probe not available", buckets[GroupWired])
	return strings.TrimRight(b.String(), "\n") + "\n"
}

func shortDetail(d string) string {
	d = strings.TrimSpace(d)
	if d == "" {
		return ""
	}
	// Drop repetitive boilerplate in DOWN lines.
	for _, prefix := range []string{
		"missing credential (run perch auth sync-env)",
		"credential/config present; live vendor status API not implemented yet",
		"deployable host status API not implemented yet",
	} {
		if d == prefix {
			return ""
		}
	}
	if len(d) > 72 {
		return d[:69] + "…"
	}
	return d
}
