package riidoaiserver

import (
	"errors"
	"regexp"
	"strings"
)

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
		if validWorkspaceRetryBranch(branch) {
			worktree.BranchName = branch
			worktree.Source = checkoutRecoverySelectionSource
			return nil
		}
	}
	return errors.New(workspaceRetryInputHelp)
}

// Match a branch name, including UTF-8 task branches. Full refs, Git revision
// syntax, and values that could be interpreted as flags are outside this input.
func validWorkspaceRetryBranch(branch string) bool {
	if branch == "" || len(branch) > 1024 || branch == "@" ||
		strings.HasPrefix(branch, "-") || strings.HasPrefix(branch, "refs/") ||
		strings.Contains(branch, "..") || strings.Contains(branch, "@{") ||
		strings.ContainsAny(branch, " ~^:?*[\\") || strings.HasSuffix(branch, ".") {
		return false
	}
	for _, r := range branch {
		if r < 32 || r == 127 {
			return false
		}
	}
	for part := range strings.SplitSeq(branch, "/") {
		if part == "" || strings.HasPrefix(part, ".") || strings.HasSuffix(part, ".lock") {
			return false
		}
	}
	return true
}
