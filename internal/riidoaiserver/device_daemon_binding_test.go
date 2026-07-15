package riidoaiserver

import (
	"context"
	"testing"
)

// A physical machine serves many workspaces under one device credential. Daemon
// bindings must include agents from every connected workspace; otherwise an
// assignment from another workspace can be stuck "queued" forever.
func TestDaemonAgentBindingsIncludeAgentsFromOtherConnectedWorkspaces(t *testing.T) {
	ctx := context.Background()
	store := NewDevelopmentAIAgentClientStore()
	const machine = "machine-binding-multi"
	const wsA = "workspace-a"
	const wsB = "workspace-b"
	const owner = "owner-user"
	runtimeID := machine + ":claude"

	enrollPrincipal := AuthorizationResult{PrincipalID: owner, WorkspaceID: wsA}
	enroll := enrollTestDevice(t, ctx, store, enrollPrincipal, wsA, machine, "Multi Mac")
	syncTestRuntime(t, ctx, store, enrollPrincipal, machine, enroll.DeviceID, runtimeID)
	if _, err := store.ConnectAIAgentDevice(ctx,
		AuthorizationResult{PrincipalID: owner, WorkspaceID: wsB}, machine); err != nil {
		t.Fatalf("connect wsB: %v", err)
	}

	created, err := store.createAIAgent(ctx,
		AuthorizationResult{PrincipalID: owner, WorkspaceID: wsB},
		CreateAgentConfigurationRequest{
			Name:       "B Agent",
			RuntimeID:  runtimeID,
			Visibility: AgentVisibilityPrivate,
		}, "")
	if err != nil {
		t.Fatalf("create agent in wsB: %v", err)
	}
	bindings, err := store.ListAIAgentDaemonAgentBindings(ctx,
		AuthorizationResult{PrincipalID: owner}, enroll.DeviceID)
	if err != nil {
		t.Fatalf("list bindings: %v", err)
	}

	var found *AgentRuntimeBinding
	for i := range bindings.Bindings {
		if bindings.Bindings[i].AgentID == created.Agent.AgentID {
			found = &bindings.Bindings[i]
			break
		}
	}
	if found == nil {
		t.Fatalf("connected workspace agent missing from bindings: %+v", bindings.Bindings)
	}
	if found.RuntimeID != runtimeID || found.DeviceID != enroll.DeviceID {
		t.Fatalf("binding has wrong runtime/device: %+v", *found)
	}
}

func TestDaemonAgentBindingsIncludeOwnerRuntimeAgentWithoutWorkspaceConnection(t *testing.T) {
	ctx := context.Background()
	store := NewDevelopmentAIAgentClientStore()
	const machine = "machine-binding-owner-cross-workspace"
	const enrolledWorkspace = "workspace-enrolled"
	const agentWorkspace = "workspace-agent"
	const owner = "owner-user"
	runtimeID := machine + ":codex"

	enrollPrincipal := AuthorizationResult{PrincipalID: owner, WorkspaceID: enrolledWorkspace}
	enroll := enrollTestDevice(t, ctx, store, enrollPrincipal, enrolledWorkspace, machine, "Owner Mac")
	syncTestRuntime(t, ctx, store, enrollPrincipal, machine, enroll.DeviceID, runtimeID)

	created, err := store.createAIAgent(ctx,
		AuthorizationResult{PrincipalID: owner, WorkspaceID: agentWorkspace},
		CreateAgentConfigurationRequest{
			Name:       "Cross Workspace Agent",
			RuntimeID:  runtimeID,
			Visibility: AgentVisibilityPrivate,
		}, "")
	if err != nil {
		t.Fatalf("create agent in unconnected workspace: %v", err)
	}

	bindings, err := store.ListAIAgentDaemonAgentBindings(ctx,
		AuthorizationResult{PrincipalID: owner}, enroll.DeviceID)
	if err != nil {
		t.Fatalf("list bindings: %v", err)
	}
	for _, binding := range bindings.Bindings {
		if binding.AgentID == created.Agent.AgentID {
			if binding.RuntimeID != runtimeID || binding.DeviceID != enroll.DeviceID {
				t.Fatalf("binding has wrong runtime/device: %+v", binding)
			}
			return
		}
	}
	t.Fatalf("owner runtime agent missing from bindings: %+v", bindings.Bindings)
}
