package main

import "slices"

func hasString(items []string, want string) bool {
	return slices.Contains(items, want)
}
