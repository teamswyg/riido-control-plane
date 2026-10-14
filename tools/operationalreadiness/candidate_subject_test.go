package main

import (
	"testing"
	"time"
)

func TestOperationalReadinessCandidatesCarryPartialSubject(t *testing.T) {
	m := candidateSubjectManifest()
	now := time.Date(2026, 6, 29, 12, 0, 0, 0, time.UTC)
	got := newCandidateEvidence(m, newEvidenceAt(m, now), now)
	if got.CandidateCount != 1 {
		t.Fatalf("candidate count = %d", got.CandidateCount)
	}
	subject := got.Candidates[0].Subject
	if subject == nil {
		t.Fatalf("candidate subject is nil: %+v", got.Candidates[0])
	}
	if subject.Kind != "operational_readiness_partial" ||
		subject.CheckID != "old" ||
		subject.Category != "stress" ||
		subject.NextArtifact != "old_evidence" ||
		subject.NextCommand != "go test ./old" ||
		subject.AgeDays != 3 ||
		!subject.Stale {
		t.Fatalf("candidate subject = %+v", subject)
	}
}

func candidateSubjectManifest() manifest {
	return manifest{
		SchemaVersion:    manifestSchema,
		Workflow:         ".github/workflows/operational-readiness.yml",
		EvidenceArtifact: "operational-readiness-evidence",
		Sources:          []producerSource{testCandidateSource()},
		Checks: []readinessCheck{{
			ID:           "old",
			Date:         "2026-06-26",
			Category:     "stress",
			Status:       "partial",
			NextArtifact: "old_evidence",
			NextCommand:  "go test ./old",
		}},
	}
}
