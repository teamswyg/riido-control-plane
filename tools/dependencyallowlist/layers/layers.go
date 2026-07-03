package layers

import (
	"sort"
	"strings"
)

var allowed = map[string]struct{}{
	"cloud":         {},
	"contract":      {},
	"observability": {},
}

func IsAllowed(layer string) bool {
	_, ok := allowed[layer]
	return ok
}

func FormatAllowed() string {
	layers := make([]string, 0, len(allowed))
	for layer := range allowed {
		layers = append(layers, layer)
	}
	sort.Strings(layers)
	return strings.Join(layers, ", ")
}
