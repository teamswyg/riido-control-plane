package riidoaiserver

import (
	"context"
	"testing"
	"time"
)

func TestOpenStoreWithConfigPreservesLiveReplayAssignment(t *testing.T) {
	for _, assignmentState := range []AssignmentState{AssignmentLeased, AssignmentReady, AssignmentRunning, AssignmentCancelling} {
		t.Run(string(assignmentState), func(t *testing.T) {
			now := time.Date(2026, 11, 12, 4, 40, 34, 0, time.UTC)
			operations := replayActiveLeaseFixture(now, assignmentState)
			store, err := OpenStoreWithConfig(context.Background(), StoreConfig{
				Now: func() time.Time { return now }, ActiveLeaseDuration: 5 * time.Minute, OperationStore: operations,
			})
			if err != nil {
				t.Fatal(err)
			}
			defer store.Close()
			metrics, err := store.Metrics(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			if metrics.AssignmentsByState[assignmentState] != 1 || metrics.AssignmentsByState[AssignmentFailed] != 0 {
				t.Fatalf("live assignment failed during replay: states=%v", metrics.AssignmentsByState)
			}
			if len(operations.records) != 1 {
				t.Fatalf("replay persisted a mutation for a live assignment: records=%d", len(operations.records))
			}
			history, _, cancel, err := store.SubscribeTask(context.Background(), "task-1")
			if err != nil {
				t.Fatal(err)
			}
			cancel()
			if len(history) != 1 || history[0].Type == EventAssignmentFailed {
				t.Fatalf("replay exposed a failure event: %v", history)
			}
			if assignmentState == AssignmentRunning {
				poll, pollErr := store.PollAgent(context.Background(), "agent-1", daemonPollRequest())
				if pollErr != nil || poll.Action != PollActive {
					t.Fatalf("resumed poll=%v err=%v", poll, pollErr)
				}
				_, eventErr := store.RecordAgentEvent(context.Background(), "agent-1", AgentEventRequest{
					AssignmentID: "asn-1", TaskID: "task-1", DaemonID: "daemon-1", RuntimeID: "runtime-1",
					State: AssignmentCompleted, EventType: EventAssignmentCompleted,
					Metadata: map[string]string{"lease_token": "lease-current"},
				})
				if eventErr != nil {
					t.Fatalf("completion after replay: %v", eventErr)
				}
				last := operations.records[len(operations.records)-1]
				if last.Assignment.State != AssignmentCompleted {
					t.Fatalf("completion persisted state=%s", last.Assignment.State)
				}
			}
		})
	}
}
