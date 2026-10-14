package main

import (
	"fmt"
	"io"
	"strings"
)

const (
	impactViolationAnnotationLimit = 2
	impactViolationEvidenceRef     = "loop-registry-evidence"
)

func writeImpactViolationAnnotation(w io.Writer, impact *impactEvidence) {
	if impact == nil || len(impact.Violations) == 0 {
		return
	}
	fmt.Fprintf(
		w,
		"::error title=%s::%s\n",
		githubAnnotationProperty("Riido impact violation"),
		githubAnnotationMessage(impactViolationAnnotationMessage(impact)),
	)
}

func impactViolationAnnotationMessage(impact *impactEvidence) string {
	limit := impactViolationAnnotationLimit
	if len(impact.Violations) < limit {
		limit = len(impact.Violations)
	}
	parts := make([]string, 0, limit+1)
	for _, violation := range impact.Violations[:limit] {
		parts = append(parts, impactViolationSummary(violation))
	}
	if remaining := len(impact.Violations) - limit; remaining > 0 {
		parts = append(parts, fmt.Sprintf("+%d more in %s", remaining, impactViolationEvidenceRef))
	}
	return strings.Join(parts, " | ")
}

func impactViolationSummary(violation impactViolation) string {
	parts := []string{violation.Scope + ":" + violation.ClaimID, violation.Reason}
	if len(violation.RequiredBoundFiles) > 0 {
		parts = append(parts, "required files: "+strings.Join(violation.RequiredBoundFiles, ", "))
	}
	if len(violation.RequiredEvidence) > 0 {
		parts = append(parts, "required evidence: "+strings.Join(violation.RequiredEvidence, ", "))
	}
	if len(violation.RequiredReasoningEvidence) > 0 {
		parts = append(parts, "required reasoning: "+strings.Join(violation.RequiredReasoningEvidence, ", "))
	}
	return strings.Join(parts, " ")
}
