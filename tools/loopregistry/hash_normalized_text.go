package main

import "regexp"

var semanticHashFieldPattern = regexp.MustCompile(`"semantic_hash":\s*"[^"]*"`)

func normalizeSemanticHashFields(text string) string {
	return semanticHashFieldPattern.ReplaceAllString(text, `"semantic_hash":"<semantic-hash>"`)
}
