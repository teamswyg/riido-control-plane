package riidoaiserver

import (
	"errors"
	"fmt"
	"strings"
)

var errTaskCheckoutSelection = errors.New("invalid task checkout selection")

func latestCheckoutPull(pulls []taskCheckoutPullRequest) (taskCheckoutPullRequest, bool) {
	var latest taskCheckoutPullRequest
	for _, pull := range pulls {
		if latest.number == "" || checkoutPullNewer(pull, latest) {
			latest = pull
		}
	}
	return latest, latest.number != ""
}

func checkoutPullNewer(a, b taskCheckoutPullRequest) bool {
	if !a.updatedAt.Equal(b.updatedAt) {
		return a.updatedAt.After(b.updatedAt)
	}
	// Validated decimal PR numbers have no leading zeroes.
	return len(a.number) > len(b.number) || (len(a.number) == len(b.number) && a.number > b.number)
}

func applyTaskCheckoutInstruction(worktree *AssignmentWorktree, pulls []taskCheckoutPullRequest, instruction string) error {
	firstLine, _, _ := strings.Cut(taskMessageContextText(instruction, taskDocumentContextLimit), "\n")
	input := strings.TrimSpace(firstLine)
	if !isWorkspaceRetryInstruction(input) || strings.EqualFold(input, "retry") || input == "다시 시도" {
		return nil
	}
	if worktree == nil {
		return fmt.Errorf("%w: connect a repository before selecting a branch or PR", errTaskCheckoutSelection)
	}
	if err := selectWorkspaceRetryCheckout(worktree, pulls, input); err != nil {
		return fmt.Errorf("%w: %w", errTaskCheckoutSelection, err)
	}
	return nil
}
