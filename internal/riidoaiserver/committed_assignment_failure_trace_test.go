package riidoaiserver

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/teamswyg/riido-contracts/metadatakeys"
)

func TestCommittedAssignmentFailureTracePreservesOnlyDurableFailures(t *testing.T) {
	for _, tc := range []struct {
		name, status    string
		saveErr         error
		conflicts, want int
	}{
		{name: "write-rejected", saveErr: errors.New("write rejected")},
		{name: "cas-lost", conflicts: 3},
		{name: "cas-retried", conflicts: 1, want: 1},
		{name: "committed-and-replayed", want: 1},
		{name: "cancelled", status: "cancelled"},
		{name: "needs-input", status: "needs_input"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a := Assignment{ID: "asn", TaskID: "task", AgentID: "agent", RuntimeProvider: "codex", State: AssignmentRunning, UpdatedAt: time.Now()}
			ops := &conditionalTestOperations{projection: AssignmentProjection{Assignment: a, LastEventSeq: 10}, saveErr: tc.saveErr, conflictsRemaining: tc.conflicts}
			store := NewStoreWithConfig(StoreConfig{OperationStore: ops})
			defer store.Close()
			recorder := &recordingTraceRecorder{}
			ctx := WithTraceRecorder(context.Background(), recorder)
			status := tc.status
			if status == "" {
				status = "failed"
			}
			req := AgentEventRequest{AssignmentID: a.ID, TaskID: a.TaskID, State: AssignmentFailed, EventType: EventAssignmentFailed, Metadata: map[string]string{metadatakeys.AssignmentEventKey.String(): "failure-key", metadatakeys.AssignmentResultStatus.String(): status, metadatakeys.AssignmentFailureCategory.String(): "provider_result_failed"}}
			_, err := store.RecordAgentEvent(ctx, a.AgentID, req)
			if err == nil {
				if _, err = store.RecordAgentEvent(ctx, a.AgentID, req); err != nil {
					t.Fatal(err)
				}
			}
			count := 0
			for _, span := range recorder.snapshot() {
				if span.Name != "assignment failure committed" {
					continue
				}
				count++
				if len(span.Errors) != 1 || span.Errors[0] != "assignment execution failed" || !span.Ended || span.Attributes["riido.assignment.provider"] != "codex" || span.Attributes["riido.assignment.failure_category"] != "provider_result_failed" {
					t.Fatalf("failure error evidence=%+v", span)
				}
			}
			if count != tc.want {
				t.Fatalf("committed failure spans=%d want=%d", count, tc.want)
			}
		})
	}
}
