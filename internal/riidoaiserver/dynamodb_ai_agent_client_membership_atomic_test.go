package riidoaiserver

import (
	"context"
	"testing"
	"time"
)

func TestMembershipFailedCommitKeepsPreviousSnapshotReadable(t *testing.T) {
	ctx := context.Background()
	now := fixedSnapshotTestNow()
	f := newSnapshotDynamoDBFixture(t, now, nil, nil)
	defer f.close()
	old := snapshotTestRecord(now)
	old.TaskAgentMemberships = map[string]map[string]AIAgentTaskMembership{"task": {"agent-a": {AgentID: "agent-a", AssignedAt: now}}}
	if err := f.store.SaveAIAgentClientSnapshot(ctx, old); err != nil {
		t.Fatal(err)
	}
	drainSnapshotRequests(f.requests)
	next := old
	next.TaskAgentMemberships = copyTaskAgentMemberships(old.TaskAgentMemberships)
	member := next.TaskAgentMemberships["task"]["agent-a"]
	removed := now.Add(time.Minute)
	member.RemovedAt = &removed
	next.TaskAgentMemberships["task"]["agent-a"] = member
	f.mu.Lock()
	f.failMembershipCommit = true
	f.mu.Unlock()
	if err := f.store.SaveAIAgentClientSnapshot(ctx, next); err == nil {
		t.Fatal("injected transaction failure did not fail save")
	}
	loaded, ok, err := f.store.LoadAIAgentClientSnapshot(ctx)
	if err != nil || !ok {
		t.Fatalf("previous snapshot unreadable: %v %v", ok, err)
	}
	if loaded.TaskAgentMemberships["task"]["agent-a"].RemovedAt != nil {
		t.Fatal("failed commit leaked tombstone")
	}
	f.mu.Lock()
	f.failMembershipCommit = false
	f.mu.Unlock()
	if err := f.store.SaveAIAgentClientSnapshot(ctx, next); err != nil {
		t.Fatal(err)
	}
	loaded, ok, err = f.store.LoadAIAgentClientSnapshot(ctx)
	if err != nil || !ok || loaded.TaskAgentMemberships["task"]["agent-a"].RemovedAt == nil {
		t.Fatalf("retry did not persist removal: %v %v", ok, err)
	}
}

func TestMembershipReadReplacesMixedQueryWithAtomicCommit(t *testing.T) {
	ctx := context.Background()
	now := fixedSnapshotTestNow()
	f := newSnapshotDynamoDBFixture(t, now, nil, nil)
	defer f.close()
	snapshot := snapshotTestRecord(now)
	snapshot.TaskAgentMemberships = map[string]map[string]AIAgentTaskMembership{"task": {"agent-a": {AgentID: "agent-a"}}}
	if err := f.store.SaveAIAgentClientSnapshot(ctx, snapshot); err != nil {
		t.Fatal(err)
	}
	items, _, _, err := encodeSplitDynamoDBAIAgentClientSnapshot(snapshot, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range items {
		if dynamoDBAIAgentClientSnapshotPartName(item) == dynamoDBAIAgentClientSnapshotPartTaskAgentMemberships {
			item["part_hash"] = map[string]string{"S": "mixed-query-hash"}
		}
	}
	if _, _, err := decodeSplitDynamoDBAIAgentClientSnapshot(items); err == nil {
		t.Fatal("mixed query fixture must fail manifest fence")
	}
	credentials := AWSCredentials{AccessKeyID: "AKID", SecretAccessKey: "SECRET"}
	consistent, err := f.store.readMembershipCommit(ctx, items, credentials)
	if err != nil {
		t.Fatal(err)
	}
	decoded, _, err := decodeSplitDynamoDBAIAgentClientSnapshot(consistent)
	if err != nil || decoded.TaskAgentMemberships["task"]["agent-a"].AgentID != "agent-a" {
		t.Fatalf("atomic read failed: %v", err)
	}
}
