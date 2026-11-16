package riidoaiserver

import (
	"testing"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
)

func checkConsecutiveQuestions(t *testing.T, reason assignmentcontract.ConfirmationReason, questions []string) {
	t.Helper()
	store := NewStore()
	defer store.Close()
	state := newStoreState()
	assignment, err := store.handleAssign(&state, "task", AssignRequest{
		ComponentID: "task", AgentID: "agent", RuntimeProvider: "codex",
		Prompt: "work", CreatedBy: "viewer", WorkspaceID: "workspace",
	}, true, false)
	if err != nil {
		t.Fatal(err)
	}
	assignment.State = AssignmentRunning
	state.assignments[assignment.ID] = assignment
	client := NewDevelopmentAIAgentClientStore()
	client.agents["agent"] = AgentClientRecord{AgentID: "agent", OwnerPrincipalID: "viewer", WorkspaceID: "workspace"}
	client.taskThreads["task"] = []AIAgentTaskThreadRecord{{
		ThreadID: "thread", TaskID: "task", AgentID: "agent", AssignmentID: assignment.ID, RunID: "run",
		WorkStatus: AgentWorkStatusRunning, AssignmentState: AgentAssignmentStateRunning,
	}}
	request := AgentEventRequest{
		AssignmentID: assignment.ID, TaskID: "task", State: AssignmentRunning,
		EventType: assignmentcontract.EventAssignmentWaitingForUser, ConfirmationReason: reason,
	}
	var previousEventID string
	for index, question := range questions {
		request.Message = question
		response, err := store.handleEvent(&state, "agent", request)
		if err != nil {
			t.Fatal(err)
		}
		if response.Event.EventID == previousEventID || response.Event.Message != question {
			t.Fatalf("current approval reused an earlier question: got %q, want %q", response.Event.Message, question)
		}
		retry, err := store.handleEvent(&state, "agent", request)
		if err != nil || retry.Event.EventID != response.Event.EventID {
			t.Fatalf("same approval retry changed identity: %+v %v", retry.Event, err)
		}
		if err := client.RecordAIAgentAssignmentEvent(t.Context(), "agent", request, response.Event); err != nil {
			t.Fatal(err)
		}
		history, err := client.ListAIAgentTaskThreadHistory(t.Context(), AuthorizationResult{PrincipalID: "viewer", WorkspaceID: "workspace"}, "task")
		if err != nil || len(history.Threads) != 1 {
			t.Fatalf("history: %+v %v", history, err)
		}
		messages := history.Threads[0].Messages
		if len(messages) != 1 || messages[0].Body != question {
			t.Fatalf("current question missing from refreshed history: %+v", messages)
		}
		statusCount := 0
		for _, event := range client.events {
			status, ok := event.Payload.(AgentWorkStatusChangedEvent)
			if event.EventType == AgentClientEventWorkStatusChanged && ok && status.TaskID == "task" && status.AssignmentID == assignment.ID && status.RunID == "run" {
				statusCount++
			}
		}
		if statusCount != index+1 {
			t.Fatalf("changed question did not refresh the open client: got %d status events, want %d", statusCount, index+1)
		}
		request.Message = ""
		empty, err := store.handleEvent(&state, "agent", request)
		if err != nil || empty.Event.EventID != response.Event.EventID {
			t.Fatalf("empty lifecycle update replaced the current question: %+v %v", empty.Event, err)
		}
		previousEventID = response.Event.EventID
	}
}
