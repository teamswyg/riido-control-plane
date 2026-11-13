package riidoaiserver

import (
	"context"
	"testing"
	"time"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
)

func TestToolApprovalQuestionReplacesPreviousRunMessage(t *testing.T) {
	for _, tc := range []struct {
		name           string
		reason         assignmentcontract.ConfirmationReason
		question, want string
	}{
		{name: "approval", reason: assignmentcontract.ConfirmationApprovalRequired, question: "Allow Riido to update this task body?", want: "Allow Riido to update this task body?"},
		{name: "answer", reason: assignmentcontract.ConfirmationAnswerRequired, question: "Which task should I update?", want: "Which task should I update?"},
		{name: "empty", reason: assignmentcontract.ConfirmationApprovalRequired, want: "previous result"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store := NewDevelopmentAIAgentClientStore()
			store.agents["agent"] = AgentClientRecord{AgentID: "agent", OwnerPrincipalID: "viewer", WorkspaceID: "workspace", Visibility: AgentVisibilityPublic}
			store.taskThreads["task"] = []AIAgentTaskThreadRecord{{
				ThreadID: "thread", TaskID: "task", AgentID: "agent", AssignmentID: "assignment", RunID: "run",
				WorkStatus: AgentWorkStatusRunning, AssignmentState: AgentAssignmentStateRunning, Message: "previous result",
			}}
			event := TaskEvent{
				EventID: "approval-event", Seq: 1, TaskID: "task",
				AssignmentID: "assignment", AgentID: "agent",
				WorkspaceID: "workspace", CreatedBy: "viewer",
				State: AssignmentRunning, Type: assignmentcontract.EventAssignmentWaitingForUser,
				ConfirmationReason: tc.reason,
				Message:            tc.question, At: time.Now().UTC(),
			}
			if err := store.RecordAIAgentAssignmentEvent(context.Background(), "agent", AgentEventRequest{}, event); err != nil {
				t.Fatal(err)
			}
			thread := store.taskThreads["task"][0]
			if thread.Message != tc.want {
				t.Fatalf("approval question = %q, want %q", thread.Message, tc.want)
			}
			if thread.WorkStatus != AgentWorkStatusWaitingForUser || thread.ConfirmationReason != string(tc.reason) {
				t.Fatalf("approval did not enable the waiting reply state: %+v", thread)
			}
			if thread.ResultMessage != "" {
				t.Fatal("approval question became a terminal result")
			}
			history, err := store.ListAIAgentTaskThreads(context.Background(), AuthorizationResult{PrincipalID: "viewer", WorkspaceID: "workspace"}, "task")
			if err != nil {
				t.Fatal(err)
			}
			if len(history.Threads) != 1 || history.Threads[0].Message != tc.want ||
				history.Threads[0].WorkStatus != AgentWorkStatusWaitingForUser {
				t.Fatalf("refreshed approval history = %+v", history.Threads)
			}
			for _, entry := range store.events {
				if entry.EventType != AgentClientEventWorkStatusChanged {
					continue
				}
				status := entry.Payload.(AgentWorkStatusChangedEvent)
				if status.TaskID != event.TaskID || status.ThreadID != thread.ThreadID {
					continue
				}
				if status.WorkStatus != AgentWorkStatusWaitingForUser || status.AssignmentID != event.AssignmentID || status.RunID != thread.RunID {
					t.Fatalf("approval stream status = %+v", status)
				}
				return
			}
			t.Fatal("approval status was not published")
		})
	}
}
