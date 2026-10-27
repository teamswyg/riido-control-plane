package riidoaiserver

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
)

type privateReplyMetadata struct {
	ID        string `json:"id"`
	Status    string `json:"status"`
	UpdatedAt string `json:"updatedAt"`
}

// Only roots in the current same-task viewer response can authorize a lookup.
// One bounded alias query avoids per-comment HTTP calls; any batch error keeps
// every reply hidden. Retained snapshots and trigger IDs never grant access.
func (c *AIAgentPrivateTaskContextClient) readCurrentReplyMetadata(ctx context.Context, team, bearer string, comments []AIAgentTaskContextComment) []AIAgentTaskContextComment {
	parents := currentReplyParents(comments)
	if len(parents) == 0 || ctx.Err() != nil {
		return nil
	}
	var query strings.Builder
	query.WriteString("query RiidoViewerCommentReplies($team:String!){")
	for i, parent := range parents {
		id, err := json.Marshal(parent)
		if err != nil {
			return nil
		}
		query.WriteString("r" + strconv.Itoa(i) + ":getCommentReplies(commentId:" + string(id) + ",teamId:$team){id status updatedAt}")
	}
	query.WriteString("}")
	var response map[string][]privateReplyMetadata
	if c.contextGraphQL(ctx, bearer, query.String(), team, "", &response) != nil {
		return nil
	}
	for i := range parents {
		if values, ok := response["r"+strconv.Itoa(i)]; !ok || values == nil {
			return nil
		}
	}
	return currentReplyMetadata(parents, comments, response)
}

func currentReplyParents(comments []AIAgentTaskContextComment) []string {
	seen := map[string]bool{}
	var parents []string
	for _, comment := range comments {
		id := strings.TrimSpace(comment.ID)
		if id == "" || seen[id] || comment.ParentID != "" || comment.ReplyCount < 1 {
			continue
		}
		seen[id] = true
		parents = append(parents, id)
		if len(parents) == taskContextMaxComments {
			break
		}
	}
	return parents
}
