package riidoaiserver

import (
	"context"
	"encoding/json"
	"sort"
	"strconv"
	"strings"
)

// readTaskContextComments reads up to taskContextMaxComments comments,
// newest first: the most recent ones get priority 2, older ones 3.
func (c *AIAgentPrivateTaskContextClient) readTaskContextComments(ctx context.Context, base, bearer string) ([]AIAgentTaskContextComment, int, string) {
	var raw []privateComment
	total, status := 0, "available"
	for page := 1; len(raw) < taskContextMaxComments; page++ {
		var body json.RawMessage
		if err := c.readPrivateContextComponent(ctx, base+"/comments?page="+strconv.Itoa(page)+"&limit=100", bearer, "comments", &body); err != nil {
			status = "partial"
			break
		}
		var decoded privateCommentPage
		if json.Unmarshal(body, &decoded) != nil {
			status = "partial"
			break
		}
		raw, total = append(raw, decoded.Comments...), max(decoded.Meta.Total, total)
		if decoded.Meta.NextPage == nil || len(decoded.Comments) == 0 {
			break
		}
	}
	comments := taskContextCommentsFromPrivate(raw)
	total = max(total, len(comments))
	if len(comments) < total {
		status = "partial"
	}
	if status == "partial" && len(comments) == 0 {
		status = "unavailable"
	}
	return comments, total, status
}

func taskContextCommentsFromPrivate(raw []privateComment) []AIAgentTaskContextComment {
	comments := make([]AIAgentTaskContextComment, 0, len(raw))
	for _, item := range raw {
		if strings.TrimSpace(item.ID) == "" || strings.EqualFold(item.Status, "deleted") || len(comments) == taskContextMaxComments {
			continue
		}
		body := boundedTaskContextText(item.Body, taskContextCommentBodyMax)
		comment := AIAgentTaskContextComment{
			ID: strings.TrimSpace(item.ID), Body: body, Truncated: body != strings.TrimSpace(strings.ToValidUTF8(item.Body, "�")),
			CreatedAt: normalizedTaskContextTime(item.CreatedAt), UpdatedAt: normalizedTaskContextTime(item.UpdatedAt), ReplyCount: max(item.ReplyCount, 0),
		}
		if item.Author != nil && item.Author.WorkspaceProfile != nil {
			comment.Author = boundedTaskContextText(item.Author.WorkspaceProfile.Nickname, 80)
		}
		comments = append(comments, comment)
	}
	sort.SliceStable(comments, func(i, j int) bool { return comments[i].CreatedAt > comments[j].CreatedAt })
	for i := range comments {
		comments[i].Priority = TaskContextCommentPriorityOlder
		if i < taskContextRecentComments {
			comments[i].Priority = TaskContextCommentPriorityRecent
		}
	}
	return comments
}
