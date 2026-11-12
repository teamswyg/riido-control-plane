package riidoaiserver

import (
	"encoding/json"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	checkoutPullFallbackPrefix      = "pull_request_fallback:"
	checkoutRecoverySelectionSource = "workspace_retry_selection"
)

type taskCheckoutPullRequest struct {
	repository string
	number     string
	updatedAt  time.Time
}

// Only saved PRs scoped to this task by the authenticated context reader may
// supply a checkout target. In particular, do not inspect task body URLs.
func taskCheckoutPullRequests(records []json.RawMessage) []taskCheckoutPullRequest {
	var pulls []taskCheckoutPullRequest
	for _, raw := range records {
		var pr struct {
			URL       string `json:"htmlUrl"`
			Number    int64  `json:"number"`
			UpdatedAt string `json:"updatedAt"`
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
			updatedAt, _ := time.Parse(time.RFC3339Nano, pr.UpdatedAt)
			pulls = append(pulls, taskCheckoutPullRequest{repository: name, number: number, updatedAt: updatedAt})
		}
	}
	return pulls
}
