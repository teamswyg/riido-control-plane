package riidoaiserver

import (
	"strconv"
	"strings"
)

func compareDaemonVersions(current, required string) (int, bool) {
	a, ok := daemonVersionParts(current)
	if !ok {
		return 0, false
	}
	b, ok := daemonVersionParts(required)
	if !ok {
		return 0, false
	}
	for i := range a {
		if a[i] < b[i] {
			return -1, true
		}
		if a[i] > b[i] {
			return 1, true
		}
	}
	return 0, true
}

func daemonVersionParts(value string) ([3]int, bool) {
	var out [3]int
	fields := strings.Fields(strings.TrimSpace(value))
	if len(fields) > 0 {
		value = fields[len(fields)-1]
	}
	value = strings.TrimPrefix(strings.TrimSpace(value), "v")
	value = strings.SplitN(value, "-", 2)[0]
	parts := strings.Split(value, ".")
	if len(parts) != len(out) {
		return out, false
	}
	for i, part := range parts {
		n, err := strconv.Atoi(part)
		if err != nil || n < 0 {
			return out, false
		}
		out[i] = n
	}
	return out, true
}
