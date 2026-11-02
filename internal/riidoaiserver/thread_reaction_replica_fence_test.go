package riidoaiserver

import (
	"context"
	"sync/atomic"
	"testing"
)

// Replica B reacts with stale state, and replica A commits a new message again
// between B's reload and save. The reaction reloads, is fenced, reapplies, and
// both A's messages and B's reaction survive.
func TestThreadReactionKeepsMessagesCommittedByAnotherReplica(t *testing.T) {
	ctx := context.Background()
	fixture := newMembershipSharedDynamoFixture(t)
	first, _ := newMembershipDynamoReplica(t, fixture)
	owner := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
	assigned, err := first.AssignAIAgentTask(ctx, owner, "task-replica", AssignAIAgentTaskRequest{AgentID: "agent-public-openclaw"})
	if err != nil {
		t.Fatalf("AssignAIAgentTask: %v", err)
	}
	if _, err := first.CreateAIAgentTaskThreadMessage(ctx, owner, "task-replica", assigned.ThreadID, CreateAIAgentTaskThreadMessageRequest{Body: "reply"}); err != nil {
		t.Fatalf("CreateAIAgentTaskThreadMessage: %v", err)
	}
	second, _ := newMembershipDynamoReplica(t, fixture)
	threadID, messageID := replicaUserMessage(t, second, owner, "reply")
	if _, err := first.CreateAIAgentTaskThreadMessage(ctx, owner, "task-replica", threadID, CreateAIAgentTaskThreadMessageRequest{Body: "stale"}); err != nil {
		t.Fatalf("stale write: %v", err)
	}
	var armed atomic.Bool
	armed.Store(true)
	fixture.mu.Lock()
	fixture.beforeOwnedSnapshotRead = func() {
		if armed.Swap(false) {
			if _, err := first.CreateAIAgentTaskThreadMessage(ctx, owner, "task-replica", threadID, CreateAIAgentTaskThreadMessageRequest{Body: "concurrent"}); err != nil {
				t.Errorf("concurrent write: %v", err)
			}
		}
	}
	fixture.mu.Unlock()

	if _, err := second.MutateTaskMessageReaction(ctx, owner, "task-replica", threadID, messageID, "1f44d", false); err != nil {
		t.Fatalf("MutateTaskMessageReaction: %v", err)
	}
	if armed.Load() {
		t.Fatal("the concurrent commit did not run before the reaction save")
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
			if message.MessageID == messageID {
				reacted = memberReactionEmoji(message.Reactions, owner.PrincipalID) == "1f44d"
			}
		}
	}
	if !bodies["stale"] || !bodies["concurrent"] || !reacted {
		t.Fatalf("stale kept=%v, concurrent kept=%v, reaction kept=%v", bodies["stale"], bodies["concurrent"], reacted)
	}
}
