package main

import "strings"

func (r *verifyResult) add(item section) {
	if item.Kind == "slice" {
		r.Slices++
	}
	if item.Kind == "validation" {
		r.ValidationGates += countListItems(item.Body)
	}
	r.RiidoReferences += strings.Count(item.Title, "RIID-")
	for _, line := range item.Body {
		r.RiidoReferences += strings.Count(line, "RIID-")
	}
}

func countListItems(lines []string) int {
	count := 0
	for _, line := range lines {
		if strings.HasPrefix(line, "- ") {
			count++
		}
	}
	return count
}
