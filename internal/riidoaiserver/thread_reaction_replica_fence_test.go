package riidoaiserver

import (
	"context"
	"testing"
)

// Another replica commits a thread message between this replica's reload and
// its reaction save. The fenced save reloads and reapplies the reaction, so
// both writes survive.
func TestThreadReactionSaveReloadsWhenAnotherReplicaCommits(t *testing.T) {
	ctx := context.Background()
	fixture := newMembershipSharedDynamoFixture(t)
	first, _ := newMembershipDynamoReplica(t, fixture)
	second, _ := newMembershipDynamoReplica(t, fixture)
	owner := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
	assigned, err := first.AssignAIAgentTask(ctx, owner, "task-replica", AssignAIAgentTaskRequest{AgentID: "agent-public-openclaw"})
	if err != nil {
		t.Fatalf("AssignAIAgentTask: %v", err)
	}
	reply, err := first.CreateAIAgentTaskThreadMessage(ctx, owner, "task-replica", assigned.ThreadID, CreateAIAgentTaskThreadMessageRequest{Body: "reply"})
	if err != nil {
		t.Fatalf("CreateAIAgentTaskThreadMessage: %v", err)
	}
	threadID := reply.ThreadID
	messageID := reactionHistoryMessage(t, first.DevelopmentAIAgentClientStore, owner, "task-replica", threadID, AIAgentTaskThreadMessageRoleUser).MessageID

	attempts := 0
	err = second.saveThreadReactionUpdate(ctx, threadID, func() (bool, error) {
		attempts++
		if attempts == 1 {
			if _, err := first.CreateAIAgentTaskThreadMessage(ctx, owner, "task-replica", threadID, CreateAIAgentTaskThreadMessageRequest{Body: "concurrent"}); err != nil {
				t.Fatalf("concurrent write: %v", err)
			}
		}
		change, err := second.DevelopmentAIAgentClientStore.ReactToAIAgentTaskThreadMessage(ctx, owner, "task-replica", threadID, messageID, "1f44d", true)
		return change.Changed, err
	})
	if err != nil || attempts != 2 {
		t.Fatalf("save = %v after %d attempts, want a fenced retry", err, attempts)
	}

	reader, _ := newMembershipDynamoReplica(t, fixture)
	history, err := reader.ListAIAgentTaskThreadHistory(ctx, owner, "task-replica")
	if err != nil {
		t.Fatal(err)
	}
	bodies, reacted := map[string]bool{}, false
	for _, thread := range history.Threads {
		for _, message := range thread.Messages {
			bodies[message.Body] = true
			reacted = reacted || (message.MessageID == messageID && len(message.Reactions) > 0)
		}
	}
	if !bodies["concurrent"] || !reacted {
		t.Fatalf("concurrent message kept=%v, reaction kept=%v", bodies["concurrent"], reacted)
	}
}
