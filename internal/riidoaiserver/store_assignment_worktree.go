package riidoaiserver

import "strings"

func normalizeAssignmentWorktree(worktree *AssignmentWorktree) *AssignmentWorktree {
	if worktree == nil {
		return nil
	}
	out := &AssignmentWorktree{
		RepositoryFullName: safeAIAgentRepositoryFullName(worktree.RepositoryFullName),
		RepositoryURL:      safeAIAgentRepositoryURL(worktree.RepositoryURL),
		BranchName:         strings.TrimSpace(worktree.BranchName),
		IsPrivate:          worktree.IsPrivate,
		Source:             strings.TrimSpace(worktree.Source),
	}
	if out.RepositoryFullName == "" && out.RepositoryURL == "" {
		return nil
	}
	return out
}

func cloneAssignmentWorktree(worktree *AssignmentWorktree) *AssignmentWorktree {
	if worktree == nil {
		return nil
	}
	out := *worktree
	return &out
}
