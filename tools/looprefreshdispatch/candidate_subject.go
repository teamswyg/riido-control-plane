package main

func ignoredCommandSubject(command selectedRefreshCommand) *candidateSubject {
	return &candidateSubject{
		Kind:              "loop_refresh_ignored_command",
		LoopID:            command.LoopID,
		CommandKind:       command.Kind,
		Command:           command.Command,
		SourceCandidateID: command.CandidateID,
		SourceSubjectKind: command.SubjectKind,
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
