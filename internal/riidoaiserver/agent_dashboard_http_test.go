package riidoaiserver

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestAgentDashboardHTTPPermissionsAndUnavailable(t *testing.T) {
	now := time.Now().UTC()
	for _, fail := range []bool{false, true} {
		t.Run(map[bool]string{false: "filter", true: "fail-closed"}[fail], func(t *testing.T) {
			state := newStoreState()
			for _, id := range []string{"visible", "secret"} {
				applyAssignmentToState(&state, Assignment{ID: id, TaskID: id, AgentID: "agent-owned-codex", WorkspaceID: defaultAIAgentClientWorkspaceID, State: AssignmentCompleted, CreatedAt: now, UpdatedAt: now})
			}
			store := newStoreWithConfig(StoreConfig{}, state)
			defer store.Close()
			contextReader := &dashboardTestContext{}
			server := NewServer(ServerConfig{Assignment: store, AIAgentClient: NewDevelopmentAIAgentClientStore(), Authorizer: dashboardTestAuthorizer{fail: fail}, TaskContext: contextReader})
			req := httptest.NewRequest(http.MethodGet, "/v2/client/workspaces/"+defaultAIAgentClientWorkspaceID+"/ai-agent/agents/agent-owned-codex/dashboard?page=1", nil)
			req.Header.Set("Authorization", "Bearer viewer-token")
			response := httptest.NewRecorder()
			var logs bytes.Buffer
			if fail {
				previous := log.Writer()
				log.SetOutput(&logs)
				t.Cleanup(func() { log.SetOutput(previous) })
			}
			server.Handler().ServeHTTP(response, req)
			if fail {
				if response.Code != 503 {
					t.Fatalf("status=%d", response.Code)
				}
				line := logs.String()
				if !strings.Contains(line, `route="/v2/client/workspaces/{workspace_id}/ai-agent/agents/{agent_id}/dashboard"`) ||
					!strings.Contains(line, `stage="dashboard_task_lookup"`) {
					t.Fatalf("missing dashboard failure observation: %s", line)
				}
				for _, sensitive := range []string{"viewer-token", "agent-owned-codex", "temporary authorization failure"} {
					if strings.Contains(line, sensitive) {
						t.Fatalf("dashboard log exposed private data: %s", line)
					}
				}
				return
			}
			if response.Code != 200 {
				t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
			}
			var view AgentDashboardResponse
			if err := json.Unmarshal(response.Body.Bytes(), &view); err != nil {
				t.Fatal(err)
			}
			if view.Total != 1 || view.CompletedRuns != 1 || strings.Contains(response.Body.String(), "secret") {
				t.Fatalf("unauthorized history leaked: %s", response.Body.String())
			}
			if len(contextReader.calls) != 1 || contextReader.calls[0] != "visible" {
				t.Fatalf("metadata calls=%v", contextReader.calls)
			}
		})
	}
}
