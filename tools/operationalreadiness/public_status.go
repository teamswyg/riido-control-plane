package main

const publicStatusNextArtifact = "public_qa_status_visual_or_pages_publish_evidence"

func newPublicStatus(m manifest, partials []partialCheck, staleCount int) publicStatus {
	notion := newNotionEvidence(m.NotionOpenLoop)
	overall := "operational"
	if notion.PartialCount > 0 || staleCount > 0 {
		overall = "degraded"
	}
	return publicStatus{
		Overall:              overall,
		Visibility:           "public_aggregate",
		StatusPage:           m.GeneratedDoc,
		EndpointDetails:      "redacted",
		P0CycleCount:         notion.P0Count,
		P0PartialCount:       notion.PartialCount,
		PartialCount:         len(partials),
		StalePartialCount:    staleCount,
		ClosedLoopCandidates: staleCount,
		NextArtifact:         publicStatusNextArtifact,
	}
}
