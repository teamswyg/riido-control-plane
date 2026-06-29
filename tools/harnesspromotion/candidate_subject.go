package main

import "encoding/json"

func subjectForCandidate(
	source promotionSource,
	summary liveSummary,
	claimID string,
) rawSubject {
	subject := candidateSubject{
		Kind:              "harness_failure",
		HarnessLoop:       source.HarnessLoop,
		SourceWorkflow:    source.SourceWorkflow,
		SummaryArtifact:   source.SummaryArtifact,
		CandidateArtifact: source.CandidateArtifact,
		LiveStatus:        summary.LiveStatus,
		ClaimID:           claimID,
	}
	data, err := json.Marshal(subject)
	if err != nil {
		return nil
	}
	return rawSubject(data)
}
