package riidoaiserver

import "strings"

func assignmentWorktreeFromTaskContext(contextSnapshot AIAgentTaskContext, repository AIAgentTaskContextRepository) *AssignmentWorktree {
	worktree := &AssignmentWorktree{
		RepositoryFullName: safeAIAgentRepositoryFullName(repository.FullName),
		RepositoryURL:      safeAIAgentRepositoryURL(repository.RepositoryURL),
		BranchName:         strings.TrimSpace(contextSnapshot.Component.BranchName),
		IsPrivate:          repository.IsPrivate,
		Source:             strings.TrimSpace(repository.Source),
	}
	if worktree.RepositoryFullName == "" && worktree.RepositoryURL == "" {
		return nil
	}
	keyNumber := strings.TrimSpace(contextSnapshot.Component.KeyNumber)
	taskGeneratedBranch := worktree.BranchName == "" || (keyNumber != "" && strings.HasPrefix(worktree.BranchName, keyNumber+"-"))
	if taskGeneratedBranch {
		pulls := checkoutPullsForRepository(contextSnapshot.linkedPullRequests, worktree)
		if pull, ok := latestCheckoutPull(pulls); ok {
			worktree.BranchName = "refs/pull/" + pull.number + "/head"
			worktree.Source = checkoutPullFallbackPrefix + pull.number
		}
	}
	return worktree
}
