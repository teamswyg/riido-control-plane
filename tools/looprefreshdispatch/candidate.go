package main

func candidateEvidenceFromPlan(plan dispatchPlan) candidateEvidence {
	liveStatus := "no_ignored_commands"
	if len(plan.IgnoredCommands) > 0 {
		liveStatus = "ignored_commands"
	}
	candidates := candidatesFromPlan(plan, liveStatus)
	return candidateEvidence{
		SchemaVersion:     candidateSchema,
		ID:                dispatchSourceID,
		Status:            "verified",
		SourceWorkflow:    dispatchSourceWorkflow,
		LiveStatus:        liveStatus,
		SourceGeneratedAt: plan.GeneratedAt,
		SourceExpiresAt:   plan.ExpiresAt,
		Run:               githubRunRecord(),
		CandidateCount:    len(candidates),
		Candidates:        candidates,
		Redaction:         candidateRedaction{true, true, true, true},
	}
}

func candidatesFromPlan(plan dispatchPlan, liveStatus string) []closedLoopCandidate {
	out := make([]closedLoopCandidate, 0, len(plan.IgnoredCommands))
	for index, command := range plan.IgnoredCommands {
		out = append(out, candidateFromIgnoredCommand(plan, liveStatus, command, index))
	}
	return out
}
