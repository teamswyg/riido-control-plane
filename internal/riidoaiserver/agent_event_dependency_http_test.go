package riidoaiserver

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestReadyPersistenceFailureReturnsRetryableStatusAndSameKeyCommitsOnce(t *testing.T) {
	for _, tc := range []struct {
		name   string
		err    error
		status int
	}{{"storage", errors.New("private-upstream-marker"), 503}, {"deadline", context.DeadlineExceeded, 504}, {"cancelled", context.Canceled, 408}} {
		t.Run(tc.name, func(t *testing.T) {
			ops, req := lifecycleACKFixture(t, AssignmentLeased)
			req.State, req.EventType = AssignmentReady, EventAssignmentReady
			req.Metadata["assignment_event_key"] = "ready:execution"
			req.Metadata["lease_token"] = "lease"
			ops.saveErr = tc.err
			store := NewStoreWithConfig(StoreConfig{OperationStore: ops})
			t.Cleanup(store.Close)
			server := Server{assignment: store, config: ServerConfig{Authorizer: assignmentHTTPAuthorizer(t, []string{"agent:agent:events:write"})}}
			first := recordAgentEventHTTP(t, server, req)
			if first.Code != tc.status || strings.Contains(first.Body.String(), "private-upstream-marker") {
				t.Fatalf("event dependency status = %d, want %d; sensitive error omitted=%t", first.Code, tc.status, !strings.Contains(first.Body.String(), "private-upstream-marker"))
			}
			ops.mu.Lock()
			ops.saveErr = nil
			if len(ops.records) != 0 || ops.projection.Assignment.State != AssignmentLeased {
				t.Fatal("failed Ready persistence changed the durable execution")
			}
			ops.mu.Unlock()
			second, third := recordAgentEventHTTP(t, server, req), recordAgentEventHTTP(t, server, req)
			if second.Code != 200 || third.Code != 200 || second.Body.String() != third.Body.String() || len(ops.records) != 1 {
				t.Fatalf("stable-key Ready retry status = %d/%d, persisted=%d", second.Code, third.Code, len(ops.records))
			}
		})
	}
}

func TestAgentEventDependencyClassificationPreservesAuthorizationAndConflict(t *testing.T) {
	for _, tc := range []struct {
		err    error
		status int
	}{{assignmentDependencyFailure(ErrAuthorizationForbidden), 403}, {assignmentDependencyFailure(ErrAuthorizationUnauthenticated), 401}, {assignmentDependencyFailure(ErrAuthorizationInvalidCredential), 401}, {errAssignmentProjectionConflict, 503}, {errors.New("invalid assignment transition ready -> queued"), 400}} {
		status, _ := agentEventErrorResponse(tc.err)
		if status != tc.status {
			t.Fatalf("event failure classification=%d, want=%d", status, tc.status)
		}
	}
}

func recordAgentEventHTTP(t *testing.T, server Server, req AgentEventRequest) *httptest.ResponseRecorder {
	t.Helper()
	body, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest(http.MethodPost, "/v1/agents/agent/events", strings.NewReader(string(body)))
	r.Header.Set("Authorization", "Bearer assignment-token")
	w := httptest.NewRecorder()
	server.handleAgentEvent(w, r, "agent")
	return w
}
