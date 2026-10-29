package riidoaiserver

import (
	"context"
	"testing"
	"time"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
)

// The first started push can leave the bounded replay buffer. A run that
// resumes after the user answered must still not push started again.
func TestStartedNotificationOncePerAssignmentAfterReplayEviction(t *testing.T) {
	ctx := context.Background()
	s := NewDevelopmentAIAgentClientStore()
	s.agents["agent"] = AgentClientRecord{AgentID: "agent", OwnerPrincipalID: "owner", WorkspaceID: "ws", Visibility: AgentVisibilityPublic}
	s.taskThreads["task"] = []AIAgentTaskThreadRecord{{ThreadID: "thread", TaskID: "task", AgentID: "agent", AssignmentID: "a", RunID: "run", WorkStatus: AgentWorkStatusQueued, AssignmentState: AgentAssignmentStateQueued}}
	s.taskThreadMessages["thread"] = []AIAgentTaskThreadHistoryMessage{{MessageID: "msg_user_1", Role: AIAgentTaskThreadMessageRoleUser, AssignmentID: "a", RunID: "run", Body: "please"}}
	record := func(eventID, eventType string, state AssignmentState, reason assignmentcontract.ConfirmationReason) {
		t.Helper()
		event := TaskEvent{EventID: eventID, TaskID: "task", AssignmentID: "a", AgentID: "agent", WorkspaceID: "ws", CreatedBy: "assigner", State: state, Type: eventType, ConfirmationReason: reason, At: time.Now()}
		if err := s.RecordAIAgentAssignmentEvent(ctx, "agent", AgentEventRequest{}, event); err != nil {
			t.Fatal(err)
		}
	}
	record("e0", EventAssignmentReady, AssignmentReady, "")
	if got := startedNotificationCount(t, s); got != 0 {
		t.Fatalf("ready pushes no started, got %d", got)
	}
	record("e1", EventAssignmentRunning, AssignmentRunning, "")
	if got := startedNotificationCount(t, s); got != 1 {
		t.Fatalf("first run pushes started once, got %d", got)
	}
	s.events = nil
	record("e2", assignmentcontract.EventAssignmentWaitingForUser, AssignmentRunning, assignmentcontract.ConfirmationAnswerRequired)
	record("e3", EventAssignmentRunning, AssignmentRunning, "")
	if got := startedNotificationCount(t, s); got != 0 {
		t.Fatalf("resumed run pushed started again: %d", got)
	}
}

func startedNotificationCount(t *testing.T, s *DevelopmentAIAgentClientStore) int {
	t.Helper()
	events, err := s.AIAgentClientEvents(context.Background(), AuthorizationResult{PrincipalID: "assigner", WorkspaceID: "ws"})
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, event := range events {
		if n, ok := event.Payload.(AgentTaskNotificationEvent); ok && n.NotificationType == agentTaskNotificationStarted {
			count++
		}
	}
	return count
}
