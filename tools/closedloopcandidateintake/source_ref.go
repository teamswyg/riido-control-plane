package main

import "fmt"

func verifyCandidateSourceRef(
	item closedLoopCandidate,
	artifact candidateEvidence,
	source intakeSource,
) error {
	ref := item.SourceRef
	if ref.HarnessLoop != item.HarnessLoop ||
		ref.HarnessLoop != source.HarnessLoop ||
		ref.SourceWorkflow != artifact.SourceWorkflow ||
		ref.LiveStatus != artifact.LiveStatus ||
		ref.SourceGeneratedAt != artifact.SourceGeneratedAt ||
		ref.SourceExpiresAt != artifact.SourceExpiresAt ||
		ref.CandidateArtifact != source.CandidateArtifact {
		return fmt.Errorf("candidate %s source_ref does not match source artifact", item.ID)
	}
	if ref.SummaryArtifact == "" || ref.Run.ID == "" {
		return fmt.Errorf("candidate %s source_ref must bind summary artifact and run", item.ID)
	}
	return nil
}
