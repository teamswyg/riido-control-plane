package riidoaiserver

import (
	"context"
	"net/http"
	"testing"
)

type cacheExpiredLeaseStore struct{ *Store }

func (cacheExpiredLeaseStore) AssignmentLeaseCurrent(context.Context, Assignment) (bool, error) {
	return false, nil
}

func mutateContentCacheBoundary(t *testing.T, f *assignmentContextFixture, req *AgentAssignmentContextRequest, name, digest string) (int, func(*http.Request)) {
	t.Helper()
	want := http.StatusForbidden
	var decorate func(*http.Request)
	switch name {
	case "device":
		decorate = func(r *http.Request) { r.Header.Set(deviceSecretHeader, "revoked") }
		want = http.StatusUnauthorized
	case "revoked-device":
		f.agents.mu.Lock()
		delete(f.agents.deviceCredentials, "device-dev-macbook")
		f.agents.mu.Unlock()
		want = http.StatusUnauthorized
	case "lease":
		req.LeaseToken = "replaced-lease"
	case "expired-lease":
		f.service.assignment = cacheExpiredLeaseStore{f.service.assignment.(*Store)}
		f.server = f.service.Handler()
		want = http.StatusConflict
	case "assignment":
		req.AssignmentID = "another-assignment"
		want = http.StatusNotFound
	case "terminal":
		f.event(AssignmentReady, EventAssignmentReady)
		f.event(AssignmentRunning, EventAssignmentRunning)
		f.event(AssignmentCompleted, EventAssignmentCompleted)
		want = http.StatusConflict
	default:
		want = mutateContentCacheStore(t, f, req, name, digest)
	}
	return want, decorate
}
