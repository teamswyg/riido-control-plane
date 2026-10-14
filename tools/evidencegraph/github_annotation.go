package main

import (
	"fmt"
	"io"
	"os"
)

func writeGitHubAnnotations(w io.Writer, impact *impactEvidence) {
	if w == nil {
		w = os.Stdout
	}
	if impact == nil || !impact.Enabled {
		return
	}
	fmt.Fprintf(
		w,
		"::notice title=%s::%s\n",
		githubAnnotationProperty("Riido evidence graph impact"),
		githubAnnotationMessage(impactAnnotationMessage(impact)),
	)
	for _, chain := range impactAnnotationChains(impact) {
		fmt.Fprintf(
			w,
			"::notice title=%s::%s\n",
			githubAnnotationProperty("Riido evidence chain impact"),
			githubAnnotationMessage(chainImpactMessage(chain)),
		)
	}
}
