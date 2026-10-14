package riidoaiserver

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func membershipAgentIDs(t *testing.T, store AIAgentClientStore, principal AuthorizationResult, taskID string) []string {
	t.Helper()
	profiles, err := store.ListWorkspaceAssignedAgentProfiles(context.Background(), principal)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(profiles)
	if err != nil {
		t.Fatal(err)
	}
	var response struct {
		Participants map[string]struct {
			Agents []struct {
				AgentID string `json:"agent_id"`
			} `json:"agents"`
		} `json:"task_agent_participants"`
	}
	if err := json.Unmarshal(raw, &response); err != nil {
		t.Fatal(err)
	}
	ids := []string{}
	for _, participant := range response.Participants[taskID].Agents {
		ids = append(ids, participant.AgentID)
	}
	return ids
}

func TestAIAgentTaskMembershipSurvivesExecutionAndRestart(t *testing.T) {
	ctx := context.Background()
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
	snapshots := &memoryAIAgentClientSnapshotStore{}
	store, err := OpenPersistentAIAgentClientStore(ctx, NewDevelopmentAIAgentClientStore(), snapshots)
	if err != nil {
		t.Fatal(err)
	}
	taskID := "task-membership"
	for _, agentID := range []string{"agent-owned-codex", "agent-public-openclaw"} {
		if _, err := store.CreateAIAgentTaskAgentAssignment(ctx, principal, taskID, AssignAIAgentTaskRequest{AgentID: agentID}); err != nil {
			t.Fatal(err)
		}
	}
	if got := membershipAgentIDs(t, store, principal, taskID); len(got) != 2 {
		t.Fatalf("joint members = %v, want 2", got)
	}
	if _, err := store.StopAIAgentTaskAgentAssignment(ctx, principal, taskID, "agent-owned-codex", AgentAssignmentActionRequest{}); err != nil {
		t.Fatal(err)
	}
	store.mu.Lock()
	threads := store.taskThreads[taskID]
	for i := range threads {
		if threads[i].AgentID == "agent-public-openclaw" {
			threads[i].WorkStatus = AgentWorkStatusCompleted
			threads[i].AssignmentState = AgentAssignmentStateCompleted
		}
	}
	store.taskThreads[taskID] = threads
	store.mu.Unlock()
	if err := store.saveSnapshot(ctx); err != nil {
		t.Fatal(err)
	}
	reopened, err := OpenPersistentAIAgentClientStore(ctx, NewDevelopmentAIAgentClientStore(), snapshots)
	if err != nil {
		t.Fatal(err)
	}
	if got := membershipAgentIDs(t, reopened, principal, taskID); len(got) != 2 {
		t.Fatalf("stopped/completed persisted members = %v", got)
	}
	if _, err := reopened.DeleteAIAgentTaskAgentAssignment(ctx, principal, taskID, "agent-owned-codex", AgentAssignmentActionRequest{}); err != nil {
		t.Fatal(err)
	}
	reopened, err = OpenPersistentAIAgentClientStore(ctx, NewDevelopmentAIAgentClientStore(), snapshots)
	if err != nil {
		t.Fatal(err)
	}
	got := membershipAgentIDs(t, reopened, principal, taskID)
	if len(got) != 1 || got[0] != "agent-public-openclaw" {
		t.Fatalf("explicit removal = %v", got)
	}
	history, err := reopened.ListAIAgentTaskThreads(ctx, principal, taskID)
	if err != nil || len(history.Threads) < 2 {
		t.Fatalf("removal lost execution history: %+v %v", history, err)
	}
}

func TestAIAgentTaskMembershipKeepsWorkspaceAndVisibilityBoundary(t *testing.T) {
	ctx := context.Background()
	store := NewDevelopmentAIAgentClientStore()
	owner := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
	if _, err := store.CreateAIAgentTaskAgentAssignment(ctx, owner, "task-private-membership", AssignAIAgentTaskRequest{AgentID: "agent-owned-codex"}); err != nil {
		t.Fatal(err)
	}
	outsider := AuthorizationResult{PrincipalID: "other-user", WorkspaceID: defaultAIAgentClientWorkspaceID}
	if got := membershipAgentIDs(t, store, outsider, "task-private-membership"); len(got) != 0 {
		t.Fatalf("private member leaked: %v", got)
	}
	otherWorkspace := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: "other-workspace"}
	if got := membershipAgentIDs(t, store, otherWorkspace, "task-private-membership"); len(got) != 0 {
		t.Fatalf("cross-workspace member leaked: %v", got)
	}
}

func TestAIAgentTaskMembershipSurvivesHistoryEvictionAndDeletedProfile(t *testing.T) {
	ctx := context.Background()
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
	store := NewDevelopmentAIAgentClientStore()
	if _, err := store.CreateAIAgentTaskAgentAssignment(ctx, principal, "task-retained", AssignAIAgentTaskRequest{AgentID: "agent-owned-codex"}); err != nil {
		t.Fatal(err)
	}
	store.mu.Lock()
	delete(store.taskThreads, "task-retained")
	delete(store.agents, "agent-owned-codex")
	store.mu.Unlock()
	if got := membershipAgentIDs(t, store, principal, "task-retained"); len(got) != 1 {
		t.Fatalf("history eviction dropped member: %v", got)
	}
	snapshot, err := store.snapshot(time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	reopened := NewDevelopmentAIAgentClientStore()
	if err := reopened.restoreSnapshot(snapshot); err != nil {
		t.Fatal(err)
	}
	if _, err := reopened.DeleteAIAgentTaskAgentAssignment(ctx, principal, "task-retained", "agent-owned-codex", AgentAssignmentActionRequest{}); err != nil {
		t.Fatal(err)
	}
	// A late execution projection cannot resurrect a removed member.
	reopened.mu.Lock()
	reopened.taskThreads["task-retained"] = []AIAgentTaskThreadRecord{{AgentID: "agent-owned-codex", AssignmentState: AgentAssignmentStateCompleted}}
	reopened.mu.Unlock()
	if got := membershipAgentIDs(t, reopened, principal, "task-retained"); len(got) != 0 {
		t.Fatalf("late history resurrected member: %v", got)
	}
}

func TestSplitSnapshotRoundTripsTaskAgentMembershipTombstones(t *testing.T) {
	ctx := context.Background()
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
	store := NewDevelopmentAIAgentClientStore()
	if _, err := store.CreateAIAgentTaskAgentAssignment(ctx, principal, "task-split", AssignAIAgentTaskRequest{AgentID: "agent-owned-codex"}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.DeleteAIAgentTaskAgentAssignment(ctx, principal, "task-split", "agent-owned-codex", AgentAssignmentActionRequest{}); err != nil {
		t.Fatal(err)
	}
	snapshot, err := store.snapshot(time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	items, _, _, err := encodeSplitDynamoDBAIAgentClientSnapshot(snapshot, nil)
	if err != nil {
		t.Fatal(err)
	}
	decoded, _, err := decodeSplitDynamoDBAIAgentClientSnapshot(items)
	if err != nil {
		t.Fatal(err)
	}
	member := decoded.TaskAgentMemberships["task-split"]["agent-owned-codex"]
	if member.RemovedAt == nil || member.AgentSnapshot == nil {
		t.Fatalf("tombstone/snapshot lost: %+v", member)
	}
	reopened := NewDevelopmentAIAgentClientStore()
	if err := reopened.restoreSnapshot(decoded); err != nil {
		t.Fatal(err)
	}
	if got := membershipAgentIDs(t, reopened, principal, "task-split"); len(got) != 0 {
		t.Fatalf("removed member restored: %v", got)
	}
}
