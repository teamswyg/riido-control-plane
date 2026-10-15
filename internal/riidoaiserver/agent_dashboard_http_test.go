package riidoaiserver

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type dashboardTestAuthorizer struct{ fail bool }

func (a dashboardTestAuthorizer) Authorize(_ context.Context, _ string, req AuthorizationRequest) (AuthorizationResult, error) {
	if a.fail && req.TaskID != "" {
		return AuthorizationResult{}, errors.New("temporary authorization failure")
	}
	if req.TaskID == "secret" {
		return AuthorizationResult{}, ErrAuthorizationForbidden
	}
	return AuthorizationResult{PrincipalID: "user-1", WorkspaceID: req.WorkspaceID}, nil
}

type dashboardTestContext struct{ calls []string }

func (c *dashboardTestContext) GetAIAgentTaskContext(context.Context, string) (AIAgentTaskContext, error) {
	panic("must not use service credentials")
}

func (c *dashboardTestContext) GetAIAgentTaskContextForRequest(_ context.Context, req AIAgentTaskContextRequest) (AIAgentTaskContext, error) {
	c.calls = append(c.calls, req.ComponentID)
	if req.BearerToken != "viewer-token" {
		panic("missing viewer credentials")
	}
	return AIAgentTaskContext{Component: AIAgentTaskContextComponent{ID: req.ComponentID, Title: "Visible task", KeyNumber: "RIID-5830"}}, nil
}

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
			server.Handler().ServeHTTP(response, req)
			if fail {
				if response.Code != 503 {
					t.Fatalf("status=%d", response.Code)
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
