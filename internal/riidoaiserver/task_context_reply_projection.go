package riidoaiserver

import (
	"strconv"
	"strings"
)

func currentReplyMetadata(parents []string, roots []AIAgentTaskContextComment, response map[string][]privateReplyMetadata) []AIAgentTaskContextComment {
	seen := map[string]bool{}
	for _, root := range roots {
		seen[root.ID] = true
	}
	var replies []AIAgentTaskContextComment
	for i, parent := range parents {
		for _, reply := range response["r"+strconv.Itoa(i)] {
			id := strings.TrimSpace(reply.ID)
			if id == "" || seen[id] || reply.Status != "active" {
				continue
			}
			seen[id] = true
			// No body was requested: an empty-body hash must never prove unchanged.
			replies = append(replies, AIAgentTaskContextComment{ID: id, ParentID: parent, UpdatedAt: normalizedTaskContextTime(reply.UpdatedAt), Truncated: true, MetadataOnly: true})
			if len(replies) == taskContextMaxComments {
				return replies
			}
		}
	}
	return replies
}
