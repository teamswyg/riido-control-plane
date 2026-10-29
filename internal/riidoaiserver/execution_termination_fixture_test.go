package riidoaiserver

import (
	"context"
	"crypto/sha256"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type terminationHTTPFixture struct {
	actor      *Store
	devices    *DevelopmentAIAgentClientStore
	handler    http.Handler
	assignment Assignment
}

func newTerminationHTTPFixture(t *testing.T, cfg StoreConfig) *terminationHTTPFixture {
	t.Helper()
	actor := NewStoreWithConfig(cfg)
	t.Cleanup(actor.Close)
	if _, err := actor.AssignTask(t.Context(), "task", AssignRequest{ComponentID: "component", AgentID: "agent", RuntimeProvider: "codex", Prompt: "fixture", CreatedBy: "requester"}); err != nil {
		t.Fatal(err)
	}
	ctx := context.WithValue(t.Context(), authenticatedClaimDeviceKey{}, authenticatedClaimDevice{deviceID: "device", ownerPrincipalID: "owner"})
	poll, err := actor.PollAgent(ctx, "agent", PollRequest{})
	if err != nil || poll.Action != PollStart {
		t.Fatalf("fixture claim: %v", err)
	}
	if _, err := actor.RecordAgentEvent(t.Context(), "agent", AgentEventRequest{AssignmentID: poll.Assignment.ID, State: AssignmentFailed, EventType: EventAssignmentFailed, Message: "fixture"}); err != nil {
		t.Fatal(err)
	}
	devices := NewDevelopmentAIAgentClientStore()
	devices.deviceCredentials["device"] = deviceCredentialRecord{deviceID: "device", secretHash: sha256.Sum256([]byte("secret-sentinel")), ownerPrincipalID: "owner"}
	return &terminationHTTPFixture{actor, devices, NewServer(ServerConfig{Assignment: actor, AIAgentClient: devices}).Handler(), *poll.Assignment}
}

func (f *terminationHTTPFixture) request(body string) *http.Request {
	r := httptest.NewRequest(http.MethodPost, "/v1/daemon/execution-claims/"+f.assignment.ID+"/termination", strings.NewReader(body))
	r.Header.Set(deviceIDHeader, "device")
	r.Header.Set(deviceSecretHeader, "secret-sentinel")
	r.Header.Set(executionLeaseHeader, f.assignment.LeaseToken)
	return r
}

func serveTermination(handler http.Handler, r *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	return w
}
