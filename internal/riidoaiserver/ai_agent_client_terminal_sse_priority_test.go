package riidoaiserver

import (
	"context"
	"testing"
)

func TestAIAgentClientTerminalProgressDisplacesBufferedProgress(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
	_, live, cancel, err := store.SubscribeAIAgentClientEvents(context.Background(), principal)
	if err != nil {
		t.Fatalf("SubscribeAIAgentClientEvents: %v", err)
	}
	defer cancel()

	store.mu.Lock()
	for seq := 1; seq <= cap(live); seq++ {
		store.appendClientEventLocked(AgentClientEventThreadProgress, bufferedProgressEvent(seq))
	}
	terminal := bufferedProgressEvent(cap(live) + 1)
	terminal.WorkStatus = AgentWorkStatusCompleted
	terminal.AssignmentState = AgentAssignmentStateCompleted
	terminal.CommentKind = AgentTaskCommentTaskCompleted
	store.appendClientEventLocked(AgentClientEventThreadProgress, terminal)
	subscriber := store.subscribers[1]
	store.mu.Unlock()

	if subscriber.droppedEvents != 1 || subscriber.terminalCompensations != 1 {
		t.Fatalf("subscriber delivery metrics = %+v", subscriber)
	}
	if !drainContainsTerminalProgress(live) {
		t.Fatal("buffered terminal progress was not prioritized")
	}
}

func bufferedProgressEvent(seq int) AgentThreadProgressEvent {
	return AgentThreadProgressEvent{
		EventType: AgentClientEventThreadProgress, SchemaVersion: SchemaVersion,
		AgentID: "agent-owned-claude", TaskID: "task-1", AssignmentID: "asn-buffered",
		ThreadID: "thread-task-1-claude-1", RunID: "run-claude-buffered",
		WorkStatus: AgentWorkStatusRunning, AssignmentState: AgentAssignmentStateRunning,
		CommentKind: AgentTaskCommentRuntimeProgress,
		Lines:       []AgentThreadProgressLine{{Seq: seq, Message: "progress"}},
	}
}

func drainContainsTerminalProgress(live <-chan ClientStreamEvent) bool {
	found := false
	for i := 0; i < cap(live); i++ {
		event := <-live
		found = found || clientStreamEventIsTerminalProgress(event)
	}
	return found
}
