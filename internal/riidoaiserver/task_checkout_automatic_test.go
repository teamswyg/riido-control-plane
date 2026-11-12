package riidoaiserver

import (
	"encoding/json"
	"testing"
)

func TestAssignmentCheckoutUsesMostRecentlyUpdatedLinkedPull(t *testing.T) {
	pulls := []json.RawMessage{
		json.RawMessage(`{"number":21,"state":"OPEN","updatedAt":"2026-11-11T02:00:00Z","htmlUrl":"https://github.com/example/repo/pull/21"}`),
		json.RawMessage(`{"number":19,"state":"MERGED","updatedAt":"2026-11-12T01:00:00Z","htmlUrl":"https://github.com/example/repo/pull/19"}`),
		json.RawMessage(`{"number":99,"updatedAt":"2026-11-13T01:00:00Z","htmlUrl":"https://github.com/example/other/pull/99"}`),
	}
	for _, records := range [][]json.RawMessage{pulls, {pulls[2], pulls[1], pulls[0]}, pulls[1:2]} {
		snapshot := AIAgentTaskContext{
			Component:          AIAgentTaskContextComponent{KeyNumber: "RIID-6119", BranchName: "RIID-6119-generated-title"},
			linkedPullRequests: taskCheckoutPullRequests(records),
		}
		for _, source := range []string{"pull_request", "component"} {
			worktree := assignmentWorktreeFromTaskContext(snapshot, AIAgentTaskContextRepository{FullName: "example/repo", Source: source})
			if worktree.BranchName != "refs/pull/19/head" {
				t.Fatalf("generated task branch must use latest linked PR head, got %+v", worktree)
			}
		}
	}
}

func TestAssignmentCheckoutPreservesExplicitBranchAndIncompleteEvidence(t *testing.T) {
	repository := AIAgentTaskContextRepository{FullName: "example/repo", Source: "pull_request"}
	snapshot := AIAgentTaskContext{
		Component:          AIAgentTaskContextComponent{KeyNumber: "RIID-6119", BranchName: "staging"},
		linkedPullRequests: taskCheckoutPullRequests([]json.RawMessage{json.RawMessage(`{"number":19,"htmlUrl":"https://github.com/example/repo/pull/19"}`)}),
	}
	if worktree := assignmentWorktreeFromTaskContext(snapshot, repository); worktree.BranchName != "staging" {
		t.Fatalf("explicit configured branch changed: %+v", worktree)
	}
	snapshot.Component.BranchName = "RIID-6119-generated-title"
	snapshot.linkedPullRequests = nil
	if worktree := assignmentWorktreeFromTaskContext(snapshot, repository); worktree.BranchName != snapshot.Component.BranchName {
		t.Fatalf("missing authorized PR evidence must not invent a checkout: %+v", worktree)
	}
}
