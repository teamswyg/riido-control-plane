package riidoaiserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestExecutionClaimLegacyBearerPollKeepsUnknownProvenance(t *testing.T) {
	snapshot, err := NewFileStoreSnapshot(t.TempDir() + "/snapshot.json")
	if err != nil {
		t.Fatal(err)
	}
	actor := NewStoreWithConfig(StoreConfig{SnapshotStore: snapshot})
	defer actor.Close()
	if _, err := actor.AssignTask(context.Background(), "task", AssignRequest{ComponentID: "component", AgentID: "agent", RuntimeProvider: "codex", Prompt: "fixture", CreatedBy: "requester"}); err != nil {
		t.Fatal(err)
	}
	server := NewServer(ServerConfig{Assignment: actor, Authorizer: assignmentHTTPAuthorizer(t, []string{"agent:agent:poll"})}).Handler()
	req := httptest.NewRequest(http.MethodPost, "/v1/agents/agent/poll", strings.NewReader(`{"device_id":"unverified-body-device"}`))
	req.Header.Set("Authorization", "Bearer assignment-token")
	resp := httptest.NewRecorder()
	server.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK || !strings.Contains(resp.Body.String(), `"action":"start"`) {
		t.Fatalf("legacy bearer status=%d", resp.Code)
	}
	stored, ok, err := snapshot.LoadStoreSnapshot(context.Background())
	if err != nil || !ok || len(stored.ExecutionClaims) != 0 {
		t.Fatal("bearer body forged device provenance")
	}
	observed, err := actor.observeDeviceExecutions(context.Background(), "unverified-body-device")
	if err != nil || observed != deviceExecutionUnknown {
		t.Fatal("legacy execution treated as safe")
	}
}
