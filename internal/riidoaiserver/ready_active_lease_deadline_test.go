package riidoaiserver

import (
	"context"
	"testing"
	"time"
)

type blockingReadyLeaseOperations struct {
	*readyActiveLeaseFailureOperations
	block     bool
	cancelled bool
}

func (s *blockingReadyLeaseOperations) LoadAgentActiveAssignment(ctx context.Context, agent string) (AssignmentActiveLease, bool, error) {
	if !s.block {
		return s.readyActiveLeaseFailureOperations.LoadAgentActiveAssignment(ctx, agent)
	}
	<-ctx.Done()
	s.cancelled = true
	return AssignmentActiveLease{}, false, ctx.Err()
}

func TestReadyActiveLeaseReadHasBoundedActorWaitAndSameKeyRetry(t *testing.T) {
	base, req := lifecycleACKFixture(t, AssignmentLeased)
	base.projection.Assignment.UpdatedAt = time.Now().Add(-time.Minute)
	ops := &blockingReadyLeaseOperations{readyActiveLeaseFailureOperations: &readyActiveLeaseFailureOperations{conditionalTestOperations: base}, block: true}
	req.State, req.EventType = AssignmentReady, EventAssignmentReady
	req.Metadata["assignment_event_key"], req.Metadata["lease_token"] = "ready:execution", "lease"
	store := NewStoreWithConfig(StoreConfig{OperationStore: ops})
	t.Cleanup(store.Close)
	server := Server{assignment: store, config: ServerConfig{Authorizer: assignmentHTTPAuthorizer(t, []string{"agent:agent:events:write"})}}
	start := time.Now()
	response := recordAgentEventHTTP(t, server, req)
	if response.Code != 504 || !ops.cancelled || time.Since(start) > 4*time.Second || len(base.records) != 0 || base.projection.Assignment.State != AssignmentLeased {
		t.Fatal("lease lookup did not release actor within its finite budget without a false ACK")
	}
	ops.block = false
	first, retry := recordAgentEventHTTP(t, server, req), recordAgentEventHTTP(t, server, req)
	if first.Code != 200 || retry.Code != 200 || first.Body.String() != retry.Body.String() || len(base.records) != 1 {
		t.Fatal("same-key Ready could not recover exactly once after lease read timeout")
	}
}
