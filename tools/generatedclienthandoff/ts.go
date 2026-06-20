package main

import "strings"

func ts(s string) string {
	return strings.ReplaceAll(s, "'", "\\'")
}

func tsUnescape(s string) string {
	return strings.ReplaceAll(s, "\\'", "'")
}
