package riidoaiserver

import "testing"

func TestNativeCheckoutAutomaticPullKeepsOnlySameRepositoryBinding(t *testing.T) {
	previous := &AssignmentWorktree{RepositoryFullName: "example/repo", RepositoryURL: "https://github.com/example/repo", BranchName: "task-branch", Source: "pull_request"}
	current := *previous
	current.BranchName, current.Source = "refs/pull/23/head", "pull_request_fallback:23"
	if !sameNativeFollowupWorktree(previous, &current) {
		t.Fatal("automatic PR discovery must retain the existing native worktree")
	}
	for _, field := range []string{"repository", "url", "private"} {
		t.Run(field, func(t *testing.T) {
			changed := current
			switch field {
			case "repository":
				changed.RepositoryFullName = "example/other"
			case "url":
				changed.RepositoryURL = "https://github.com/example/other"
			case "private":
				changed.IsPrivate = true
			}
			if sameNativeFollowupWorktree(previous, &changed) {
				t.Fatal("automatic checkout bypassed repository binding")
			}
		})
	}
}

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
