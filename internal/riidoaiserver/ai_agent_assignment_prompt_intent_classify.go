package riidoaiserver

import "strings"

func classifyTaskContextIntent(
	component AIAgentTaskContextComponent,
	document AIAgentTaskContextDocument,
	userInstruction string,
) string {
	text := strings.ToLower(component.Title + "\n" + document.Content + "\n" + userInstruction)
	explicit := hasAnswerOnlyInstruction(component.Title, true) ||
		hasAnswerOnlyInstruction(userInstruction, true) ||
		hasExplicitInstructionSignal(text) ||
		hasAnswerOnlyInstruction(document.Content, false)
	if strings.TrimSpace(document.Content) == "" && strings.TrimSpace(userInstruction) == "" {
		if explicit {
			return taskIntentExplicit
		}
		if containsAny(text, intentOrientedTaskMarkers()) {
			return taskIntentIntentOriented
		}
		return taskIntentMetadataOnly
	}
	if explicit {
		return taskIntentExplicit
	}
	if containsAny(text, intentOrientedTaskMarkers()) || nonCommandComponentType(component.ComponentType) {
		return taskIntentIntentOriented
	}
	return taskIntentExplicit
}

func nonCommandComponentType(componentType string) bool {
	switch strings.ToLower(strings.TrimSpace(componentType)) {
	case "project", "milestone", "task", "subtask":
		return true
	default:
		return false
	}
}
