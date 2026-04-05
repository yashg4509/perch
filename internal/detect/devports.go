package detect

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// PortListener is a listening TCP port and optional process name (from injectable scan output in tests).
type PortListener struct {
	Port    int
	Command string
}

var portListenLine = regexp.MustCompile(`(?i)\*:(\d+)\s+\(LISTEN\)\s+(\S+)`)

// ParsePortListeners extracts listeners from simplified lsof/ss-style text (tests and fake runners).
func ParsePortListeners(text string) ([]PortListener, error) {
	var out []PortListener
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		m := portListenLine.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		port, err := strconv.Atoi(m[1])
		if err != nil {
			return nil, fmt.Errorf("detect: port: %w", err)
		}
		out = append(out, PortListener{Port: port, Command: m[2]})
	}
	return out, nil
}

// DevServiceHeuristic maps common dev ports to a short role label (spec: init --env dev).
func DevServiceHeuristic(port int) (string, bool) {
	switch port {
	case 3000:
		return "next-frontend", true
	case 4000, 8080:
		return "api", true
	case 5432:
		return "postgres", true
	case 6379:
		return "redis", true
	case 8000:
		return "python", true
	case 54321:
		return "supabase-local", true
	default:
		return "", false
	}
}
