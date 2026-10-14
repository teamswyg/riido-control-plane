package riidoaiserver

import (
	"errors"
	"net/http"
	"strings"
	"testing"
)

func TestHTTPAIAgentClientLegacyAssignIntentGateWaitsBeforeDurableAssignment(t *testing.T) {
	handler, assignmentStore := newIntentGateHTTPTestServer(t)
	resp := serveAssignTaskBoundary(handler, "/v1/client/ai-agent/tasks/task-copy/assignment", "user-token", `{"agent_id":"agent-owned-codex"}`)
	if resp.Code != http.StatusAccepted {
		t.Fatalf("status=%d body=%s", resp.Code, resp.Body.String())
	}
	assigned := decodeAIAgentTaskActionResponse(t, resp.Body.Bytes())
	assertIntentGateAction(t, assigned)
	poll, err := assignmentStore.PollAgent(t.Context(), assigned.AgentID, PollRequest{
		DaemonID: "daemon-dev-macbook", DeviceID: "device-dev-macbook", RuntimeID: "runtime-codex-dev",
	})
	if err != nil || poll.Action != PollNone || poll.Assignment != nil {
		t.Fatalf("legacy intent-gated assignment reached daemon: poll=%+v err=%v", poll, err)
	}
}

func TestHTTPAIAgentClientLegacyAssignSurfacesAssignmentStoreFailure(t *testing.T) {
	errStore := legacyAssignReplacementErrorStore{
		handlerAssignmentStore: &handlerAssignmentStore{},
		err:                    errors.New("assignment queue failed"),
	}
	server := legacyAssignErrorServer(t, NewDevelopmentAIAgentClientStore(), errStore)
	resp := serveAssignTaskBoundary(server, "/v1/client/ai-agent/tasks/task-boundary/assignment", "ai-agent-token", `{"agent_id":"agent-owned-codex"}`)
	if resp.Code != http.StatusBadRequest || !strings.Contains(resp.Body.String(), errStore.err.Error()) {
		t.Fatalf("status=%d body=%s", resp.Code, resp.Body.String())
	}
}

func TestHTTPAIAgentClientLegacyAssignSurfacesActionStoreFailure(t *testing.T) {
	actionStore := legacyAssignActionErrorStore{
		DevelopmentAIAgentClientStore: NewDevelopmentAIAgentClientStore(),
		err:                           errors.New("assignment projection failed"),
	}
	assignmentStore := NewStoreWithConfig(StoreConfig{AgentRegistry: actionStore})
	t.Cleanup(func() { assignmentStore.Close() })
	server := legacyAssignErrorServer(t, actionStore, assignmentStore)
	resp := serveAssignTaskBoundary(server, "/v1/client/ai-agent/tasks/task-boundary/assignment", "ai-agent-token", `{"agent_id":"agent-owned-codex"}`)
	if resp.Code != http.StatusBadRequest || !strings.Contains(resp.Body.String(), actionStore.err.Error()) {
		t.Fatalf("status=%d body=%s", resp.Code, resp.Body.String())
	}
}
