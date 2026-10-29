package riidoaiserver

import (
	"context"
	"testing"
	"time"
)

func TestStartedNotificationOnlyForRunsWithStatusReaction(t *testing.T) {
	ctx := context.Background()
	s := NewDevelopmentAIAgentClientStore()
	s.agents["agent"] = AgentClientRecord{AgentID: "agent", OwnerPrincipalID: "owner", WorkspaceID: "ws", Visibility: AgentVisibilityPublic}
	s.taskThreads["task"] = []AIAgentTaskThreadRecord{
		{ThreadID: "from-message", TaskID: "task", AgentID: "agent", AssignmentID: "a-message", RunID: "run", WorkStatus: AgentWorkStatusQueued, AssignmentState: AgentAssignmentStateQueued},
		{ThreadID: "from-dropdown", TaskID: "task", AgentID: "agent", AssignmentID: "a-dropdown", RunID: "run", WorkStatus: AgentWorkStatusQueued, AssignmentState: AgentAssignmentStateQueued},
	}
	s.taskThreadMessages["from-message"] = []AIAgentTaskThreadHistoryMessage{{MessageID: "msg_user_1", Role: AIAgentTaskThreadMessageRoleUser, AssignmentID: "a-message", RunID: "run", Body: "please"}}
	record := func(eventID, assignmentID, eventType string, state AssignmentState) {
		t.Helper()
		event := TaskEvent{EventID: eventID, TaskID: "task", AssignmentID: assignmentID, AgentID: "agent", WorkspaceID: "ws", CreatedBy: "assigner", State: state, Type: eventType, At: time.Now()}
		if err := s.RecordAIAgentAssignmentEvent(ctx, "agent", AgentEventRequest{}, event); err != nil {
			t.Fatal(err)
		}
	}
	record("e1", "a-message", EventAssignmentRunning, AssignmentRunning)
	record("e2", "a-message", EventAssignmentRunning, AssignmentRunning)
	record("e3", "a-dropdown", EventAssignmentRunning, AssignmentRunning)
	record("e4", "a-message", EventAssignmentCancelling, AssignmentCancelling)
	record("e5", "a-message", EventAssignmentCancelled, AssignmentCancelled)

	events, err := s.AIAgentClientEvents(ctx, AuthorizationResult{PrincipalID: "assigner", WorkspaceID: "ws"})
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, event := range events {
		if n, ok := event.Payload.(AgentTaskNotificationEvent); ok {
			got = append(got, n.AssignmentID+":"+n.NotificationType)
		}
	}
	if len(got) != 1 || got[0] != "a-message:started" {
		t.Fatalf("notifications = %v, want one started push for the message run and none on cancel", got)
	}
}
