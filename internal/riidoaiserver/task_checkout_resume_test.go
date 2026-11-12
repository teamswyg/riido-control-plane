package riidoaiserver

import "testing"

func TestNativeCheckoutSelectionCannotMoveExistingSession(t *testing.T) {
	for _, source := range []string{checkoutRecoverySelectionSource, "pull_request_fallback:19"} {
		previous := &AssignmentWorktree{RepositoryFullName: "example/repo", BranchName: "refs/pull/19/head", Source: source}
		current := *previous
		current.Source = checkoutRecoverySelectionSource
		if !sameNativeFollowupWorktree(previous, &current) {
			t.Fatal("explicit selection of the existing branch must remain resumable")
		}
		current.BranchName = "staging"
		if sameNativeFollowupWorktree(previous, &current) {
			t.Fatal("explicit selection of another branch must not silently reuse the old session")
		}
	}
	previous := &AssignmentWorktree{RepositoryFullName: "example/repo", BranchName: "staging", Source: checkoutRecoverySelectionSource}
	automatic := &AssignmentWorktree{RepositoryFullName: "example/repo", BranchName: "refs/pull/19/head", Source: "pull_request_fallback:19"}
	if !sameNativeFollowupWorktree(previous, automatic) {
		t.Fatal("ordinary replies must retain their existing explicitly selected worktree")
	}
}
