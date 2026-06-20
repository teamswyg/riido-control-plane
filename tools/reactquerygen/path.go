package main

import (
	"fmt"
	"regexp"
	"strings"
)

var pathParamPattern = regexp.MustCompile(`\{([^}/]+)\}`)

func pathParams(path string) []string {
	matches := pathParamPattern.FindAllStringSubmatch(path, -1)
	out := make([]string, 0, len(matches))
	for _, match := range matches {
		out = append(out, match[1])
	}
	return out
}

func pathTemplate(path string, params []string) string {
	out := fmt.Sprintf("%q", path)
	for _, param := range params {
		out = strings.ReplaceAll(out, "{"+param+"}", "${params."+safeIdentifier(param)+"}")
	}
	if len(params) > 0 {
		return "`" + strings.Trim(out, "\"") + "`"
	}
	return out
}
