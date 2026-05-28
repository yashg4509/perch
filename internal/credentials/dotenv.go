package credentials

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
)

// ParseDotenv parses KEY=VALUE lines (no variable expansion). Empty values are omitted.
func ParseDotenv(raw string) (map[string]string, error) {
	out := make(map[string]string)
	sc := bufio.NewScanner(bytes.NewBufferString(raw))
	var lineNo int
	for sc.Scan() {
		lineNo++
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		}
		eq := strings.IndexByte(line, '=')
		if eq <= 0 {
			return nil, fmt.Errorf("credentials: dotenv line %d: missing '='", lineNo)
		}
		key := strings.TrimSpace(line[:eq])
		val := strings.TrimSpace(line[eq+1:])
		if len(val) >= 2 {
			if (val[0] == '"' && val[len(val)-1] == '"') || (val[0] == '\'' && val[len(val)-1] == '\'') {
				val = val[1 : len(val)-1]
			}
		}
		if key == "" || val == "" {
			continue
		}
		out[key] = val
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
