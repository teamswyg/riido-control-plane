package riidoaiserver

import (
	"context"
	"testing"
	"time"
)

func TestAIAgentClientSubscribeCompensatesMissingRecentTerminalProgressOnce(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	thread := recentCompletedClaudeThread()
	store.taskThreads[thread.TaskID] = []AIAgentTaskThreadRecord{thread}
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}

	first, _, cancelFirst, err := store.SubscribeAIAgentClientEvents(context.Background(), principal)
	if err != nil {
		t.Fatalf("first SubscribeAIAgentClientEvents: %v", err)
	}
	cancelFirst()
	if got := countTerminalProgressForThread(first, thread); got != 1 {
		t.Fatalf("first terminal compensation count = %d, want 1", got)
	}

	second, _, cancelSecond, err := store.SubscribeAIAgentClientEvents(context.Background(), principal)
	if err != nil {
		t.Fatalf("second SubscribeAIAgentClientEvents: %v", err)
	}
	cancelSecond()
	if got := countTerminalProgressForThread(second, thread); got != 1 {
		t.Fatalf("second terminal compensation count = %d, want 1", got)
	}
}

func recentCompletedClaudeThread() AIAgentTaskThreadRecord {
	now := time.Now().UTC()
	return AIAgentTaskThreadRecord{
		ThreadID: "thread-claude-terminal", TaskID: "task-claude-terminal",
		AssignmentID: "asn-claude-terminal", AgentID: "agent-owned-claude",
		RunID: "run-claude-terminal", WorkStatus: AgentWorkStatusCompleted,
		AssignmentState: AgentAssignmentStateCompleted, CommentKind: AgentTaskCommentTaskCompleted,
		StartedAt: now.Add(-time.Minute), CompletedAt: now,
	}
}

func countTerminalProgressForThread(events []ClientStreamEvent, thread AIAgentTaskThreadRecord) int {
	count := 0
	for _, event := range events {
		if terminalProgressMatchesThread(event, thread) {
			count++
		}
	}
	return count
}
