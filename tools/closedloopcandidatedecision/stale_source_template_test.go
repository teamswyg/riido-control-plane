package main

import (
	"encoding/json"
	"testing"
)

func TestCandidateDecisionUsesTemplateForLoopRefreshStaleSource(t *testing.T) {
	root := repoRootForTest(t)
	out := t.TempDir() + "/stale-source-candidates.json"
	if err := writeJSON(out, staleSourceCandidateArtifact(t)); err != nil {
		t.Fatal(err)
	}
	pinCandidateFreshnessClock(t)
	result, err := verifyCandidateDecisions(root, loadDecisionManifestForTest(t), out)
	if err != nil {
		t.Fatalf("verify candidate decision: %v", err)
	}
	if len(result.DecisionArtifacts) != 1 {
		t.Fatalf("decision artifacts = %+v", result.DecisionArtifacts)
	}
	artifact := result.DecisionArtifacts[0]
	if artifact.NextArtifact != "redacted_evidence" ||
		artifact.DecisionTemplateSubjectKind != "loop_refresh_stale_source" {
		t.Fatalf("stale-source template artifact = %+v", artifact)
	}
}

func staleSourceCandidateArtifact(t *testing.T) candidateEvidence {
	t.Helper()
	generatedAt := "2026-06-24T01:00:00Z"
	expiresAt := "2026-06-25T01:00:00Z"
	source := ignoredCommandSource(generatedAt, expiresAt, ignoredCommandRun())
	source.LiveStatus = "stale_sources"
	return candidateEvidence{
		SchemaVersion: candidateSchema, ID: "loop-refresh-dispatch", Status: "verified",
		SourceWorkflow: ".github/workflows/loop-refresh-dispatch.yml", LiveStatus: "stale_sources",
		SourceGeneratedAt: generatedAt, SourceExpiresAt: expiresAt, CandidateCount: 1,
		Candidates: []closedLoopCandidate{staleSourceCandidate(t, source)},
		Redaction:  candidateRedaction{SummaryOnly: true, NoRawSecrets: true, NoRawEndpoints: true},
	}
}

func staleSourceCandidate(t *testing.T, source candidateSourceRef) closedLoopCandidate {
	t.Helper()
	return closedLoopCandidate{
		ID: "loop-refresh-dispatch:stale-source:01", SourceRef: source,
		Subject: staleSourceSubject(t), HarnessLoop: "loop-refresh-dispatch",
		PromotionTarget: "closed_loop_candidate_decision",
		PromotionEdge:   graphEdge{"loop-refresh-dispatch", "closed_loop_candidate_decision", "promotes_failure_to"},
		Observation:     "Loop refresh dispatch could not use stale refresh command source.",
		Hypothesis:      "Refresh command evidence must be regenerated.",
		RequiredNextArtifacts: []string{
			"claim_binding", "verifier", "ci_gate", "redacted_evidence", "decision_record", "evidence_graph_edge",
		},
		AdoptionPlan: staleSourceDecisionAdoptionPlan(),
	}
}

func staleSourceSubject(t *testing.T) rawSubject {
	t.Helper()
	body, err := json.Marshal(map[string]string{"kind": "loop_refresh_stale_source"})
	if err != nil {
		t.Fatal(err)
	}
	return rawSubject(body)
}
