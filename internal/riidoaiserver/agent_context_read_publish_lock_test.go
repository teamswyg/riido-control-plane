package riidoaiserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type contextReadBlockingSnapshot struct {
	memoryAIAgentClientSnapshotStore
	entered, release chan struct{}
}

func (s *contextReadBlockingSnapshot) SaveAIAgentClientSnapshot(ctx context.Context, snapshot AIAgentClientSnapshot) error {
	if s.entered != nil {
		close(s.entered)
		select {
		case <-s.release:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return s.memoryAIAgentClientSnapshotStore.SaveAIAgentClientSnapshot(ctx, snapshot)
}

func TestContextReadLiveDeliveryWaitsForItsPublishingSnapshot(t *testing.T) {
	s, _, _, _ := contextReadStreamServer(t)
	snapshots := &contextReadBlockingSnapshot{}
	store, err := OpenPersistentAIAgentClientStore(t.Context(), s.aiAgent.(*DevelopmentAIAgentClientStore), snapshots)
	if err != nil {
		t.Fatal(err)
	}
	s.aiAgent = store
	principal := AuthorizationResult{PrincipalID: "viewer", WorkspaceID: defaultAIAgentClientWorkspaceID}
	_, live, unsubscribe, err := store.SubscribeAIAgentClientEvents(t.Context(), principal)
	if err != nil {
		t.Fatal(err)
	}
	defer unsubscribe()
	snapshots.entered, snapshots.release = make(chan struct{}), make(chan struct{})
	released := false
	defer func() {
		if !released {
			close(snapshots.release)
		}
	}()
	saved := make(chan error, 1)
	go func() {
		saved <- store.RecordAIAgentContextReads(t.Context(), []AssignmentContextRead{{TaskID: "task", AgentID: "agent-owned-codex", AssignmentID: "assignment", SourceKind: "task", SourceID: "task", SourcePart: "body", Status: "served"}})
	}()
	<-snapshots.entered
	event := <-live
	r := httptest.NewRequest(http.MethodGet, "/events", nil)
	r.Header.Set(aiAgentTokenHeader, "viewer-token")
	delivered := make(chan bool, 1)
	go func() { _, allowed := s.clientContextReadForDelivery(r, principal, event); delivered <- allowed }()
	select {
	case <-delivered:
		t.Fatal("live hint was dropped while its own publishing snapshot held the lock")
	case <-time.After(50 * time.Millisecond):
	}
	close(snapshots.release)
	released = true
	if err := <-saved; err != nil {
		t.Fatal(err)
	}
	select {
	case allowed := <-delivered:
		if !allowed {
			t.Fatal("authorized read was not delivered after publication committed")
		}
	case <-time.After(time.Second):
		t.Fatal("read delivery did not resume after publication committed")
	}
}
