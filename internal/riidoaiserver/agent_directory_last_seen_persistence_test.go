package riidoaiserver

import (
	"context"
	"fmt"
	"testing"
	"time"
)

type directoryHeartbeatCountingStore struct {
	*PersistentAIAgentClientStore
	lists      int
	daemonGets int
}

func (s *directoryHeartbeatCountingStore) ListAIAgentTaskAssignableAgents(ctx context.Context, p AuthorizationResult, task string) (AgentClientListResponse, error) {
	s.lists++
	return s.PersistentAIAgentClientStore.ListAIAgentTaskAssignableAgents(ctx, p, task)
}

func (s *directoryHeartbeatCountingStore) GetAIAgentDaemon(ctx context.Context, p AuthorizationResult, agent string) (DeviceDaemonDetailResponse, error) {
	s.daemonGets++
	return s.PersistentAIAgentClientStore.GetAIAgentDaemon(ctx, p, agent)
}

func TestAgentDirectoryHeartbeatUsesOneSnapshotForListAndShow(t *testing.T) {
	ctx := context.Background()
	base := NewDevelopmentAIAgentClientStore()
	for i := range 80 {
		agent := base.agents["agent-public-openclaw"]
		agent.AgentID, agent.Name = fmt.Sprintf("batch-%03d", i), fmt.Sprintf("Batch %03d", i)
		base.agents[agent.AgentID] = agent
	}
	snapshots := &memoryAIAgentClientSnapshotStore{}
	persistent, err := OpenPersistentAIAgentClientStore(ctx, base, snapshots)
	if err != nil {
		t.Fatal(err)
	}
	persistent.snapshotReloadInterval = 0 // Every public read would reload; no cache can hide N+1 calls.
	store := &directoryHeartbeatCountingStore{PersistentAIAgentClientStore: persistent}
	server := Server{aiAgent: store}
	assignment := Assignment{AgentID: "agent-owned-codex", CreatedBy: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID, TaskID: "task-1"}
	for _, request := range []AgentAssignmentContextRequest{
		{View: AssignmentContextViewAgents, Cursor: "1"},
		{View: AssignmentContextViewAgent, Target: "agent-public-openclaw"},
		{View: AssignmentContextViewAgent, Target: "agent-private-cursor"},
	} {
		beforeLoads, beforeLists := snapshots.loads, store.lists
		data, _, found, err := server.assignmentAgentDirectoryView(ctx, assignment, false, request, time.Now().UTC())
		if err != nil || found != (request.Target != "agent-private-cursor") {
			t.Fatalf("directory found=%v err=%v", found, err)
		}
		if snapshots.loads != beforeLoads+1 || store.lists != beforeLists+1 || store.daemonGets != 0 {
			t.Fatalf("directory performed extra store reads: reloads=%d lists=%d daemonGets=%d", snapshots.loads-beforeLoads, store.lists-beforeLists, store.daemonGets)
		}
		if page, ok := data.(AgentDirectoryPage); ok {
			if len(page.Agents) == 0 || len(page.Agents) > agentDirectoryPageSize {
				t.Fatalf("unexpected bounded page size: %d", len(page.Agents))
			}
			for _, entry := range page.Agents {
				if entry.Runtime.LastSeenAt.IsZero() {
					t.Fatalf("visible agent heartbeat missing: %s", entry.AgentID)
				}
			}
		}
	}
}
