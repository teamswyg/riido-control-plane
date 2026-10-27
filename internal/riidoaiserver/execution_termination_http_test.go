package riidoaiserver

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const terminationFixtureBody = `{"schema_version":"riido-execution-termination.v1","kind":"local_invocations_quiescent","started_invocations":2}`

func TestExecutionTerminationHTTPRecordsOriginalClaimAfterTerminal(t *testing.T) {
	actor := NewStore()
	defer actor.Close()
	_, err := actor.AssignTask(t.Context(), "task", AssignRequest{ComponentID: "component", AgentID: "agent", RuntimeProvider: "codex", Prompt: "fixture", CreatedBy: "requester"})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.WithValue(t.Context(), authenticatedClaimDeviceKey{}, authenticatedClaimDevice{"device", "owner"})
	poll, err := actor.PollAgent(ctx, "agent", PollRequest{})
	if err != nil || poll.Action != PollStart {
		t.Fatalf("claim: %v", err)
	}
	if _, err := actor.RecordAgentEvent(t.Context(), "agent", AgentEventRequest{AssignmentID: poll.Assignment.ID, State: AssignmentFailed, EventType: EventAssignmentFailed, Message: "fixture"}); err != nil {
		t.Fatal(err)
	}
	devices := NewDevelopmentAIAgentClientStore()
	devices.deviceCredentials["device"] = deviceCredentialRecord{deviceID: "device", secretHash: sha256.Sum256([]byte("secret-sentinel")), ownerPrincipalID: "owner"}
	handler := NewServer(ServerConfig{AIAgentClient: devices, Assignment: actor}).Handler()
	path := "/v1/daemon/execution-claims/" + poll.Assignment.ID + "/termination"
	request := func(body string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
		r.Header.Set(deviceIDHeader, "device")
		r.Header.Set(deviceSecretHeader, "secret-sentinel")
		r.Header.Set("X-Riido-Assignment-Lease-Token", poll.Assignment.LeaseToken)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		return w
	}
	first := request(terminationFixtureBody)
	if first.Code != http.StatusOK {
		t.Fatalf("termination status=%d", first.Code)
	}
	var receipt map[string]any
	if err := json.Unmarshal(first.Body.Bytes(), &receipt); err != nil {
		t.Fatal(err)
	}
	if receipt["lease_epoch"] != leaseEpoch(poll.Assignment.LeaseToken) || receipt["started_invocations"] != float64(2) || receipt["recorded_at"] == "" {
		t.Fatal("claim receipt mismatch")
	}
	retry := request(terminationFixtureBody)
	if retry.Code != http.StatusOK || retry.Body.String() != first.Body.String() {
		t.Fatal("retry changed original receipt")
	}
	changed := request(strings.Replace(terminationFixtureBody, `:2}`, `:3}`, 1))
	if changed.Code != http.StatusConflict {
		t.Fatalf("changed receipt status=%d", changed.Code)
	}
	observation, err := actor.observeDeviceExecutions(t.Context(), "device")
	if err != nil || observation != deviceExecutionUnknown {
		t.Fatal("local receipt granted permission")
	}
	if strings.Contains(first.Body.String(), poll.Assignment.LeaseToken) || strings.Contains(first.Body.String(), "secret-sentinel") {
		t.Fatal("response leaked credential")
	}
}
