package rendertext

import (
	"fmt"
	"strings"
)

func CodeList(items []string) string {
	var parts []string
	for _, item := range items {
		parts = append(parts, fmt.Sprintf("`%s`", item))
	}
	return strings.Join(parts, ", ")
}

func EmptyAwareCodeList(items []string) string {
	if len(items) == 0 {
		return "`none`"
	}
	return CodeList(items)
}
