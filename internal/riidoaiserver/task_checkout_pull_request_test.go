package riidoaiserver

import (
	"encoding/json"
	"testing"
)

func TestCheckoutPullRequestsRequireSameRepositoryAndSelectDeterministically(t *testing.T) {
	pulls := taskCheckoutPullRequests([]json.RawMessage{
		json.RawMessage(`{"number":3,"htmlUrl":"https://github.com/example/other/pull/3"}`),
		json.RawMessage(`{"number":4,"htmlUrl":"https://github.com/example/repo/pull/5"}`),
		json.RawMessage(`{"number":6,"htmlUrl":"https://github.com.example.invalid/example/repo/pull/6"}`),
		json.RawMessage(`{"number":7,"htmlUrl":"https://github.com/example/repo/pull/7"}`),
		json.RawMessage(`{"number":8,"htmlUrl":"https://github.com/example/repo/pull/8"}`),
	})
	repository := AIAgentTaskContextRepository{FullName: "example/repo", Source: "pull_request"}
	snapshot := AIAgentTaskContext{linkedPullRequests: pulls}
	if worktree := assignmentWorktreeFromTaskContext(snapshot, repository); worktree.BranchName != "refs/pull/8/head" {
		t.Fatalf("equal-date linked PRs must use a deterministic PR-number tie break: %+v", worktree)
	}
	snapshot.linkedPullRequests = pulls[:len(pulls)-1]
	if worktree := assignmentWorktreeFromTaskContext(snapshot, repository); worktree.Source != "pull_request_fallback:7" {
		t.Fatalf("same-repository PR was not selected: %+v", worktree)
	}
	snapshot.Component.BranchName = "explicit-branch"
	if worktree := assignmentWorktreeFromTaskContext(snapshot, repository); worktree.Source != "pull_request" {
		t.Fatalf("explicit branch must have no automatic fallback: %+v", worktree)
	}
}

func TestWorkspaceRetryAllowsKoreanBranchWithoutAcceptingGitRevisionSyntax(t *testing.T) {
	worktree := &AssignmentWorktree{}
	if err := selectWorkspaceRetryCheckout(worktree, nil, "브랜치: 작업/수정"); err != nil || worktree.BranchName != "작업/수정" {
		t.Fatalf("valid UTF-8 branch was rejected: %v", err)
	}
	for _, invalid := range []string{"-h", "refs/pull/3/head", "branch@{1}", "staging~1", "a//b", "foo/.bar", "foo/bar.lock"} {
		if err := selectWorkspaceRetryCheckout(worktree, nil, "branch: "+invalid); err == nil {
			t.Fatalf("invalid branch %q was accepted", invalid)
		}
	}
}
