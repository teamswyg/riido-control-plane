package riidoaiserver

import (
	"context"
	"errors"
	"testing"
	"time"
)

type readyActiveLeaseFailureOperations struct {
	*conditionalTestOperations
	err error
}

func (s *readyActiveLeaseFailureOperations) LoadAgentActiveAssignment(context.Context, string) (AssignmentActiveLease, bool, error) {
	if s.err != nil {
		return AssignmentActiveLease{}, false, s.err
	}
	return AssignmentActiveLease{ActiveAssignmentID: s.projection.Assignment.ID, LeaseToken: "lease", LeaseExpiresAt: time.Now().Add(time.Minute)}, true, nil
}

func (*readyActiveLeaseFailureOperations) RefreshAgentActiveAssignment(context.Context, Assignment, time.Time) error {
	return nil
}

func TestReadyActiveLeaseReadFailureKeepsRetryableStatusAndNoACK(t *testing.T) {
	for _, tc := range []struct {
		name   string
		err    error
		status int
	}{{"storage", errors.New("private lease body"), 503}, {"deadline", context.DeadlineExceeded, 504}, {"cancelled", context.Canceled, 408}} {
		t.Run(tc.name, func(t *testing.T) {
			base, req := lifecycleACKFixture(t, AssignmentLeased)
			base.projection.Assignment.UpdatedAt = time.Now().Add(-time.Minute)
			ops := &readyActiveLeaseFailureOperations{conditionalTestOperations: base, err: tc.err}
			req.State, req.EventType = AssignmentReady, EventAssignmentReady
			req.Metadata["assignment_event_key"], req.Metadata["lease_token"] = "ready:execution", "lease"
			store := NewStoreWithConfig(StoreConfig{OperationStore: ops})
			t.Cleanup(store.Close)
			server := Server{assignment: store, config: ServerConfig{Authorizer: assignmentHTTPAuthorizer(t, []string{"agent:agent:events:write"})}}
			if response := recordAgentEventHTTP(t, server, req); response.Code != tc.status {
				t.Fatalf("active lease read status=%d want=%d", response.Code, tc.status)
			}
			if len(base.records) != 0 || base.projection.Assignment.State != AssignmentLeased {
				t.Fatal("failed active lease read committed Ready")
			}
			ops.err = nil
			first, retry := recordAgentEventHTTP(t, server, req), recordAgentEventHTTP(t, server, req)
			if first.Code != 200 || retry.Code != 200 || first.Body.String() != retry.Body.String() || len(base.records) != 1 {
				t.Fatal("lease read recovery did not acknowledge exactly once")
			}
		})
	}
}
