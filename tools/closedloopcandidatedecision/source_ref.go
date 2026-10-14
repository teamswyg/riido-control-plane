package main

import "fmt"

func verifyCandidateSourceRef(item closedLoopCandidate, artifact candidateEvidence) error {
	ref := item.SourceRef
	if ref.HarnessLoop != item.HarnessLoop ||
		ref.SourceWorkflow != artifact.SourceWorkflow ||
		ref.LiveStatus != artifact.LiveStatus ||
		ref.SourceGeneratedAt != artifact.SourceGeneratedAt ||
		ref.SourceExpiresAt != artifact.SourceExpiresAt {
		return fmt.Errorf("candidate %s source_ref does not match candidate artifact", item.ID)
	}
	if ref.SummaryArtifact == "" || ref.CandidateArtifact == "" || ref.Run.ID == "" {
		return fmt.Errorf("candidate %s source_ref must bind summary, candidate, and run", item.ID)
	}
	return nil
}

func sourceRefEvidence(item closedLoopCandidate) candidateSourceRefEvidence {
	return candidateSourceRefEvidence{
		CandidateID: item.ID,
		SourceRef:   item.SourceRef,
	}
}
