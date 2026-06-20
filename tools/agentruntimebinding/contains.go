package main

import "strings"

func containsText(value, want string) bool {
	return strings.Contains(value, want)
}
