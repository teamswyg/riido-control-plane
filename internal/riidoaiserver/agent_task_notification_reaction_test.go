package riidoaiserver

import (
	"testing"
)

func TestTaskNotificationReactionLifecycleExclusions(t *testing.T) {
	for _, tc := range []struct {
		name       string
		target     string
		invocation *CommentInvocation
		state      AssignmentState
		work       AgentWorkStatus
		want       int
	}{
		{name: "started once", target: "message", state: AssignmentRunning, work: AgentWorkStatusRunning, want: 1},
		{name: "without reaction target", state: AssignmentRunning, work: AgentWorkStatusRunning, want: 0},
		{name: "agent to agent", target: "message", invocation: &CommentInvocation{OriginMessageID: "parent-message"}, state: AssignmentCompleted, work: AgentWorkStatusCompleted, want: 0},
		{name: "cancelled", target: "message", state: AssignmentCancelled, work: AgentWorkStatusCompleted, want: 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store := NewDevelopmentAIAgentClientStore()
			store.taskThreads["task"] = []AIAgentTaskThreadRecord{{TaskID: "task", AssignmentID: "run", SourceMessageID: tc.target, ReactionSourceMessageID: &tc.target, CommentInvocation: tc.invocation}}
			for _, id := range []string{"accepted", "running"} {
				store.appendTaskNotificationLocked(TaskEvent{EventID: id, TaskID: "task", AssignmentID: "run", WorkspaceID: "ws", CreatedBy: "user", State: tc.state}, AIAgentTaskActionResponse{WorkStatus: tc.work})
			}
			count := 0
			for _, e := range store.events {
				if n, ok := e.Payload.(AgentTaskNotificationEvent); ok {
					count++
					if n.NotificationType != "started" {
						t.Fatalf("unexpected kind %s", n.NotificationType)
					}
				}
			}
			if count != tc.want {
				t.Fatalf("notifications=%d want %d", count, tc.want)
			}
		})
	}
}
