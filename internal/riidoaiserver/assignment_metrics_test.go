package riidoaiserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHTTPMetricsReturnsAssignmentSnapshot(t *testing.T) {
	store := NewStore()
	defer store.Close()
	assignment, err := store.AssignTask(context.Background(), "task-a", AssignRequest{
		ComponentID:     "component-a",
		AgentID:         "agent-a",
		RuntimeProvider: "codex",
		Prompt:          "ship it",
	})
	if err != nil {
		t.Fatalf("AssignTask: %v", err)
	}
	if _, err := store.PollAgent(context.Background(), "agent-a", PollRequest{DaemonID: "daemon-a", RuntimeID: "runtime-a"}); err != nil {
		t.Fatalf("PollAgent: %v", err)
	}
	if _, err := store.RecordAgentEvent(context.Background(), "agent-a", AgentEventRequest{
		AssignmentID: assignment.ID,
		DaemonID:     "daemon-a",
		RuntimeID:    "runtime-a",
		State:        AssignmentRunning,
		EventType:    EventAssignmentRunning,
	}); err != nil {
		t.Fatalf("RecordAgentEvent: %v", err)
	}
	server := NewServer(ServerConfig{Assignment: store, Authorizer: assignmentHTTPAuthorizer(t, []string{"metrics:read"})}).Handler()

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	req.Header.Set("Authorization", "Bearer assignment-token")
	resp := httptest.NewRecorder()
	server.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("metrics status=%d body=%s", resp.Code, resp.Body.String())
	}
	var snapshot MetricsSnapshot
	if err := json.Unmarshal(resp.Body.Bytes(), &snapshot); err != nil {
		t.Fatalf("metrics json: %v", err)
	}
	if snapshot.SchemaVersion != MetricsSchemaVersion {
		t.Fatalf("schema_version = %q", snapshot.SchemaVersion)
	}
	if snapshot.TasksTotal != 1 || snapshot.AssignmentsTotal != 1 || snapshot.AssignmentsByState[AssignmentRunning] != 1 {
		t.Fatalf("metrics assignments = %+v", snapshot)
	}
	if snapshot.PollRequestsTotal != 1 || snapshot.PollActionsTotal[PollStart] != 1 || snapshot.AgentEventsTotal != 1 || snapshot.TaskEventsTotal != 3 {
		t.Fatalf("metrics counters = %+v", snapshot)
	}
	if snapshot.EventAppendLatencySamplesTotal != snapshot.TaskEventsTotal {
		t.Fatalf("event append samples = %d, want task events %d", snapshot.EventAppendLatencySamplesTotal, snapshot.TaskEventsTotal)
	}
	if snapshot.StoreOperationCallsTotal != 5 || len(snapshot.StoreOperations) != 5 {
		t.Fatalf("metrics store operations = %+v", snapshot.StoreOperations)
	}
}

func TestRecordEventAppendLatencyCountsSubMillisecondSamples(t *testing.T) {
	state := newStoreState()

	recordEventAppendLatency(&state, time.Nanosecond)

	if state.eventAppendLatency.samplesTotal != 1 ||
		state.eventAppendLatency.totalMilliseconds != 1 ||
		state.eventAppendLatency.maxMilliseconds != 1 ||
		state.eventAppendLatency.lastMilliseconds != 1 {
		t.Fatalf("event append latency = %+v", state.eventAppendLatency)
	}
}

func TestHTTPMetricsRequiresScopedAuthorization(t *testing.T) {
	store := NewStore()
	defer store.Close()
	server := NewServer(ServerConfig{Assignment: store, Authorizer: assignmentHTTPAuthorizer(t, []string{"agent:agent-a:poll"})}).Handler()

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	req.Header.Set("Authorization", "Bearer assignment-token")
	resp := httptest.NewRecorder()
	server.ServeHTTP(resp, req)
	if resp.Code != http.StatusForbidden {
		t.Fatalf("metrics forbidden status=%d body=%s", resp.Code, resp.Body.String())
	}
}

func TestHTTPMetricsStoreNotConfiguredFailsClosed(t *testing.T) {
	server := NewServer(ServerConfig{Authorizer: assignmentHTTPAuthorizer(t, []string{"metrics:read"})}).Handler()

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	req.Header.Set("Authorization", "Bearer assignment-token")
	resp := httptest.NewRecorder()
	server.ServeHTTP(resp, req)
	if resp.Code != http.StatusServiceUnavailable {
		t.Fatalf("metrics unconfigured status=%d body=%s", resp.Code, resp.Body.String())
	}
	if strings.Contains(resp.Body.String(), "token") {
		t.Fatalf("metrics unconfigured leaked token wording: %s", resp.Body.String())
	}
}
