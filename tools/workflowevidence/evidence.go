package main

func newEvidence(m manifest, result auditResult) evidence {
	status := "verified"
	if len(result.Unregistered) > 0 || len(result.AcceptedUnused) > 0 {
		status = "failed"
	}
	unregistered := append([]string{}, result.Unregistered...)
	acceptedUnused := append([]string{}, result.AcceptedUnused...)
	return evidence{
		SchemaVersion:    evidenceSchema,
		ID:               m.ID,
		Status:           status,
		WorkflowCount:    len(result.Records),
		CoveredCount:     result.Covered,
		AcceptedGapCount: result.Accepted,
		Unregistered:     unregistered,
		AcceptedUnused:   acceptedUnused,
		Records:          result.Records,
		Loop:             m.Loop,
	}
}
