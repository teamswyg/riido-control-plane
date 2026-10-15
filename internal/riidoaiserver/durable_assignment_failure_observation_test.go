package riidoaiserver

import (
	"bytes"
	"context"
	"errors"
	"log"
	"strings"
	"testing"
	"time"
)

func TestDurableFailureLogsOnlyCommittedTransition(t *testing.T) {
	for _, tc := range []struct {
		name      string
		saveErr   error
		conflicts int
		want      int
	}{
		{name: "save-rejected", saveErr: errors.New("write rejected"), want: 0},
		{name: "conflict-retried", conflicts: 1, want: 1},
		{name: "committed", want: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var output bytes.Buffer
			previous := log.Writer()
			log.SetOutput(&output)
			defer log.SetOutput(previous)
			assignment := Assignment{ID: "failure-asn", TaskID: "task", AgentID: "agent", RuntimeProvider: "codex", State: AssignmentRunning, UpdatedAt: time.Now()}
			ops := &conditionalTestOperations{projection: AssignmentProjection{Assignment: assignment, LastEventSeq: 10}, saveErr: tc.saveErr, conflictsRemaining: tc.conflicts}
			store := NewStoreWithConfig(StoreConfig{OperationStore: ops})
			defer store.Close()
			_, err := store.RecordAgentEvent(context.Background(), "agent", AgentEventRequest{AssignmentID: assignment.ID, TaskID: "task", State: AssignmentFailed, EventType: EventAssignmentFailed})
			if (err != nil) != (tc.saveErr != nil) {
				t.Fatalf("error=%v", err)
			}
			if count := strings.Count(output.String(), "event=assignment_failed"); count != tc.want {
				t.Fatalf("failure observations=%d want=%d logs=%q", count, tc.want, output.String())
			}
		})
	}
}

type rejectingEventWriteStore struct{}

func (*rejectingEventWriteStore) Close() error { return nil }
func (*rejectingEventWriteStore) SaveAssignmentOperation(_ context.Context, r AssignmentOperationRecord) error {
	if r.OperationType == AssignmentOperationAgentEvent {
		return errors.New("write rejected")
	}
	return nil
}

func TestRejectedNonConditionalFailureWriteDoesNotLogCommit(t *testing.T) {
	var output bytes.Buffer
	previous := log.Writer()
	log.SetOutput(&output)
	defer log.SetOutput(previous)
	store := NewStoreWithConfig(StoreConfig{OperationStore: &rejectingEventWriteStore{}})
	defer store.Close()
	created, err := store.AssignTask(context.Background(), "task", AssignRequest{AgentID: "agent", ComponentID: "task", RuntimeProvider: "codex", Prompt: "run"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.RecordAgentEvent(context.Background(), "agent", AgentEventRequest{AssignmentID: created.ID, State: AssignmentFailed, EventType: EventAssignmentFailed})
	if err == nil {
		t.Fatal("write failure was ignored")
	}
	if strings.Contains(output.String(), "event=assignment_failed") {
		t.Fatalf("uncommitted failure logged: %s", output.String())
	}
}
