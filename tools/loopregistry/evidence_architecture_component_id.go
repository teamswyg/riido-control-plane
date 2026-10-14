package main

import "strings"

func architectureComponentID(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) >= 3 && parts[0] == ".github" && parts[1] == "workflows" {
		return ".github/workflows"
	}
	if len(parts) >= 2 && (parts[0] == "tools" || parts[0] == "internal") {
		return parts[0] + "/" + parts[1]
	}
	if len(parts) >= 2 && parts[0] == "docs" {
		return parts[0] + "/" + parts[1]
	}
	if len(parts) >= 2 && parts[0] == "cmd" {
		return parts[0] + "/" + parts[1]
	}
	if len(parts) > 0 && parts[0] != "" {
		return parts[0]
	}
	return path
}
