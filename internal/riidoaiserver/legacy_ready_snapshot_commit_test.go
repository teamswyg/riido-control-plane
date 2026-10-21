package riidoaiserver

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestLegacyReadySnapshotFailureDoesNotCommitReceiptStateOrPublication(t *testing.T) {
	for _, tc := range []struct {
		name   string
		err    error
		status int
	}{{"dependency", errors.New("private-snapshot-marker"), 503}, {"deadline", context.DeadlineExceeded, 504}, {"cancelled", context.Canceled, 408}} {
		t.Run(tc.name, func(t *testing.T) {
			snapshot, sink := &legacyEventSnapshot{err: tc.err}, &legacyEventSink{}
			store, server, req := legacyReadyFixture(t, StoreConfig{SnapshotStore: snapshot, Outbox: sink})
			_, events, cancel, err := store.SubscribeTask(context.Background(), req.TaskID)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(cancel)
			for range 2 {
				response := recordAgentEventHTTP(t, server, req)
				if response.Code != tc.status || strings.Contains(response.Body.String(), "private-snapshot-marker") {
					t.Fatalf("uncommitted Ready status=%d want=%d", response.Code, tc.status)
				}
			}
			projection, _, err := store.LoadAssignmentProjection(context.Background(), req.AssignmentID)
			if err != nil || projection.Assignment.State != AssignmentLeased || projection.LastEventSeq != 0 || snapshot.calls != 2 || len(snapshot.saved) != 0 || len(sink.events) != 0 {
				t.Fatal("failed persistence changed state, local receipt, publication or event sequence")
			}
			select {
			case <-events:
				t.Fatal("unpersisted event reached a subscriber")
			default:
			}
			snapshot.err = nil
			first, retry := recordAgentEventHTTP(t, server, req), recordAgentEventHTTP(t, server, req)
			if first.Code != 200 || retry.Code != 200 || first.Body.String() != retry.Body.String() || snapshot.calls != 3 || len(snapshot.saved) != 1 || len(sink.events) != 1 {
				t.Fatal("successful same-key retry did not commit and publish exactly once")
			}
			select {
			case <-events:
			default:
				t.Fatal("committed event was not published to its subscriber")
			}
			select {
			case <-events:
				t.Fatal("duplicate ACK published another event")
			default:
			}
		})
	}
}
