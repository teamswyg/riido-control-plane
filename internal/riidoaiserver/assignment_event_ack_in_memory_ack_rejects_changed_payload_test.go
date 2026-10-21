package riidoaiserver

import (
	"reflect"
	"testing"
	"time"

	"github.com/teamswyg/riido-contracts/metadatakeys"
)

func TestInMemoryACKRejectsChangedPayload(t *testing.T) {
	state := newStoreState()
	now := time.Now().UTC()
	a := Assignment{ID: "local-ack", TaskID: "task", AgentID: "agent", State: AssignmentRunning, UpdatedAt: now}
	state.assignments[a.ID] = a
	store := NewStore()
	defer store.Close()
	req := AgentEventRequest{AssignmentID: a.ID, TaskID: a.TaskID, State: AssignmentFailed, EventType: EventAssignmentFailed, Message: "first", Metadata: map[string]string{metadatakeys.AssignmentEventKey.String(): "local-key"}}
	ack, err := store.handleEvent(&state, "agent", req)
	if err != nil {
		t.Fatal(err)
	}
	next := req
	next.Message = "changed"
	if _, err := store.handleEvent(&state, "agent", next); err == nil {
		t.Fatal("in-memory key reuse acknowledged changed payload")
	}
	again, err := store.handleEvent(&state, "agent", req)
	if err != nil || !reflect.DeepEqual(ack, again) || len(state.events[a.TaskID]) != 1 {
		t.Fatal("in-memory original ACK changed")
	}
}
