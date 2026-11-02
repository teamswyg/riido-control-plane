package riidoaiserver

import (
	"encoding/json"
	"testing"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
)

func TestWaitingReasonAndRecipientSurviveAssignmentEventPersistence(t *testing.T) {
	store := NewStore()
	defer store.Close()
	state := newStoreState()
	assignment, err := store.handleAssign(&state, "task", AssignRequest{ComponentID: "task", AgentID: "agent", RuntimeProvider: "codex", Prompt: "work", CreatedBy: "assigner", WorkspaceID: "workspace", TeamID: "team"}, true, false)
	if err != nil {
		t.Fatal(err)
	}
	assignment.State = AssignmentRunning
	state.assignments[assignment.ID] = assignment
	request := AgentEventRequest{AssignmentID: assignment.ID, TaskID: "task", State: AssignmentRunning, EventType: assignmentcontract.EventAssignmentWaitingForUser, ConfirmationReason: assignmentcontract.ConfirmationAnswerRequired}
	response, err := store.handleEvent(&state, "agent", request)
	if err != nil {
		t.Fatal(err)
	}
	event := response.Event
	retry, err := store.handleEvent(&state, "agent", request)
	if err != nil || retry.Event.EventID != event.EventID {
		t.Fatalf("waiting retry changed event identity: %+v %v", retry, err)
	}
	if event.EventID == "" || event.CreatedBy != "assigner" || event.WorkspaceID != "workspace" || event.TeamID != "team" || event.ConfirmationReason != assignmentcontract.ConfirmationAnswerRequired {
		t.Fatalf("event lost routing/recipient: %+v", event)
	}
	raw, err := json.Marshal(response)
	if err != nil {
		t.Fatal(err)
	}
	var restored AgentEventResponse
	if err := json.Unmarshal(raw, &restored); err != nil {
		t.Fatal(err)
	}
	if restored.Assignment.ConfirmationReason != assignmentcontract.ConfirmationAnswerRequired || restored.Event.EventID != event.EventID {
		t.Fatal("snapshot lost waiting/identity")
	}
	request.EventType = EventAssignmentRunning
	request.ConfirmationReason = ""
	response, err = store.handleEvent(&state, "agent", request)
	if err != nil || response.Assignment.ConfirmationReason != "" {
		t.Fatalf("resume: %+v %v", response, err)
	}
	request.State = AssignmentFailed
	request.EventType = EventAssignmentFailed
	response, err = store.handleEvent(&state, "agent", request)
	if err != nil {
		t.Fatal(err)
	}
	retry, err = store.handleEvent(&state, "agent", request)
	if err != nil || retry.Event.EventID != response.Event.EventID {
		t.Fatalf("terminal retry changed event identity: %+v %v", retry, err)
	}
	client := NewDevelopmentAIAgentClientStore()
	client.appendTaskNotificationLocked(response.Event, AIAgentTaskActionResponse{WorkStatus: AgentWorkStatusFailed, RunID: "run", AssignmentState: AgentAssignmentStateFailed})
	last := client.events[len(client.events)-1].Payload.(AgentTaskNotificationEvent)
	if last.NotificationType != "failed" || last.RecipientPrincipalID != "assigner" {
		t.Fatalf("terminal failure: %+v", last)
	}
}
