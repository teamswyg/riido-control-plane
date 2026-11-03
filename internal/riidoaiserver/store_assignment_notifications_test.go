package riidoaiserver

import (
	"context"
	"reflect"
	"testing"
	"time"
)

func TestAssignmentNotificationsReadDoesNotClaimOrRefreshLease(t *testing.T) {
	ctx := context.Background()
	s := NewStore()
	defer s.Close()
	wake, release, err := s.WatchAssignmentChanges(ctx, []string{"agent-1"})
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	a, err := s.AssignTask(ctx, "task-a", AssignRequest{ComponentID: "component-1", AgentID: "agent-1", RuntimeProvider: "codex", Prompt: "hello"})
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-wake:
	case <-time.After(time.Second):
		t.Fatal("missing queued wake")
	}
	before, err := s.AssignmentSignalVersions(ctx, []string{"agent-1"})
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 10; i++ {
		after, err := s.AssignmentSignalVersions(ctx, []string{"agent-1"})
		if err != nil || !reflect.DeepEqual(before, after) {
			t.Fatalf("read mutated state: %v %v", after, err)
		}
	}
	poll, err := s.PollAgent(ctx, "agent-1", daemonPollRequest())
	if err != nil {
		t.Fatal(err)
	}
	if poll.Action != PollStart || poll.Assignment.ID != a.ID {
		t.Fatalf("stream consumed claim: %+v", poll)
	}
}

type notificationQueueReader struct {
	runtimeFakeAssignmentOperationStore
	queues map[string][]Assignment
	reads  []string
}

func (q *notificationQueueReader) LoadAgentQueueAssignments(_ context.Context, id string) ([]Assignment, error) {
	q.reads = append(q.reads, id)
	return q.queues[id], nil
}

func TestAssignmentNotificationsCrossInstanceReadAndBoundedQueries(t *testing.T) {
	ctx := context.Background()
	operations := &notificationQueueReader{queues: map[string][]Assignment{}}
	s := NewStoreWithConfig(StoreConfig{OperationStore: operations})
	defer s.Close()
	before, err := s.AssignmentSignalVersions(ctx, []string{"a", "b"})
	if err != nil {
		t.Fatal(err)
	}
	// Simulate a queue commit on another replica: no local Store notification.
	operations.queues["b"] = []Assignment{{ID: "remote", AgentID: "b", State: AssignmentQueued}}
	after, err := s.AssignmentSignalVersions(ctx, []string{"a", "b"})
	if err != nil {
		t.Fatal(err)
	}
	if before["b"] == after["b"] || before["a"] != after["a"] {
		t.Fatal("remote assignment was not isolated to b")
	}
	if !reflect.DeepEqual(operations.reads, []string{"a", "b", "a", "b"}) {
		t.Fatalf("unexpected DB query fanout: %v", operations.reads)
	}
}
