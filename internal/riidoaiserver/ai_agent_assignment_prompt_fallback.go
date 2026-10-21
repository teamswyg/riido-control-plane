package riidoaiserver

import "strings"

func ComposeAIAgentAssignmentPromptWithoutTaskContext(taskID, componentID string) (AIAgentAssignmentPrompt, error) {
	return composeAIAgentAssignmentFallback(taskID, componentID, "")
}

func composeAIAgentAssignmentFallback(taskID, componentID, instruction string) (AIAgentAssignmentPrompt, error) {
	taskID = strings.TrimSpace(taskID)
	componentID = strings.TrimSpace(componentID)
	if componentID == "" {
		componentID = taskID
	}
	return ComposeAIAgentAssignmentPrompt(AIAgentAssignmentPromptInput{
		TaskID:          taskID,
		UserInstruction: instruction,
		Context: AIAgentTaskContext{
			Component: AIAgentTaskContextComponent{
				ID:            componentID,
				ComponentType: "task",
				Title:         componentID,
			},
			Document: AIAgentTaskContextDocument{
				Content:       taskContextUnavailableDocumentContent,
				ContentFormat: "markdown",
			},
		},
	})
}
