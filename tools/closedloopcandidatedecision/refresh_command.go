package main

import "strings"

const refreshCommandSchema = "riido-control-plane-loop-refresh-commands.v1"

func newRefreshCommandEvidence(result verifyResult) refreshCommandEvidence {
	generatedAt, expiresAt := evidenceWindow(candidateDecisionEvidenceTTLHours)
	commands := selectedRefreshCommands(result)
	return refreshCommandEvidence{
		SchemaVersion:     refreshCommandSchema,
		Status:            refreshCommandStatus(commands),
		GeneratedAt:       generatedAt,
		ExpiresAt:         expiresAt,
		SourceGeneratedAt: generatedAt,
		SourceExpiresAt:   expiresAt,
		SelectedLoopCount: len(selectedRefreshLoops(commands, expiresAt)),
		CommandCount:      len(commands),
		SelectedLoops:     selectedRefreshLoops(commands, expiresAt),
		Commands:          commands,
	}
}

func selectedRefreshCommands(result verifyResult) []selectedRefreshCommand {
	commands := []selectedRefreshCommand{}
	subjectKinds := subjectKindsByCandidate(result.CandidateSubjects)
	for _, item := range result.DecisionArtifacts {
		command := strings.TrimSpace(item.NextCommand)
		if command == "" {
			continue
		}
		commands = append(commands, selectedRefreshCommand{
			LoopID:      refreshLoopID(item),
			Kind:        refreshCommandKind(command),
			Command:     command,
			CandidateID: item.CandidateID,
			SubjectKind: subjectKinds[item.CandidateID],
		})
	}
	return commands
}

func refreshCommandStatus(commands []selectedRefreshCommand) string {
	if len(commands) == 0 {
		return "fresh"
	}
	return "refresh_required"
}
