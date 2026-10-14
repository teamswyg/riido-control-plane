package main

import (
	"fmt"
	"strings"
)

func exportedName(name string) string {
	if name == "" {
		return ""
	}
	return strings.ToUpper(name[:1]) + name[1:]
}

func optionalMark(required bool) string {
	if required {
		return ""
	}
	return "?"
}

func quoteProperty(name string) string {
	if safeIdentifier(name) == name {
		return name
	}
	return fmt.Sprintf("%q", name)
}

func safeIdentifier(name string) string {
	return strings.ReplaceAll(name, "-", "_")
}
