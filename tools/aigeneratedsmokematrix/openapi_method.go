package main

import "strings"

func isHTTPMethod(value string) bool {
	switch strings.ToLower(value) {
	case "get", "post", "put", "patch", "delete", "head", "options", "trace":
		return true
	default:
		return false
	}
}
