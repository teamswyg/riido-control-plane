package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPAIAgentClientAssignUsesRequestScopedTaskContext(t *testing.T) {
	authorizer, err := NewStaticTokenAuthorizer([]StaticTokenCredential{{
		PrincipalID: "user-1",
		Token:       "user-token",
		Scopes:      []string{"ai-agent:*", "task:task-private:assign", "agent:agent-owned-codex:poll"},
	}})
	if err != nil {
		t.Fatalf("NewStaticTokenAuthorizer: %v", err)
	}
	aiAgentStore := NewDevelopmentAIAgentClientStore()
	aiAgentStore.devices[0].Runtimes[0].RequiresExperimentalOptIn = true
	assignmentStore := NewStoreWithConfig(StoreConfig{AgentRegistry: aiAgentStore})
	t.Cleanup(func() { assignmentStore.Close() })
	taskContext := &assignmentHTTPRequestTaskContextReader{
		contextSnapshot: AIAgentTaskContext{
			Component: AIAgentTaskContextComponent{
				ID:            "task-private",
				ComponentType: "task",
				Title:         "JWT로 task context 조회",
				KeyNumber:     "RIID-4873",
				BranchName:    "RIID-4964-agent-profile-upload",
			},
			Document: AIAgentTaskContextDocument{
				Content:       "<p>private JWT component body</p>",
				ContentFormat: "html",
			},
			Hierarchy: AIAgentTaskContextHierarchy{
				Project:   AIAgentTaskContextReference{ID: "project-a", Title: "AI Agent", KeyNumber: "RIID-4800"},
				Milestone: AIAgentTaskContextReference{ID: "milestone-a", Title: "JWT task context", KeyNumber: "RIID-4872"},
			},
			Repositories: []AIAgentTaskContextRepository{{
				FullName:      "teamswyg/riido-daemon",
				RepositoryURL: "https://github.com/teamswyg/riido-daemon",
				Source:        TaskContextRepositorySourceConnectedPullRequest,
			}},
		},
	}
	server := NewServer(ServerConfig{
		AIAgentClient: aiAgentStore,
		Assignment:    assignmentStore,
		TaskContext:   taskContext,
		Authorizer:    authorizer,
	}).Handler()

	assignReq := httptest.NewRequest(http.MethodPost, "/v2/client/workspaces/workspace-dev-riid/ai-agent/tasks/task-private/assignment", strings.NewReader(`{"agent_id":"agent-owned-codex"}`))
	assignReq.Header.Set(aiAgentTokenHeader, "user-token")
	assignResp := httptest.NewRecorder()
	server.ServeHTTP(assignResp, assignReq)
	if assignResp.Code != http.StatusAccepted {
		t.Fatalf("assign status=%d body=%s", assignResp.Code, assignResp.Body.String())
	}
	if len(taskContext.requests) != 1 {
		t.Fatalf("task context requests = %+v", taskContext.requests)
	}
	gotContextReq := taskContext.requests[0]
	if gotContextReq.ComponentID != "task-private" ||
		gotContextReq.WorkspaceID != "workspace-dev-riid" ||
		gotContextReq.BearerToken != "user-token" {
		t.Fatalf("task context request = %+v", gotContextReq)
	}

	pollReq := httptest.NewRequest(http.MethodPost, "/v1/agents/agent-owned-codex/poll", strings.NewReader(`{"daemon_id":"daemon-dev-macbook","device_id":"device-dev-macbook","runtime_id":"runtime-codex-dev"}`))
	pollReq.Header.Set("Authorization", "Bearer user-token")
	pollResp := httptest.NewRecorder()
	server.ServeHTTP(pollResp, pollReq)
	if pollResp.Code != http.StatusOK {
		t.Fatalf("poll status=%d body=%s", pollResp.Code, pollResp.Body.String())
	}
	var pollOut PollResponse
	if err := json.Unmarshal(pollResp.Body.Bytes(), &pollOut); err != nil {
		t.Fatalf("poll json: %v", err)
	}
	if pollOut.Action != PollStart || pollOut.Assignment == nil {
		t.Fatalf("poll response = %+v", pollOut)
	}
	if !strings.Contains(pollOut.Assignment.Prompt, "private JWT component body") ||
		!strings.Contains(pollOut.Assignment.Prompt, "JWT로 task context 조회") {
		t.Fatalf("assignment prompt = %s", pollOut.Assignment.Prompt)
	}
	if !pollOut.Assignment.AllowExperimentalRuntime {
		t.Fatalf("assignment experimental opt-in = false, assignment=%+v", *pollOut.Assignment)
	}
	if pollOut.Assignment.ModelID != "codex-default" {
		t.Fatalf("assignment model_id = %q, want codex-default", pollOut.Assignment.ModelID)
	}
	if pollOut.Assignment.Worktree == nil ||
		pollOut.Assignment.Worktree.RepositoryFullName != "teamswyg/riido-daemon" ||
		pollOut.Assignment.Worktree.RepositoryURL != "https://github.com/teamswyg/riido-daemon" ||
		pollOut.Assignment.Worktree.BranchName != "RIID-4964-agent-profile-upload" ||
		pollOut.Assignment.Worktree.Source != TaskContextRepositorySourceConnectedPullRequest {
		t.Fatalf("assignment worktree = %+v", pollOut.Assignment.Worktree)
	}
}
