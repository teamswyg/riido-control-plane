package riidoaiserver

import (
	"context"
	"testing"
)

func TestConnectedAccountAgentRoutesThroughExistingDaemonWithoutReplacingDevice(t *testing.T) {
	ctx := context.Background()
	store := NewDevelopmentAIAgentClientStore()
	const (
		machineID      = "machine-cross-account"
		ownerPrincipal = "owner-account"
		ownerWorkspace = "workspace-owner"
		guestPrincipal = "guest-account"
		guestWorkspace = "workspace-guest"
	)
	owner := AuthorizationResult{PrincipalID: ownerPrincipal, WorkspaceID: ownerWorkspace}
	guest := AuthorizationResult{PrincipalID: guestPrincipal, WorkspaceID: guestWorkspace}

	enrollment := enrollTestDevice(t, ctx, store, owner, ownerWorkspace, machineID, "Shared Mac")
	runtimeID := enrollment.DeviceID + ":claude"
	syncTestRuntime(t, ctx, store, owner, machineID, enrollment.DeviceID, runtimeID)

	connected, err := store.ConnectAIAgentDevice(ctx, guest, machineID)
	if err != nil {
		t.Fatalf("connect guest account: %v", err)
	}
	if connected.DeviceID != enrollment.DeviceID {
		t.Fatalf("connected device = %q, want stable %q", connected.DeviceID, enrollment.DeviceID)
	}

	created, err := store.createAIAgent(ctx, guest, CreateAgentConfigurationRequest{
		Name:       "Guest Claude",
		RuntimeID:  runtimeID,
		Visibility: AgentVisibilityPrivate,
	}, "")
	if err != nil {
		t.Fatalf("create guest agent: %v", err)
	}
	bindings, err := store.ListAIAgentDaemonAgentBindings(ctx, owner, enrollment.DeviceID)
	if err != nil {
		t.Fatalf("list existing daemon bindings: %v", err)
	}
	if !bindingListContainsAgent(bindings.Bindings, created.Agent.AgentID, runtimeID) {
		t.Fatalf("guest agent missing from existing daemon bindings: %+v", bindings.Bindings)
	}
	if bindings.ConnectedPrincipalCount != 2 || bindings.ConnectionRevision == "" {
		t.Fatalf("connection summary = %+v, want two principals and a revision", bindings)
	}

	reconnected, err := store.ConnectAIAgentDevice(ctx, owner, machineID)
	if err != nil {
		t.Fatalf("reconnect original account: %v", err)
	}
	if reconnected.DeviceID != enrollment.DeviceID || len(reconnected.Runtimes) != 1 || reconnected.Runtimes[0].RuntimeID != runtimeID {
		t.Fatalf("original daemon data changed after account round trip: %+v", reconnected)
	}
	if reconnected.OwnerPrincipalID != ownerPrincipal {
		t.Fatalf("device owner changed during account connection: %q", reconnected.OwnerPrincipalID)
	}
	afterReconnect, err := store.ListAIAgentDaemonAgentBindings(ctx, owner, enrollment.DeviceID)
	if err != nil {
		t.Fatalf("list bindings after reconnect: %v", err)
	}
	if afterReconnect.ConnectionRevision != bindings.ConnectionRevision || afterReconnect.ConnectedPrincipalCount != 2 {
		t.Fatalf("reconnect changed semantic connection summary: before=%+v after=%+v", bindings, afterReconnect)
	}
}

func bindingListContainsAgent(bindings []AgentRuntimeBinding, agentID, runtimeID string) bool {
	for _, binding := range bindings {
		if binding.AgentID == agentID && binding.RuntimeID == runtimeID {
			return true
		}
	}
	return false
}
