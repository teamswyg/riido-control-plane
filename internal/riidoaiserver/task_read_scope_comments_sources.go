package riidoaiserver

import "encoding/json"

func readScopeCommentSources(page *taskReadCommentPage, resource AssignmentScopedResource) ([]AssignmentContextSource, []string, error) {
	sources := []AssignmentContextSource{readScopeSource(resource, "comments", page.Version)}
	var replies []string
	seen := map[string]bool{}
	for _, raw := range page.Items {
		var comment struct {
			ID         string  `json:"id"`
			Body       *string `json:"body"`
			UpdatedAt  string  `json:"updatedAt"`
			ReplyCount *int    `json:"replyCount"`
		}
		if json.Unmarshal(raw, &comment) != nil || comment.ID == "" || comment.Body == nil || comment.ReplyCount == nil || *comment.ReplyCount < 0 || seen[comment.ID] {
			return nil, nil, taskContextReadError{"invalid"}
		}
		seen[comment.ID] = true
		sources = append(sources, AssignmentContextSource{Kind: "comment", ID: comment.ID, Part: "body", Version: assignmentContextVersion(*comment.Body), UpdatedAt: normalizedTaskContextTime(comment.UpdatedAt)})
		if *comment.ReplyCount > 0 {
			replies = append(replies, comment.ID)
		}
	}
	return sources, replies, nil
}
