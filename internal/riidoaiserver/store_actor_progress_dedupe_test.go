package riidoaiserver

import (
	"context"
	"testing"

	"github.com/teamswyg/riido-contracts/metadatakeys"
)

func TestStoreDeduplicatesThreadProgressSeqEvents(t *testing.T) {
	ctx := context.Background()
	operations := &runtimeFakeAssignmentOperationStore{}
	store := NewStoreWithConfig(StoreConfig{OperationStore: operations})
	defer store.Close()

	assignment := mustAssignActorTask(t, store, ctx, "task-progress-dedupe", AssignRequest{
		ComponentID:     "component-progress-dedupe",
		AgentID:         "agent-1",
		RuntimeProvider: "codex",
		Prompt:          "run once",
	})
	mustPollActor(t, store, ctx, "agent-1")
	req := AgentEventRequest{
		AssignmentID: assignment.ID,
		TaskID:       assignment.TaskID,
		State:        AssignmentRunning,
		EventType:    EventRiidoLog,
		Message:      "working",
		Metadata:     map[string]string{metadatakeys.ThreadProgressSeq.String(): "7"},
	}
	first := mustRecordActorEvent(t, store, ctx, "agent-1", req)
	operationCount := len(operations.records)
	req.Message = "working duplicate"
	second := mustRecordActorEvent(t, store, ctx, "agent-1", req)
	if second.Event.Seq != first.Event.Seq || second.Event.Message != first.Event.Message {
		t.Fatalf("duplicate progress event should return first event: first=%+v second=%+v", first.Event, second.Event)
	}
	if len(operations.records) != operationCount {
		t.Fatalf("duplicate progress event persisted operation: got %d want %d", len(operations.records), operationCount)
	}
}

func TestStoreAssignmentEventKeyKeepsDifferentEventTypes(t *testing.T) {
	ctx := context.Background()
	operations := &runtimeFakeAssignmentOperationStore{}
	store := NewStoreWithConfig(StoreConfig{OperationStore: operations})
	defer store.Close()

	assignment := mustAssignActorTask(t, store, ctx, "task-event-key-type", AssignRequest{
		ComponentID:     "component-event-key-type",
		AgentID:         "agent-1",
		RuntimeProvider: "codex",
		Prompt:          "run once",
	})
	mustPollActor(t, store, ctx, "agent-1")
	metadata := map[string]string{metadatakeys.AssignmentEventKey.String(): "event-key-shared"}
	first := mustRecordActorEvent(t, store, ctx, "agent-1", AgentEventRequest{
		AssignmentID: assignment.ID, TaskID: assignment.TaskID, State: AssignmentRunning,
		EventType: EventRiidoLog, Message: "working", Metadata: metadata,
	})
	second := mustRecordActorEvent(t, store, ctx, "agent-1", AgentEventRequest{
		AssignmentID: assignment.ID, TaskID: assignment.TaskID, EventType: EventProviderWarning,
		Message: "warning", Metadata: metadata,
	})
	if second.Event.Seq == first.Event.Seq || second.Event.Type != EventProviderWarning {
		t.Fatalf("different event type was incorrectly deduped: first=%+v second=%+v", first.Event, second.Event)
	}
}
