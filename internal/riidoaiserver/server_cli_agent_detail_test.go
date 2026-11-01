package riidoaiserver

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestCLIAgentReadDetailUsesCurrentTitleAndHidesExistence(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	a := authorizationFunc(func(_ context.Context, _ string, req AuthorizationRequest) (AuthorizationResult, error) {
		return AuthorizationResult{PrincipalID: "user-3", WorkspaceID: req.WorkspaceID}, nil
	})
	s := NewServer(ServerConfig{AIAgentClient: store, CLIAgentReadAuthorizer: a})
	hidden := cliAgentDirectoryRecorder(s, "/agent-private-cursor")
	absent := cliAgentDirectoryRecorder(s, "/agent-absent")
	if hidden.Code != http.StatusForbidden || hidden.Code != absent.Code || hidden.Body.String() != absent.Body.String() {
		t.Fatal("hidden and absent directory entries differ")
	}
	agent := store.agents["agent-public-openclaw"]
	agent.Name = "Updated public role"
	store.agents[agent.AgentID] = agent
	w := cliAgentDirectoryRecorder(s, "/agent-public-openclaw")
	var detail AgentDirectoryDetail
	if w.Code != http.StatusOK || json.Unmarshal(w.Body.Bytes(), &detail) != nil || detail.Agent.Name != agent.Name || detail.PolicyChecked {
		t.Fatal("current public title lost")
	}
}
