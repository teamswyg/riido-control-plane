package riidoaiserver

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type recoveryProjectionSnapshot struct {
	memoryAIAgentClientSnapshotStore
	fail bool
}

func (s *recoveryProjectionSnapshot) SaveAIAgentClientSnapshot(ctx context.Context, snapshot AIAgentClientSnapshot) error {
	if s.fail {
		return errors.New("private snapshot dependency")
	}
	return s.memoryAIAgentClientSnapshotStore.SaveAIAgentClientSnapshot(ctx, snapshot)
}

type recoveryProjectionHTTPFixture struct {
	client    *PersistentAIAgentClientStore
	snapshots *recoveryProjectionSnapshot
	ops       *conditionalTestOperations
	server    Server
	current   Assignment
	recovery  AgentEventRequest
}

func newRecoveryProjectionHTTPFixture(t *testing.T) recoveryProjectionHTTPFixture {
	t.Helper()
	snapshots := &recoveryProjectionSnapshot{}
	client, err := OpenPersistentAIAgentClientStore(context.Background(), NewDevelopmentAIAgentClientStore(), snapshots)
	if err != nil {
		t.Fatal(err)
	}
	binding, _ := client.LookupDaemonBinding("agent-owned-codex")
	current := Assignment{ID: "projection-asn", TaskID: "projection-task", AgentID: binding.AgentID, RuntimeProvider: "codex", State: AssignmentRunning, LeaseToken: "private-lease", UpdatedAt: time.Now()}
	ops := &conditionalTestOperations{projection: AssignmentProjection{Assignment: current, LastEventSeq: 10}}
	store := NewStoreWithConfig(StoreConfig{OperationStore: ops, AgentRegistry: client})
	t.Cleanup(func() { store.Close() })
	if _, err := client.AssignAIAgentTask(context.Background(), AuthorizationResult{PrincipalID: "user-1"}, current.TaskID, AssignAIAgentTaskRequest{AgentID: current.AgentID, AssignmentID: current.ID, durableState: AssignmentRunning}); err != nil {
		t.Fatal(err)
	}
	server := Server{assignment: store, aiAgent: client, config: ServerConfig{Authorizer: assignmentHTTPAuthorizer(t, []string{"agent:" + current.AgentID + ":events:write"})}}
	recovery := recoveryEventRequest(current)
	recovery.DaemonID, recovery.DeviceID, recovery.RuntimeID = binding.DaemonID, binding.DeviceID, binding.RuntimeID
	return recoveryProjectionHTTPFixture{client, snapshots, ops, server, current, recovery}
}

func (f recoveryProjectionHTTPFixture) post(t *testing.T, event AgentEventRequest, expected int) AgentEventResponse {
	t.Helper()
	body, err := json.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/v1/agents/"+f.current.AgentID+"/events", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer assignment-token")
	response := httptest.NewRecorder()
	f.server.handleAgentEvent(response, req, f.current.AgentID)
	if response.Code != expected {
		t.Fatalf("HTTP=%d want=%d", response.Code, expected)
	}
	var result AgentEventResponse
	_ = json.Unmarshal(response.Body.Bytes(), &result)
	return result
}
