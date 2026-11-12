package riidoaiserver

import (
	"errors"
	"regexp"
	"strings"
)

var workspaceRetryBranch = regexp.MustCompile(`^[a-zA-Z0-9_][a-zA-Z0-9._/-]{0,199}$`)
var workspaceRetryPull = regexp.MustCompile(`(?i)^PR #[1-9][0-9]{0,9}$`)

func isWorkspaceRetryInstruction(input string) bool {
	return strings.EqualFold(input, "retry") || input == "다시 시도" ||
		strings.HasPrefix(strings.ToLower(input), "branch:") || strings.HasPrefix(input, "브랜치:") ||
		workspaceRetryPull.MatchString(input)
}

func selectWorkspaceRetryCheckout(worktree *AssignmentWorktree, pulls []taskCheckoutPullRequest, input string) error {
	switch {
	case strings.EqualFold(input, "retry"), input == "다시 시도":
		return nil
	case workspaceRetryPull.MatchString(input):
		number := input[4:]
		for _, pull := range checkoutPullsForRepository(pulls, worktree) {
			if pull.number == number {
				worktree.BranchName = "refs/pull/" + number + "/head"
				worktree.Source = checkoutRecoverySelectionSource
				return nil
			}
		}
	case strings.HasPrefix(strings.ToLower(input), "branch:"), strings.HasPrefix(input, "브랜치:"):
		_, branch, _ := strings.Cut(input, ":")
		branch = strings.TrimSpace(branch)
		if workspaceRetryBranch.MatchString(branch) && !strings.Contains(branch, "..") &&
			!strings.Contains(branch, "//") && !strings.HasPrefix(branch, "refs/") &&
			!strings.HasSuffix(branch, "/") && !strings.HasSuffix(branch, ".") {
			worktree.BranchName = branch
			worktree.Source = checkoutRecoverySelectionSource
			return nil
		}
	}
	return errors.New(workspaceRetryInputHelp)
}
