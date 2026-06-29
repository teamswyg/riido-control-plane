package main

func ignoredCommandSubject(command selectedRefreshCommand) *candidateSubject {
	return &candidateSubject{
		Kind:                        "loop_refresh_ignored_command",
		NextArtifact:                "verifier",
		LoopID:                      command.LoopID,
		CommandKind:                 command.Kind,
		Command:                     command.Command,
		SourceCandidateID:           command.CandidateID,
		SourceSubjectKind:           command.SubjectKind,
		DecisionSource:              command.DecisionSource,
		DecisionTemplateSubjectKind: command.DecisionTemplateSubjectKind,
		ClaimIDs:                    sortedCopy(command.ClaimIDs),
		EvidenceChainIDs:            sortedCopy(command.EvidenceChainIDs),
	}
}

func staleSourceSubject(source staleRefreshSource) *candidateSubject {
	return &candidateSubject{
		Kind:        "loop_refresh_stale_source",
		SourcePath:  source.SourcePath,
		GeneratedAt: source.GeneratedAt,
		ExpiresAt:   source.ExpiresAt,
		Reason:      source.Reason,
	}
}
