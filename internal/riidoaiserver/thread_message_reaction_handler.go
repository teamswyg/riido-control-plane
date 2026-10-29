package riidoaiserver

import (
	"context"
	"net/http"
	"strings"
)

type threadMessageReactor interface {
	ReactToAIAgentTaskThreadMessage(ctx context.Context, principal AuthorizationResult, taskID, threadID, messageID, emoji string, add bool) (threadMessageReactionChange, error)
}

// threadMessageReactionSuffix parses threads/{id}/messages/{id}/reactions and,
// with withEmoji, the trailing /{emoji}.
func threadMessageReactionSuffix(suffix string, withEmoji bool) (threadID, messageID, emoji string, ok bool) {
	parts := strings.Split(strings.Trim(suffix, "/"), "/")
	want := 5
	if withEmoji {
		want = 6
	}
	if len(parts) != want || parts[0] != "threads" || parts[2] != "messages" || parts[4] != "reactions" {
		return "", "", "", false
	}
	for _, part := range parts {
		if strings.TrimSpace(part) == "" {
			return "", "", "", false
		}
	}
	if withEmoji {
		emoji = parts[5]
	}
	return parts[1], parts[3], emoji, true
}

func (s Server) handleAIAgentClientThreadMessageReaction(w http.ResponseWriter, r *http.Request, taskID, suffix string) {
	add := r.Method == http.MethodPost
	threadID, messageID, emoji, ok := threadMessageReactionSuffix(suffix, !add)
	reactor, supported := s.aiAgent.(threadMessageReactor)
	if !ok || !supported {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	// Add and delete need the thread message scope; clients never act as an agent.
	principal, ok := s.authorizeAIAgentClient(w, r, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionCreate, TaskID: taskID})
	if !ok {
		return
	}
	if add {
		var req AddAIAgentTaskThreadMessageReactionRequest
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		emoji = req.Emoji
	}
	change, err := reactor.ReactToAIAgentTaskThreadMessage(r.Context(), principal, taskID, threadID, messageID, emoji, add)
	if err != nil {
		writeAIAgentClientError(w, err)
		return
	}
	if change.Notify != nil {
		s.queueThreadReactionNotification(*change.Notify)
	}
	writeJSON(w, http.StatusOK, change.Response)
}
