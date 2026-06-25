package riidoaiserver

import "strings"

func classifyTaskContextIntent(
	component AIAgentTaskContextComponent,
	document AIAgentTaskContextDocument,
) string {
	text := strings.ToLower(component.Title + "\n" + document.Content)
	if strings.TrimSpace(document.Content) == "" {
		if containsAny(text, intentOrientedTaskMarkers()) {
			return taskIntentIntentOriented
		}
		return taskIntentMetadataOnly
	}
	if hasExplicitInstructionSignal(text) {
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
