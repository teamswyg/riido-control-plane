package riidoaiserver

import (
	"context"
	"strings"
	"testing"
)

func TestAIAgentTaskThreadHistoryFollowupUsesNewAssignmentRunIdentity(t *testing.T) {
	ctx := context.Background()
	store := NewDevelopmentAIAgentClientStore()
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
	root := assignConversationThread(t, store, ctx, principal, "task-run-identity", "asn-root")
	markConversationThreadCompleted(store, "task-run-identity", root.ThreadID, root.AgentID)
	followup, err := store.CreateAIAgentTaskThreadMessage(ctx, principal, "task-run-identity", root.ThreadID, CreateAIAgentTaskThreadMessageRequest{
		Body:         "이어서 다시 해줘",
		AssignmentID: "asn-followup",
	})
	if err != nil {
		t.Fatalf("CreateAIAgentTaskThreadMessage: %v", err)
	}
	if followup.ThreadID == root.ThreadID || followup.RunID == root.RunID {
		t.Fatalf("followup reused execution identity: root=%+v followup=%+v", root, followup)
	}
	if followup.AssignmentID != "asn-followup" || !strings.Contains(followup.RunID, "asn-followup") {
		t.Fatalf("followup missing assignment-scoped run identity: %+v", followup)
	}
}
