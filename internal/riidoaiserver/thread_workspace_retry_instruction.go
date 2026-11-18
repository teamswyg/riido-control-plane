package riidoaiserver

import "strings"

func workspaceRetryOriginalInstruction(thread AIAgentTaskThreadRecord, history []AIAgentTaskThreadHistoryRecord, principalID string) string {
	threadID := thread.ThreadID
	for range history {
		var parent string
		for _, row := range history {
			if row.ThreadID != threadID || row.TaskID != thread.TaskID || row.AgentID != thread.AgentID || row.ConversationID != thread.ConversationID {
				continue
			}
			for i := len(row.Messages) - 1; i >= 0; i-- {
				message := row.Messages[i]
				if message.Role != AIAgentTaskThreadMessageRoleUser || message.AssignmentID != row.AssignmentID || message.AuthorPrincipalID != principalID {
					continue
				}
				line, instruction, _ := strings.Cut(taskMessageContextText(message.Body, taskDocumentContextLimit), "\n")
				if !isWorkspaceRetryInstruction(strings.TrimSpace(line)) {
					return message.Body
				}
				if instruction = strings.TrimSpace(instruction); instruction != "" {
					return instruction
				}
				parent = row.ParentThreadID
				break
			}
			break
		}
		if parent == "" {
			break
		}
		threadID = parent
	}
	return ""
}

func workspaceRetryInstruction(instruction string, worktree *AssignmentWorktree, pulls []taskCheckoutPullRequest) (string, error) {
	firstLine, rest, _ := strings.Cut(instruction, "\n")
	if input := strings.TrimSpace(firstLine); isWorkspaceRetryInstruction(input) {
		if err := selectWorkspaceRetryCheckout(worktree, pulls, input); err != nil {
			return "", err
		}
		instruction = strings.TrimSpace(rest)
	} else if number, fallback := strings.CutPrefix(worktree.Source, checkoutPullFallbackPrefix); fallback {
		// The generated task branch just failed. Use the authorized linked
		// PR selected by the refreshed context for this attempt.
		if err := selectWorkspaceRetryCheckout(worktree, pulls, "PR #"+number); err != nil {
			return "", err
		}
	}
	return instruction, nil
}
