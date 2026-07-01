package main

const publicStatusNextArtifact = "public_qa_status_pages_live_evidence"

func newPublicStatus(
	m manifest,
	partials []partialCheck,
	staleCount int,
	generatedAt string,
	expiresAt string,
) publicStatus {
	notion := newNotionEvidence(m.NotionOpenLoop)
	overall := "operational"
	if notion.PartialCount > 0 || staleCount > 0 {
		overall = "degraded"
	}
	return publicStatus{
		Overall:              overall,
		Visibility:           "public_aggregate",
		StatusPage:           m.GeneratedDoc,
		GeneratedAt:          generatedAt,
		ExpiresAt:            expiresAt,
		EvidenceTTLHours:     readinessEvidenceTTLHours,
		EndpointDetails:      "redacted",
		P0CycleCount:         notion.P0Count,
		P0PartialCount:       notion.PartialCount,
		PartialCount:         len(partials),
		StalePartialCount:    staleCount,
		ClosedLoopCandidates: staleCount,
		NextArtifact:         publicStatusNextArtifact,
	}
}
