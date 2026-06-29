package main

import (
	"fmt"
	"strings"
)

const annotationChangedFileLimit = 5

func changedFileAnnotationSuffix(files []string) string {
	if len(files) == 0 {
		return ""
	}
	visible := files
	if len(visible) > annotationChangedFileLimit {
		visible = visible[:annotationChangedFileLimit]
	}
	suffix := " (" + strings.Join(visible, ", ")
	if extra := len(files) - len(visible); extra > 0 {
		suffix += fmt.Sprintf(" (+%d more)", extra)
	}
	return suffix + ")"
}
