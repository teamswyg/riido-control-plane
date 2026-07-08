package riidoaiserver

import "testing"

func TestClientEventForPrincipalLocalizesWorkStatusAuthFailure(t *testing.T) {
	t.Parallel()
	store := visibleDeletedAgentThreadStore()
	raw := "Failed to authenticate. API Error: 401 Invalid authentication credentials"
	event := ClientStreamEvent{Payload: AgentWorkStatusChangedEvent{
		AgentID: "agent-deleted", TaskID: "task-a", ThreadID: "thread-a", ResultMessage: raw,
		FailureDiagnostics: &AIAgentTaskThreadFailureDiagnostics{Message: raw},
	}}
	visible, ok := clientEventForPrincipalLocked(store, visibleViewer(), event)
	if !ok {
		t.Fatal("expected visible event")
	}
	status := visible.Payload.(AgentWorkStatusChangedEvent)
	if status.ResultMessage != clientMessageProviderAuthFailed {
		t.Fatalf("result_message = %q, want %q", status.ResultMessage, clientMessageProviderAuthFailed)
	}
	if status.FailureDiagnostics.Message != clientMessageProviderAuthFailed {
		t.Fatalf("diagnostic message = %q, want %q", status.FailureDiagnostics.Message, clientMessageProviderAuthFailed)
	}
}

func TestClientEventForLiveFanoutLocalizesWorkStatusAuthFailure(t *testing.T) {
	t.Parallel()
	raw := "Failed to authenticate. API Error: 401 Invalid authentication credentials"
	event := ClientStreamEvent{Payload: AgentWorkStatusChangedEvent{ResultMessage: raw}}
	visible, progressFanout := clientEventForLiveFanout(event)
	if progressFanout {
		t.Fatal("work status event should not be progress fanout")
	}
	status := visible.Payload.(AgentWorkStatusChangedEvent)
	if status.ResultMessage != clientMessageProviderAuthFailed {
		t.Fatalf("result_message = %q, want %q", status.ResultMessage, clientMessageProviderAuthFailed)
	}
}

func visibleDeletedAgentThreadStore() *DevelopmentAIAgentClientStore {
	return &DevelopmentAIAgentClientStore{
		workspaceID: "workspace-a",
		agents:      map[string]AgentClientRecord{},
		taskThreads: map[string][]AIAgentTaskThreadRecord{"task-a": {{
			TaskID: "task-a", ThreadID: "thread-a", AgentID: "agent-deleted",
			AgentSnapshot: &AIAgentTaskThreadAgentSnapshot{
				AgentID: "agent-deleted", WorkspaceID: "workspace-a",
				OwnerPrincipalID: "owner-a", Visibility: AgentVisibilityPublic,
			},
		}}},
	}
}

func visibleViewer() AuthorizationResult {
	return AuthorizationResult{PrincipalID: "viewer-a", WorkspaceID: "workspace-a"}
}
