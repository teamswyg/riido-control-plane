package riidoaiserver

import (
	"context"
	"strings"
	"testing"

	"github.com/teamswyg/riido-contracts/hostintegration"
	"github.com/teamswyg/riido-contracts/provider/capability"
)

func TestStoreActorRejectsStoreUnsafeProviderStatus(t *testing.T) {
	ctx := context.Background()
	store := NewStore()
	defer store.Close()

	if _, err := store.SyncProviderStatus(ctx, "agent-1", ProviderStatusSyncRequest{
		DaemonID:            "daemon-1",
		RuntimeID:           "runtime-1",
		DistributionChannel: hostintegration.DistributionChannelDevLocal,
		Providers: []ProviderStatusRecord{{
			ProviderKind:  capability.ProviderKind("codex"),
			RoutingStatus: hostintegration.ProviderRoutingLoginRequired,
		}},
	}); err != nil {
		t.Fatalf("SyncProviderStatus: %v", err)
	}

	_, err := store.AssignTask(ctx, "task-a", AssignRequest{
		ComponentID: "component-a", AgentID: "agent-1", RuntimeProvider: "codex", Prompt: "blocked",
	})
	if err == nil || !strings.Contains(err.Error(), "provider codex cannot be assigned") {
		t.Fatalf("AssignTask blocked err=%v", err)
	}
}

func TestStoreActorValidatesAgentRuntimeBinding(t *testing.T) {
	ctx := context.Background()
	registry, err := NewStaticAgentRegistry([]AgentRuntimeBinding{{
		AgentID: "agent-1", DaemonID: "daemon-1", RuntimeID: "runtime-1", RuntimeProvider: "codex",
	}})
	if err != nil {
		t.Fatalf("NewStaticAgentRegistry: %v", err)
	}
	store := NewStoreWithConfig(StoreConfig{AgentRegistry: registry})
	defer store.Close()

	_, err = store.AssignTask(ctx, "task-a", AssignRequest{
		ComponentID: "component-a", AgentID: "agent-1", RuntimeProvider: "claude", Prompt: "wrong provider",
	})
	if err == nil || !strings.Contains(err.Error(), "bound to runtime_provider codex") {
		t.Fatalf("AssignTask wrong provider err=%v", err)
	}
	mustAssignActorTask(t, store, ctx, "task-a", AssignRequest{
		ComponentID: "component-a", AgentID: "agent-1", RuntimeProvider: "codex", Prompt: "right provider",
	})
	_, err = store.PollAgent(ctx, "agent-1", PollRequest{DaemonID: "other", RuntimeID: "runtime-1"})
	if err == nil || !strings.Contains(err.Error(), "bound to daemon_id daemon-1") {
		t.Fatalf("PollAgent wrong daemon err=%v", err)
	}
}
