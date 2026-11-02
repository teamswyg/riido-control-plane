package riidoaiserver

import (
	"testing"
)

func TestMessageReactionKeepsOriginalTriggerDuringRunningFollowup(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
	action, err := store.CreateAIAgentTaskAgentAssignment(t.Context(), principal, "reaction-binding", AssignAIAgentTaskRequest{AgentID: "agent-owned-codex", AssignmentID: "run", conversationID: "conversation", SourceCommentID: "source-a", SourceMessageID: "source-a", Instruction: "A"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.CreateAIAgentTaskThreadMessage(t.Context(), principal, action.TaskID, action.ThreadID, CreateAIAgentTaskThreadMessageRequest{Body: "B", SourceMessageID: "source-b"})
	if err != nil {
		t.Fatal(err)
	}
	for _, status := range []struct {
		work  AgentWorkStatus
		state AgentAssignmentState
		emoji string
	}{{AgentWorkStatusRunning, "running", "1f440"}, {AgentWorkStatusCompleted, "completed", "2705"}} {
		store.mu.Lock()
		store.taskThreads[action.TaskID][0].WorkStatus = status.work
		store.taskThreads[action.TaskID][0].AssignmentState = status.state
		store.mu.Unlock()
		history, err := store.ListAIAgentTaskThreadHistory(t.Context(), principal, action.TaskID)
		if err != nil {
			t.Fatal(err)
		}
		for _, row := range history.Threads {
			for _, m := range row.Messages {
				if m.SourceMessageID == "source-a" && (len(m.Reactions) != 1 || m.Reactions[0].Emoji != status.emoji) {
					t.Fatalf("original trigger lost reaction: %+v", m)
				}
				if m.SourceMessageID == "source-b" && len(m.Reactions) != 0 {
					t.Fatalf("undelivered followup got reaction: %+v", m)
				}
			}
		}
	}
}
