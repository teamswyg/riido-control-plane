package riidoaiserver

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type legacyEventSnapshot struct {
	err   error
	calls int
	saved []StoreSnapshot
}

func (s *legacyEventSnapshot) LoadStoreSnapshot(context.Context) (StoreSnapshot, bool, error) {
	if len(s.saved) == 0 {
		return StoreSnapshot{}, false, nil
	}
	return s.saved[len(s.saved)-1], true, nil
}

func (s *legacyEventSnapshot) SaveStoreSnapshot(_ context.Context, snapshot StoreSnapshot) error {
	s.calls++
	if s.err != nil {
		return s.err
	}
	s.saved = append(s.saved, snapshot)
	return nil
}

func (*legacyEventSnapshot) Close() error { return nil }

type legacyEventOperations struct {
	err   error
	calls int
	saved []AssignmentOperationRecord
}

func (s *legacyEventOperations) SaveAssignmentOperation(_ context.Context, record AssignmentOperationRecord) error {
	s.calls++
	if s.err != nil {
		return s.err
	}
	s.saved = append(s.saved, record)
	return nil
}

func (*legacyEventOperations) Close() error { return nil }

type legacyEventSink struct{ events []TaskEvent }

func (s *legacyEventSink) AppendTaskEvent(_ context.Context, event TaskEvent) error {
	s.events = append(s.events, event)
	return nil
}

func (*legacyEventSink) Close() error { return nil }

func legacyReadyFixture(t *testing.T, config StoreConfig) (*Store, Server, AgentEventRequest) {
	t.Helper()
	ops, req := lifecycleACKFixture(t, AssignmentLeased)
	req.State, req.EventType = AssignmentReady, EventAssignmentReady
	req.Metadata["assignment_event_key"], req.Metadata["lease_token"] = "ready:execution", "lease"
	state := newStoreState()
	applyAssignmentToState(&state, ops.projection.Assignment)
	store := newStoreWithConfig(config, state)
	t.Cleanup(store.Close)
	return store, Server{assignment: store, config: ServerConfig{Authorizer: assignmentHTTPAuthorizer(t, []string{"agent:agent:events:write"})}}, req
}

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

func TestLegacyReadyOperationFailureRetriesPersistenceWithoutFalseACK(t *testing.T) {
	ops := &legacyEventOperations{err: errors.New("private-operation-marker")}
	store, server, req := legacyReadyFixture(t, StoreConfig{OperationStore: ops})
	for range 2 {
		response := recordAgentEventHTTP(t, server, req)
		if response.Code != 503 || strings.Contains(response.Body.String(), "private-operation-marker") {
			t.Fatal("legacy operation failure did not retain safe retry semantics")
		}
	}
	projection, _, err := store.LoadAssignmentProjection(context.Background(), req.AssignmentID)
	if err != nil || projection.Assignment.State != AssignmentLeased || ops.calls != 2 || len(ops.saved) != 0 {
		t.Fatal("failed operation changed execution or retained a false ACK")
	}
	ops.err = nil
	first, retry := recordAgentEventHTTP(t, server, req), recordAgentEventHTTP(t, server, req)
	if first.Code != 200 || retry.Code != 200 || first.Body.String() != retry.Body.String() || ops.calls != 3 || len(ops.saved) != 1 {
		t.Fatal("legacy journal retry did not commit exactly once")
	}
}

func TestLegacyDualStorageFailsClosedBeforeEitherNonAtomicWrite(t *testing.T) {
	ops, snapshot := &legacyEventOperations{}, &legacyEventSnapshot{err: errors.New("snapshot cannot commit")}
	store, server, req := legacyReadyFixture(t, StoreConfig{OperationStore: ops, SnapshotStore: snapshot})
	for range 2 {
		if response := recordAgentEventHTTP(t, server, req); response.Code != 503 {
			t.Fatal("non-atomic legacy dual storage was accepted")
		}
	}
	projection, _, err := store.LoadAssignmentProjection(context.Background(), req.AssignmentID)
	if err != nil || projection.Assignment.State != AssignmentLeased || projection.LastEventSeq != 0 || ops.calls != 0 || snapshot.calls != 0 {
		t.Fatal("legacy dual storage partially wrote or changed state")
	}
}
