package riidoaiserver

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
)

type unavailableQuotaStore struct {
	AssignmentContextStore
	loaded bool
}

func (s *unavailableQuotaStore) ConsumeAssignmentContextRead(context.Context, string) (bool, error) {
	return false, errors.New("quota backend unavailable")
}

func (s *unavailableQuotaStore) LoadAssignmentContextSnapshot(ctx context.Context, id string) (AssignmentContextSnapshot, bool, error) {
	s.loaded = true
	return s.AssignmentContextStore.LoadAssignmentContextSnapshot(ctx, id)
}

func TestAssignmentContextQuotaUnavailableFailsBeforeAnyDataRead(t *testing.T) {
	f := newAssignmentContextFixture(t, NewMemoryAssignmentContextStore())
	fault := &unavailableQuotaStore{AssignmentContextStore: f.contexts}
	f.service.config.AssignmentContext = fault
	f.server = f.service.Handler()
	for _, view := range []string{AssignmentContextViewBase, AssignmentContextViewAgents} {
		code, body := f.read(view, "", f.leased.LeaseToken)
		if code != http.StatusServiceUnavailable || fault.loaded || strings.Contains(body, `"data"`) {
			t.Fatalf("view=%s status=%d loaded=%v body=%s", view, code, fault.loaded, body)
		}
	}
}

func TestAssignmentContextQuotaCountsUnavailableSnapshotAttempts(t *testing.T) {
	f := newAssignmentContextFixture(t, NewMemoryAssignmentContextStore())
	f.contexts.mu.Lock()
	delete(f.contexts.snapshots, f.leased.ID)
	f.contexts.mu.Unlock()
	for i := range 100 {
		if code, _ := f.read(AssignmentContextViewBase, "", f.leased.LeaseToken); code != http.StatusNotFound {
			t.Fatalf("missing snapshot attempt %d status=%d", i+1, code)
		}
	}
	if code, _ := f.read(AssignmentContextViewBase, "", f.leased.LeaseToken); code != http.StatusTooManyRequests {
		t.Fatalf("attempt101=%d", code)
	}
}

func TestAssignmentContextQuotaExcludesViewerReadHistory(t *testing.T) {
	server, _, _ := runReadsTestServer(t, nil)
	store := server.config.AssignmentContext
	for range 101 {
		resp := runReadsRequest(server, "/v1/client/ai-agent/runs/run/reads")
		if resp.Code != http.StatusOK {
			t.Fatalf("viewer reads status=%d body=%s", resp.Code, resp.Body.String())
		}
	}
	for i := range 100 {
		if ok, err := store.ConsumeAssignmentContextRead(t.Context(), "assignment"); !ok || err != nil {
			t.Fatalf("viewer consumed CLI budget at %d: allowed=%v err=%v", i+1, ok, err)
		}
	}
	if resp := runReadsRequest(server, "/v1/client/ai-agent/runs/run/reads"); resp.Code != http.StatusOK {
		t.Fatalf("viewer quota-limited: %d", resp.Code)
	}
}
