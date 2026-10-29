package riidoaiserver

import (
	"context"
	"testing"
)

func TestThreadMessageReactionFollowsAgentMessageAcrossRunStates(t *testing.T) {
	ctx := context.Background()
	store, owner, _, taskID, threadID := seedReactionThread(t)
	agent := reactionHistoryMessage(t, store, owner, taskID, threadID, AIAgentTaskThreadMessageRoleAgent)
	change, err := store.ReactToAIAgentTaskThreadMessage(ctx, owner, taskID, threadID, agent.MessageID, "1f44d", true)
	if err != nil || change.Notify != nil {
		t.Fatalf("agent message reaction = %+v, %v (agent messages never notify)", change, err)
	}
	setReactionThreadState(store, taskID, threadID, AgentAssignmentStateCompleted)

	completed := reactionHistoryMessage(t, store, owner, taskID, threadID, AIAgentTaskThreadMessageRoleAgent)
	if completed.MessageID == agent.MessageID {
		t.Fatal("test needs the agent message id to change with the run state")
	}
	assertReactions(t, completed.Reactions, "1f44d:member:user-1")
	user := reactionHistoryMessage(t, store, owner, taskID, threadID, AIAgentTaskThreadMessageRoleUser)
	assertReactions(t, user.Reactions, "2705:agent:agent-public-openclaw")

	setReactionThreadState(store, taskID, threadID, AgentAssignmentStateStopped)
	stopped := reactionHistoryMessage(t, store, owner, taskID, threadID, AIAgentTaskThreadMessageRoleUser)
	if len(stopped.Reactions) != 0 {
		t.Fatalf("a stopped run keeps no status reaction: %+v", stopped.Reactions)
	}
}

func setReactionThreadState(store *DevelopmentAIAgentClientStore, taskID, threadID string, state AgentAssignmentState) {
	store.mu.Lock()
	defer store.mu.Unlock()
	threads := append([]AIAgentTaskThreadRecord(nil), store.taskThreads[taskID]...)
	for i := range threads {
		if threads[i].ThreadID == threadID {
			threads[i].AssignmentState = state
		}
	}
	store.taskThreads[taskID] = threads
}
