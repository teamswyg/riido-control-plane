package main

import "strings"

func newCandidate(source promotionSource, summary liveSummary, claimID, claimSummary string) closedLoopCandidate {
	return closedLoopCandidate{
		ID:                    source.ID + ":" + sanitizeID(claimID),
		SourceRef:             sourceRefForCandidate(source, summary),
		Subject:               subjectForCandidate(source, summary, claimID),
		HarnessLoop:           source.HarnessLoop,
		PromotionTarget:       source.PromotionTarget,
		PromotionEdge:         promotionEdge(source),
		Observation:           "Harness " + source.ID + " reported unverified claim " + claimID + ".",
		Hypothesis:            claimSummary,
		RequiredNextArtifacts: append([]string(nil), source.RequiredNextArtifacts...),
		AdoptionPlan:          adoptionPlan(source),
	}
}

func sourceRefForCandidate(source promotionSource, summary liveSummary) candidateSourceRef {
	return candidateSourceRef{
		HarnessLoop:       source.HarnessLoop,
		SourceWorkflow:    source.SourceWorkflow,
		SummaryArtifact:   source.SummaryArtifact,
		CandidateArtifact: source.CandidateArtifact,
		LiveStatus:        summary.LiveStatus,
		SourceGeneratedAt: summary.GeneratedAt,
		SourceExpiresAt:   summary.ExpiresAt,
		Run:               summary.Run,
	}
}

func sanitizeID(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "workflow"
	}
	return strings.NewReplacer(" ", "_", "/", "_", ":", "_").Replace(value)
}

func promotionEdge(source promotionSource) graphEdge {
	return graphEdge{
		From:     source.HarnessLoop,
		To:       source.PromotionTarget,
		Relation: "promotes_failure_to",
	}
}
