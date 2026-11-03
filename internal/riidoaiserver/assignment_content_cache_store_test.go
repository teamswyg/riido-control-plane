package riidoaiserver

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

type cacheAuditFailureStore struct{ *MemoryAssignmentContextStore }

func (cacheAuditFailureStore) AppendAssignmentContextReads(context.Context, []AssignmentContextRead) error {
	return errors.New("synthetic audit failure")
}

func mutateContentCacheStore(t *testing.T, f *assignmentContextFixture, req *AgentAssignmentContextRequest, name, digest string) int {
	t.Helper()
	want := http.StatusForbidden
	switch name {
	case "scope":
		f.contexts.mu.Lock()
		snapshot := f.contexts.snapshots[f.leased.ID]
		snapshot.Content = nil
		f.contexts.snapshots[f.leased.ID] = snapshot
		f.contexts.mu.Unlock()
	case "cursor":
		req.Cursor = "uncaptured"
	case "page":
		f.contexts.mu.Lock()
		f.contexts.content[f.leased.ID] = nil
		f.contexts.mu.Unlock()
		want = http.StatusServiceUnavailable
	case "page-version":
		f.contexts.mu.Lock()
		f.contexts.content[f.leased.ID][0].Version = "unpublished-version"
		f.contexts.mu.Unlock()
		want = http.StatusServiceUnavailable
	case "audit":
		config := f.service.config
		config.AssignmentContext = cacheAuditFailureStore{f.contexts}
		f.service = NewServer(config)
		f.server = f.service.Handler()
		want = http.StatusServiceUnavailable
	case "quota":
		for range 99 {
			code, hit := readContentCache(t, f, *req, "1", digest, nil)
			if code != http.StatusOK || !hit.NotModified {
				t.Fatal("read within quota failed", code)
			}
		}
		want = http.StatusTooManyRequests
	case "quota-error":
		f.service.config.AssignmentContext = &unavailableQuotaStore{AssignmentContextStore: f.contexts}
		f.server = f.service.Handler()
		want = http.StatusServiceUnavailable
	default:
		t.Fatal("unknown boundary mutation", name)
	}
	return want
}
