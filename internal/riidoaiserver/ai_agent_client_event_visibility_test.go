package riidoaiserver

import "testing"

func TestClientEventVisibleToLockedFallsBackToThreadSnapshot(t *testing.T) {
	store := &DevelopmentAIAgentClientStore{
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
	viewer := AuthorizationResult{PrincipalID: "viewer-a", WorkspaceID: "workspace-a"}
	event := ClientStreamEvent{Payload: AgentThreadProgressEvent{
		AgentID: "agent-deleted", TaskID: "task-a", ThreadID: "thread-a",
	}}
	if !clientEventVisibleToLocked(store, viewer, event) {
		t.Fatal("deleted public agent thread event should remain visible")
	}
}

func TestClientEventVisibleToLockedUnknownAgentRequiresAdmin(t *testing.T) {
	store := &DevelopmentAIAgentClientStore{workspaceID: "workspace-a", agents: map[string]AgentClientRecord{}}
	event := ClientStreamEvent{Payload: AgentWorkStatusChangedEvent{AgentID: "agent-missing"}}
	viewer := AuthorizationResult{PrincipalID: "viewer-a", WorkspaceID: "workspace-a"}
	if clientEventVisibleToLocked(store, viewer, event) {
		t.Fatal("unknown agent event without thread ref should be hidden from non-admin")
	}
	admin := AuthorizationResult{
		PrincipalID: "admin-a", WorkspaceID: "workspace-a", Roles: []AgentCatalogRole{AgentCatalogRoleAdmin},
	}
	if !clientEventVisibleToLocked(store, admin, event) {
		t.Fatal("admin should see unknown agent event without thread ref")
	}
}

func TestClientEventForSubscriberLockedSelectsFanoutEvent(t *testing.T) {
	store := &DevelopmentAIAgentClientStore{workspaceID: "workspace-a", agents: map[string]AgentClientRecord{}}
	viewer := AuthorizationResult{PrincipalID: "viewer-a", WorkspaceID: "workspace-a"}
	original := ClientStreamEvent{Seq: 1, Payload: struct{ Name string }{"original"}}
	fanout := ClientStreamEvent{Seq: 2, Payload: struct{ Name string }{"fanout"}}
	if got, ok := clientEventForSubscriberLocked(store, viewer, original, fanout, false); !ok || got.Seq != 1 {
		t.Fatalf("non-fanout got (%+v,%v), want original", got, ok)
	}
	if got, ok := clientEventForSubscriberLocked(store, viewer, original, fanout, true); !ok || got.Seq != 2 {
		t.Fatalf("fanout got (%+v,%v), want fanout", got, ok)
	}
	invisible := ClientStreamEvent{Payload: AgentWorkStatusChangedEvent{AgentID: "missing"}}
	if _, ok := clientEventForSubscriberLocked(store, viewer, original, invisible, true); ok {
		t.Fatal("invisible fanout event should be hidden")
	}
}
