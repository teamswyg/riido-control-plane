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
	parts := []string{changedFilesImpactMessage(impact)}
	if claimSummary := changedClaimImpactMessage(impact); claimSummary != "" {
		parts = append(parts, claimSummary)
	}
	if surfaceSummary := boundSurfaceImpactMessage(impact); surfaceSummary != "" {
		parts = append(parts, surfaceSummary)
	}
	if targetSummary := targetVerifierImpactMessage(impact); targetSummary != "" {
		parts = append(parts, targetSummary)
	}
	return strings.Join(parts, " | ")
}

func changedFilesImpactMessage(impact *impactEvidence) string {
	if len(impact.ChangedFiles) == 0 {
		return "0 changed files"
	}
	return fmt.Sprintf("%d changed files: %s",
		impact.ChangedFileCount,
		strings.Join(impact.ChangedFiles, ", "))
}

func targetVerifierImpactMessage(impact *impactEvidence) string {
	if impact.TargetVerifierPlan == nil {
		return ""
	}
	plan := impact.TargetVerifierPlan
	summary := fmt.Sprintf("target verifiers: %d matched paths, %d commands",
		plan.MatchedPathCount,
		plan.CommandCount,
	)
	return summary
}
