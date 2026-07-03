package setutil

import "strings"

func ContainsExact(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func ContainsText(value, want string) bool {
	return strings.Contains(value, want)
}
