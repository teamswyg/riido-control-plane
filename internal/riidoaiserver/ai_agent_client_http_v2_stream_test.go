package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPAIAgentClientDevelopmentV2WorkspaceScopedCreateAndThreadStream(t *testing.T) {
	server := newAIAgentClientHTTPTestServer(t, []StaticTokenCredential{{
		PrincipalID: "user-1",
		Token:       "owner-token",
		Scopes:      []string{"ai-agent:*", "agent:*:poll"},
	}})

	body, err := json.Marshal(CreateAgentConfigurationRequest{
		Name:       "워크스페이스 A 에이전트",
		Visibility: AgentVisibilityPublic,
		RuntimeID:  "runtime-cursor-dev",
		ModelID:    stringPtr("cursor-fast"),
	})
	if err != nil {
		t.Fatalf("marshal create body: %v", err)
	}
	createReq := httptest.NewRequest(http.MethodPost, "/v2/client/workspaces/workspace-a/ai-agent/agents", strings.NewReader(string(body)))
	createReq.Header.Set(aiAgentTokenHeader, "owner-token")
	createResp := httptest.NewRecorder()
	server.ServeHTTP(createResp, createReq)
	if createResp.Code != http.StatusCreated {
		t.Fatalf("v2 create status=%d body=%s", createResp.Code, createResp.Body.String())
	}
	var created AgentClientRecordResponse
	if err := json.Unmarshal(createResp.Body.Bytes(), &created); err != nil {
		t.Fatalf("v2 create json: %v", err)
	}
	if created.Agent.WorkspaceID != "workspace-a" ||
		created.Agent.OwnerPrincipalID != "user-1" ||
		created.Agent.RuntimeID != "runtime-cursor-dev" ||
		created.Agent.ModelID != "cursor-fast" {
		t.Fatalf("v2 created agent = %+v", created.Agent)
	}

	bootstrapAReq := httptest.NewRequest(http.MethodGet, "/v2/client/workspaces/workspace-a/ai-agent/bootstrap?client_kind=web", nil)
	bootstrapAReq.Header.Set(aiAgentTokenHeader, "owner-token")
	bootstrapAResp := httptest.NewRecorder()
	server.ServeHTTP(bootstrapAResp, bootstrapAReq)
	if bootstrapAResp.Code != http.StatusOK {
		t.Fatalf("v2 workspace-a bootstrap status=%d body=%s", bootstrapAResp.Code, bootstrapAResp.Body.String())
	}
	var bootstrapA ClientBootstrapResponse
	if err := json.Unmarshal(bootstrapAResp.Body.Bytes(), &bootstrapA); err != nil {
		t.Fatalf("v2 workspace-a bootstrap json: %v", err)
	}
	if bootstrapA.WorkspaceID != "workspace-a" || bootstrapA.ClientKind != ClientKindWeb {
		t.Fatalf("v2 workspace-a bootstrap = %+v", bootstrapA)
	}
	if got, ok := findAIAgent(bootstrapA.Agents, created.Agent.AgentID); !ok || got.WorkspaceID != "workspace-a" {
		t.Fatalf("v2 workspace-a created agent = %+v, ok=%v", got, ok)
	}
	if _, ok := findAIAgent(bootstrapA.Agents, "agent-owned-codex"); ok {
		t.Fatalf("v2 workspace-a leaked default workspace agent: %+v", bootstrapA.Agents)
	}

	assignReq := httptest.NewRequest(http.MethodPost, "/v2/client/workspaces/workspace-a/ai-agent/tasks/task-v2/assignment", strings.NewReader(`{"agent_id":"`+created.Agent.AgentID+`"}`))
	assignReq.Header.Set(aiAgentTokenHeader, "owner-token")
	assignResp := httptest.NewRecorder()
	server.ServeHTTP(assignResp, assignReq)
	if assignResp.Code != http.StatusAccepted {
		t.Fatalf("v2 assign status=%d body=%s", assignResp.Code, assignResp.Body.String())
	}
	var assigned AIAgentTaskActionResponse
	if err := json.Unmarshal(assignResp.Body.Bytes(), &assigned); err != nil {
		t.Fatalf("v2 assign json: %v", err)
	}
	if assigned.AgentID != created.Agent.AgentID || assigned.ThreadID == "" {
		t.Fatalf("v2 assign response = %+v", assigned)
	}
	if assigned.ActiveStream == nil ||
		assigned.ActiveStream.Href != "/v2/client/workspaces/workspace-a/ai-agent/events" ||
		assigned.ActiveStream.ThreadID != assigned.ThreadID ||
		assigned.ActiveStream.RunID != assigned.RunID {
		t.Fatalf("v2 assign active_stream = %+v, assigned=%+v", assigned.ActiveStream, assigned)
	}

	pollReq := httptest.NewRequest(http.MethodPost, "/v1/agents/"+created.Agent.AgentID+"/poll", strings.NewReader(`{"daemon_id":"daemon-dev-macbook","device_id":"device-dev-macbook","runtime_id":"runtime-cursor-dev"}`))
	pollReq.Header.Set("Authorization", "Bearer owner-token")
	pollResp := httptest.NewRecorder()
	server.ServeHTTP(pollResp, pollReq)
	if pollResp.Code != http.StatusOK {
		t.Fatalf("v2 assignment poll status=%d body=%s", pollResp.Code, pollResp.Body.String())
	}
	var poll PollResponse
	if err := json.Unmarshal(pollResp.Body.Bytes(), &poll); err != nil {
		t.Fatalf("v2 assignment poll json: %v", err)
	}
	if poll.Action != PollStart || poll.Assignment == nil ||
		poll.Assignment.TaskID != "task-v2" ||
		poll.Assignment.ComponentID != "task-v2" ||
		poll.Assignment.AgentID != created.Agent.AgentID ||
		poll.Assignment.RuntimeProvider != "cursor" ||
		!strings.Contains(poll.Assignment.AgentInstruction, created.Agent.Instruction) ||
		!strings.Contains(poll.Assignment.AgentInstruction, "한국어") ||
		!strings.Contains(poll.Assignment.Prompt, "branch_name: RIID-4800-server-task-context-http-client-assignment-prompt-wiring") {
		if poll.Assignment != nil {
			t.Fatalf("v2 assignment poll action=%s assignment=%+v", poll.Action, *poll.Assignment)
		}
		t.Fatalf("v2 assignment poll = %+v", poll)
	}

	threadsReq := httptest.NewRequest(http.MethodGet, "/v2/client/workspaces/workspace-a/ai-agent/tasks/task-v2/threads", nil)
	threadsReq.Header.Set(aiAgentTokenHeader, "owner-token")
	threadsResp := httptest.NewRecorder()
	server.ServeHTTP(threadsResp, threadsReq)
	if threadsResp.Code != http.StatusOK {
		t.Fatalf("v2 threads status=%d body=%s", threadsResp.Code, threadsResp.Body.String())
	}
	var threads AIAgentTaskThreadCollectionResponse
	if err := json.Unmarshal(threadsResp.Body.Bytes(), &threads); err != nil {
		t.Fatalf("v2 threads json: %v", err)
	}
	if len(threads.Threads) != 1 || threads.Threads[0].AgentID != created.Agent.AgentID {
		t.Fatalf("v2 threads = %+v", threads)
	}
	if threads.ActiveStream == nil || threads.ActiveStream.Href != "/v2/client/workspaces/workspace-a/ai-agent/events" {
		t.Fatalf("v2 active stream = %+v", threads.ActiveStream)
	}

	bootstrapBReq := httptest.NewRequest(http.MethodGet, "/v2/client/workspaces/workspace-b/ai-agent/bootstrap", nil)
	bootstrapBReq.Header.Set(aiAgentTokenHeader, "owner-token")
	bootstrapBResp := httptest.NewRecorder()
	server.ServeHTTP(bootstrapBResp, bootstrapBReq)
	if bootstrapBResp.Code != http.StatusOK {
		t.Fatalf("v2 workspace-b bootstrap status=%d body=%s", bootstrapBResp.Code, bootstrapBResp.Body.String())
	}
	var bootstrapB ClientBootstrapResponse
	if err := json.Unmarshal(bootstrapBResp.Body.Bytes(), &bootstrapB); err != nil {
		t.Fatalf("v2 workspace-b bootstrap json: %v", err)
	}
	if _, ok := findAIAgent(bootstrapB.Agents, created.Agent.AgentID); ok {
		t.Fatalf("v2 workspace-b leaked workspace-a agent: %+v", bootstrapB.Agents)
	}
}
