package main

import "strings"

func containsExact(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func containsText(value, want string) bool {
	return strings.Contains(value, want)
}
