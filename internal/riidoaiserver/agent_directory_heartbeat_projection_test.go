package riidoaiserver

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestAgentDirectoryLegacyHeartbeatProjectionStaysPrivate(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	heartbeat := time.Now().UTC().Add(-time.Hour).Truncate(time.Second)
	setDirectoryHeartbeat(t, store, "agent-public-openclaw", heartbeat, false)
	list, err := store.ListAIAgentTaskAssignableAgents(context.Background(), AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}, "task-1")
	if err != nil {
		t.Fatal(err)
	}
	if !list.agentDaemonLastSeenAt["agent-public-openclaw"].Equal(heartbeat) {
		t.Fatal("legacy runtime must preserve the device heartbeat")
	}
	visible := make(map[string]bool, len(list.Agents))
	for _, agent := range list.Agents {
		visible[agent.AgentID] = true
	}
	for id := range list.agentDaemonLastSeenAt {
		if !visible[id] {
			t.Fatalf("heartbeat projected an invisible agent: %s", id)
		}
	}
	if _, leaked := list.agentDaemonLastSeenAt["agent-private-cursor"]; leaked {
		t.Fatal("private agent included in heartbeat sidecar")
	}
	body, err := json.Marshal(list)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(body), "last_seen") || strings.Contains(string(body), "agentDaemonLastSeen") {
		t.Fatal("ordinary assignable-list API exposed private sidecar")
	}
}

func TestAgentDirectoryHeartbeatDoesNotUseAnotherDaemonIdentity(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	setDirectoryHeartbeat(t, store, "agent-public-openclaw", time.Now().UTC(), true)
	store.mu.Lock()
	for i := range store.devices {
		for j := range store.devices[i].Runtimes {
			runtime := &store.devices[i].Runtimes[j]
			if runtime.RuntimeID == store.agents["agent-public-openclaw"].RuntimeID {
				runtime.DaemonProfile = "unobserved-profile"
			}
		}
	}
	store.mu.Unlock()
	list, err := store.ListAIAgentTaskAssignableAgents(context.Background(), AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}, "task-1")
	if err != nil {
		t.Fatal(err)
	}
	if _, invented := list.agentDaemonLastSeenAt["agent-public-openclaw"]; invented {
		t.Fatal("another daemon or aggregate device heartbeat cannot replace the explicit identity")
	}
}
