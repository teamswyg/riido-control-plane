package riidoaiserver

import (
	"context"
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
