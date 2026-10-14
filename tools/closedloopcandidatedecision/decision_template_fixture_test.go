package main

import "testing"

func ignoredCommandCandidateArtifact(t *testing.T, command string) candidateEvidence {
	t.Helper()
	generatedAt := "2026-06-24T01:00:00Z"
	expiresAt := "2026-06-25T01:00:00Z"
	run := ignoredCommandRun()
	source := ignoredCommandSource(generatedAt, expiresAt, run)
	return candidateEvidence{
		SchemaVersion:     candidateSchema,
		ID:                "loop-refresh-dispatch",
		Status:            "verified",
		SourceWorkflow:    ".github/workflows/loop-refresh-dispatch.yml",
		LiveStatus:        "ignored_commands",
		SourceGeneratedAt: generatedAt,
		SourceExpiresAt:   expiresAt,
		CandidateCount:    1,
		Candidates:        []closedLoopCandidate{ignoredCommandCandidate(t, source, command)},
		Redaction: candidateRedaction{
			SummaryOnly: true, NoRawSecrets: true, NoRawEndpoints: true,
		},
	}
}

func ignoredCommandRun() runRecord {
	return runRecord{ID: "123", Attempt: "1", SHA: "abc", RefName: "main", Event: "workflow_run"}
}

func ignoredCommandSource(generatedAt, expiresAt string, run runRecord) candidateSourceRef {
	return candidateSourceRef{
		HarnessLoop:       "loop-refresh-dispatch",
		SourceWorkflow:    ".github/workflows/loop-refresh-dispatch.yml",
		SummaryArtifact:   "loop-refresh-dispatch-plan",
		CandidateArtifact: "loop-refresh-dispatch-closed-loop-candidates",
		LiveStatus:        "ignored_commands",
		SourceGeneratedAt: generatedAt,
		SourceExpiresAt:   expiresAt,
		Run:               run,
	}
}

func ignoredCommandCandidate(t *testing.T, source candidateSourceRef, command string) closedLoopCandidate {
	t.Helper()
	return closedLoopCandidate{
		ID:              "loop-refresh-dispatch:01:closed_loop_candidate_decision:target_verifier",
		SourceRef:       source,
		Subject:         ignoredCommandSubject(t),
		HarnessLoop:     "loop-refresh-dispatch",
		PromotionTarget: "closed_loop_candidate_decision",
		PromotionEdge: graphEdge{
			From: "loop-refresh-dispatch", To: "closed_loop_candidate_decision", Relation: "promotes_failure_to",
		},
		Observation:           "Workflow dispatch ignored a verifier command.",
		Hypothesis:            "Candidate decision should choose verifier.",
		RequiredNextArtifacts: []string{"claim_binding", "verifier"},
		AdoptionPlan:          ignoredCommandAdoptionPlan(command),
	}
}
