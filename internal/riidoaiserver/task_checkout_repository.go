package riidoaiserver

import "strings"

func checkoutPullsForRepository(pulls []taskCheckoutPullRequest, worktree *AssignmentWorktree) []taskCheckoutPullRequest {
	if worktree == nil {
		return nil
	}
	repo := worktree.RepositoryFullName
	if repo == "" {
		repo = strings.TrimSuffix(strings.TrimPrefix(worktree.RepositoryURL, "https://github.com/"), ".git")
	}
	var matches []taskCheckoutPullRequest
	seen := make(map[string]int)
	for _, pull := range pulls {
		if !strings.EqualFold(pull.repository, repo) {
			continue
		}
		if index, exists := seen[pull.number]; exists {
			if checkoutPullNewer(pull, matches[index]) {
				matches[index] = pull
			}
			continue
		}
		seen[pull.number] = len(matches)
		matches = append(matches, pull)
	}
	return matches
}

func sameCheckoutRepository(a, b *AssignmentWorktree) bool {
	return a != nil && b != nil && strings.EqualFold(a.RepositoryFullName, b.RepositoryFullName) &&
		strings.EqualFold(a.RepositoryURL, b.RepositoryURL) && a.IsPrivate == b.IsPrivate
}
