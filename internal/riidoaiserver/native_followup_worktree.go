package riidoaiserver

import (
	"reflect"
	"strings"
)

func sameNativeFollowupWorktree(previous, current *AssignmentWorktree) bool {
	if previous == nil || current == nil {
		return previous == current
	}
	a, b := *previous, *current
	// A new linked PR changes the default for new assignments, not the worktree
	// owned by an existing native conversation. The caller retains previous.
	if strings.HasPrefix(b.Source, checkoutPullFallbackPrefix) && sameCheckoutRepository(&a, &b) {
		return true
	}
	if strings.HasPrefix(a.Source, checkoutPullFallbackPrefix) {
		a.Source = "pull_request"
	}
	if strings.HasPrefix(b.Source, checkoutPullFallbackPrefix) {
		b.Source = "pull_request"
	}
	if b.Source == checkoutRecoverySelectionSource && sameCheckoutRepository(&a, &b) {
		return a.BranchName == b.BranchName
	}
	if a.Source == checkoutRecoverySelectionSource && sameCheckoutRepository(&a, &b) {
		return true
	}
	return reflect.DeepEqual(a, b)
}
