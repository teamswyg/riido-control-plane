package main

import (
	"fmt"
	"io"
	"strings"
)

func writeImpactAnnotation(w io.Writer, impact *impactEvidence) {
	if impact == nil || !impact.Enabled {
		return
	}
	fmt.Fprintf(
		w,
		"::notice title=%s::%s\n",
		githubAnnotationProperty("Riido impact scope"),
		githubAnnotationMessage(impactAnnotationMessage(impact)),
	)
}

func impactAnnotationMessage(impact *impactEvidence) string {
	if len(impact.ChangedFiles) == 0 {
		return "0 changed files"
	}
	return fmt.Sprintf(
		"%d changed files: %s",
		impact.ChangedFileCount,
		strings.Join(impact.ChangedFiles, ", "),
	)
}
