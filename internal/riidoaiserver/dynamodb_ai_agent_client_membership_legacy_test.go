package riidoaiserver

import (
	"context"
	"testing"
	"time"
)

func TestSplitMembershipUncommittedTombstoneUsesLegacyHistory(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
	if _, err := store.CreateAIAgentTaskAgentAssignment(context.Background(), principal, "task-legacy-commit", AssignAIAgentTaskRequest{AgentID: "agent-owned-codex"}); err != nil {
		t.Fatal(err)
	}
	snapshot, err := store.snapshot(time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	removed := time.Now().UTC()
	member := snapshot.TaskAgentMemberships["task-legacy-commit"]["agent-owned-codex"]
	member.RemovedAt = &removed
	snapshot.TaskAgentMemberships["task-legacy-commit"]["agent-owned-codex"] = member
	items, hashes, _, err := encodeSplitDynamoDBAIAgentClientSnapshot(snapshot, nil)
	if err != nil {
		t.Fatal(err)
	}
	delete(hashes, dynamoDBAIAgentClientSnapshotPartTaskAgentMemberships)
	dynamoDBAIAgentClientSnapshotCurrentItem(items)["part_hashes_json"] = map[string]string{"S": mustJSONDynamoDBAIAgentClientSnapshotPartHashes(hashes)}
	decoded, _, err := decodeSplitDynamoDBAIAgentClientSnapshot(items)
	if err != nil {
		t.Fatal(err)
	}
	reopened := NewDevelopmentAIAgentClientStore()
	if err := reopened.restoreSnapshot(decoded); err != nil {
		t.Fatal(err)
	}
	if ids := membershipAgentIDs(t, reopened, principal, "task-legacy-commit"); len(ids) != 1 || ids[0] != "agent-owned-codex" {
		t.Fatalf("uncommitted tombstone overrode committed legacy assignment: %v", ids)
	}
}
