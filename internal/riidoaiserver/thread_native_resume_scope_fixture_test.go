package riidoaiserver

import (
	"context"
	"testing"
)

type nativeResumeProjectionStore struct {
	*Store
	projection AssignmentProjection
	found      bool
	loadErr    error
}

func (s *nativeResumeProjectionStore) LoadAssignmentProjection(context.Context, string) (AssignmentProjection, bool, error) {
	return s.projection, s.found, s.loadErr
}

func nativeResumeScopeFixture(t *testing.T) (Server, *nativeResumeProjectionStore, AIAgentTaskThreadRecord, AuthorizationResult, AssignRequest) {
	t.Helper()
	client := NewDevelopmentAIAgentClientStore()
	base := NewStoreWithConfig(StoreConfig{AgentRegistry: client})
	t.Cleanup(base.Close)
	previous := Assignment{ID: "selected-source", TaskID: "same-task", AgentID: "agent-owned-codex", WorkspaceID: defaultAIAgentClientWorkspaceID, RuntimeProvider: "codex", ModelID: client.agents["agent-owned-codex"].ModelID, CreatedBy: "user-1", ProviderSessionID: "selected-native", State: AssignmentCompleted}
	source := &nativeResumeProjectionStore{Store: base, projection: AssignmentProjection{Assignment: previous}, found: true}
	server := NewServer(ServerConfig{Assignment: source, AIAgentClient: client})
	thread := AIAgentTaskThreadRecord{ThreadID: "selected-thread", AssignmentID: previous.ID, TaskID: previous.TaskID, AgentID: previous.AgentID, AgentSnapshot: &AIAgentTaskThreadAgentSnapshot{AgentID: previous.AgentID, OwnerPrincipalID: "user-1"}}
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: previous.WorkspaceID}
	req := AssignRequest{AgentID: previous.AgentID, WorkspaceID: previous.WorkspaceID, RuntimeProvider: previous.RuntimeProvider, ModelID: previous.ModelID, CreatedBy: principal.PrincipalID}
	return server, source, thread, principal, req
}
