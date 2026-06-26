package main

import "strings"

const refreshCommandSchema = "riido-control-plane-loop-refresh-commands.v1"

func newRefreshCommandEvidence(result verifyResult) refreshCommandEvidence {
	generatedAt, expiresAt := evidenceWindow(candidateDecisionEvidenceTTLHours)
	commands := selectedRefreshCommands(result.DecisionArtifacts)
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

func selectedRefreshCommands(items []decisionArtifactEvidence) []selectedRefreshCommand {
	commands := []selectedRefreshCommand{}
	for _, item := range items {
		command := strings.TrimSpace(item.NextCommand)
		if command == "" {
			continue
		}
		commands = append(commands, selectedRefreshCommand{
			LoopID:  refreshLoopID(item),
			Kind:    refreshCommandKind(command),
			Command: command,
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
