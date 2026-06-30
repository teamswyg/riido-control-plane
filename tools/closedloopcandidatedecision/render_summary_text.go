package main

import (
	"strconv"
	"strings"
)

func summaryCountText(counts []summaryCount) string {
	parts := make([]string, 0, len(counts))
	for _, count := range counts {
		parts = append(parts, "`"+count.Key+"="+strconv.Itoa(count.Count)+"`")
	}
	return strings.Join(parts, ", ")
}
