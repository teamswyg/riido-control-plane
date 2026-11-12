package riidoaiserver

import (
	"encoding/json"
	"net/url"
	"strconv"
	"strings"
)

const (
	checkoutPullFallbackPrefix      = "pull_request_fallback:"
	checkoutRecoverySelectionSource = "workspace_retry_selection"
)

type taskCheckoutPullRequest struct {
	repository string
	number     string
}

// Only saved PRs scoped to this task by the authenticated context reader may
// supply a checkout target. In particular, do not inspect task body URLs.
func taskCheckoutPullRequests(records []json.RawMessage) []taskCheckoutPullRequest {
	var pulls []taskCheckoutPullRequest
	for _, raw := range records {
		var pr struct {
			URL    string `json:"htmlUrl"`
			Number int64  `json:"number"`
		}
		if json.Unmarshal(raw, &pr) != nil || pr.Number < 1 || pr.Number > 9999999999 {
			continue
		}
		number := strconv.FormatInt(pr.Number, 10)
		parsed, err := url.Parse(pr.URL)
		if err != nil || parsed.Scheme != "https" || !strings.EqualFold(parsed.Host, "github.com") || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
			continue
		}
		suffix := "/pull/" + number
		if !strings.HasSuffix(parsed.EscapedPath(), suffix) {
			continue
		}
		name := safeAIAgentRepositoryFullName(strings.TrimPrefix(strings.TrimSuffix(parsed.EscapedPath(), suffix), "/"))
		if name != "" {
			pulls = append(pulls, taskCheckoutPullRequest{repository: name, number: number})
		}
	}
	return pulls
}

func checkoutPullsForRepository(pulls []taskCheckoutPullRequest, worktree *AssignmentWorktree) []taskCheckoutPullRequest {
	if worktree == nil {
		return nil
	}
	repo := worktree.RepositoryFullName
	if repo == "" {
		repo = strings.TrimSuffix(strings.TrimPrefix(worktree.RepositoryURL, "https://github.com/"), ".git")
	}
	var matches []taskCheckoutPullRequest
	seen := make(map[string]bool)
	for _, pull := range pulls {
		if strings.EqualFold(pull.repository, repo) && !seen[pull.number] {
			matches = append(matches, pull)
			seen[pull.number] = true
		}
	}
	return matches
}

func sameCheckoutRepository(a, b *AssignmentWorktree) bool {
	return a != nil && b != nil && strings.EqualFold(a.RepositoryFullName, b.RepositoryFullName) &&
		strings.EqualFold(a.RepositoryURL, b.RepositoryURL) && a.IsPrivate == b.IsPrivate
}
