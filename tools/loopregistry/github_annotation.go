package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

func writeGitHubAnnotations(w io.Writer, result verifyResult, impact *impactEvidence) {
	if w == nil {
		w = os.Stdout
	}
	writeImpactAnnotation(w, impact)
	writeTargetVerifierAnnotation(w, impact)
	for _, surface := range result.ClaimSurfaces {
		if len(surface.VerifierCommands) == 0 {
			continue
		}
		fmt.Fprintf(
			w,
			"::notice title=%s::%s\n",
			githubAnnotationProperty("Riido claim verifier"),
			githubAnnotationMessage(claimVerifierAnnotationMessage(surface)),
		)
	}
}

func claimVerifierAnnotationMessage(surface claimSurface) string {
	return surface.ID + " => " + strings.Join(surface.VerifierCommands, " && ")
}
