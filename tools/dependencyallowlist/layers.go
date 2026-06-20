package main

import (
	"sort"
	"strings"
)

var allowedLayers = map[string]struct{}{
	"cloud":         {},
	"contract":      {},
	"observability": {},
}

func formatAllowedLayers() string {
	layers := make([]string, 0, len(allowedLayers))
	for layer := range allowedLayers {
		layers = append(layers, layer)
	}
	sort.Strings(layers)
	return strings.Join(layers, ", ")
}
