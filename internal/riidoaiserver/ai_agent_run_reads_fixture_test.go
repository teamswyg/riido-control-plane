package riidoaiserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func runReadsTestServer(t *testing.T, reads []AssignmentContextRead) (Server, *runReadsTestContext, *runReadsTestAuthorizer) {
	t.Helper()
	store := NewDevelopmentAIAgentClientStore()
	store.taskThreads["task"] = []AIAgentTaskThreadRecord{{TaskID: "task", RunID: "run", ThreadID: "thread", AssignmentID: "assignment", AgentID: "agent-owned-codex"}}
	contexts := NewMemoryAssignmentContextStore()
	for i := range reads {
		reads[i].AssignmentID, reads[i].TaskID, reads[i].AgentID = "assignment", "task", "agent-owned-codex"
		reads[i].RequesterPrincipalID, reads[i].RequestID, reads[i].TriggerID = "secret-requester", "secret-request", "secret-trigger"
		reads[i].Status, reads[i].Via, reads[i].ServedAt = "served", "cli", time.Date(2026, 10, 26, 0, 0, i, 0, time.UTC)
	}
	if err := contexts.AppendAssignmentContextReads(context.Background(), reads); err != nil {
		t.Fatal(err)
	}
	reader := &runReadsTestContext{value: AIAgentTaskContext{Component: AIAgentTaskContextComponent{ID: "task"}, Document: AIAgentTaskContextDocument{Content: "current body"}}}
	authorizer := &runReadsTestAuthorizer{}
	return NewServer(ServerConfig{AIAgentClient: store, AssignmentContext: contexts, TaskContext: reader, Authorizer: authorizer}), reader, authorizer
}

func runReadsRequest(server Server, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.Header.Set("Authorization", "Bearer viewer-token")
	req.Header.Set(aiAgentTokenHeader, "unrelated-agent-token")
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, req)
	return response
}
