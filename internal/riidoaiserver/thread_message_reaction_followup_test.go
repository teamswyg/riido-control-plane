package riidoaiserver

import (
	"testing"
)

// A follow-up without request_id or source_message_id starts a new execution
// that reacts on its generated user message and sends the started push.
func TestFollowupWithoutSourceReactsToGeneratedMessage(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
	action, err := store.CreateAIAgentTaskAgentAssignment(t.Context(), principal, "reaction-followup", AssignAIAgentTaskRequest{AgentID: "agent-owned-codex", AssignmentID: "run-1", conversationID: "conversation", SourceCommentID: "source-a", SourceMessageID: "source-a", Instruction: "A"})
	if err != nil {
		t.Fatal(err)
	}
	store.mu.Lock()
	store.taskThreads[action.TaskID][0].WorkStatus = AgentWorkStatusCompleted
	store.taskThreads[action.TaskID][0].AssignmentState = AgentAssignmentStateCompleted
	store.mu.Unlock()
	followup, err := store.CreateAIAgentTaskThreadMessage(t.Context(), principal, action.TaskID, action.ThreadID, CreateAIAgentTaskThreadMessageRequest{Body: "next", AssignmentID: "run-2"})
	if err != nil {
		t.Fatal(err)
	}
	if followup.ThreadID == action.ThreadID {
		t.Fatal("follow-up did not start a new execution")
	}
	store.mu.Lock()
	for i, thread := range store.taskThreads[action.TaskID] {
		if thread.ThreadID == followup.ThreadID {
			store.taskThreads[action.TaskID][i].WorkStatus = AgentWorkStatusRunning
			store.taskThreads[action.TaskID][i].AssignmentState = AgentAssignmentStateRunning
		}
	}
	store.mu.Unlock()

	history, err := store.ListAIAgentTaskThreadHistory(t.Context(), principal, action.TaskID)
	if err != nil {
		t.Fatal(err)
	}
	reacted := false
	for _, thread := range history.Threads {
		for _, message := range thread.Messages {
			if message.Role == AIAgentTaskThreadMessageRoleUser && message.Body == "next" {
				reacted = len(message.Reactions) == 1 && message.Reactions[0].Emoji == "1f440"
			}
		}
	}
	if !reacted {
		t.Fatal("generated follow-up message has no status reaction")
	}
	if got := startedNotifications(store, "run-2"); got != 1 {
		t.Fatalf("started notifications for the follow-up = %d, want 1", got)
	}
}

func startedNotifications(store *DevelopmentAIAgentClientStore, assignmentID string) int {
	store.mu.Lock()
	defer store.mu.Unlock()
	count := 0
	for _, event := range store.events {
		if notification, ok := event.Payload.(AgentTaskNotificationEvent); ok && notification.AssignmentID == assignmentID && notification.NotificationType == "started" {
			count++
		}
	}
	return count
}
