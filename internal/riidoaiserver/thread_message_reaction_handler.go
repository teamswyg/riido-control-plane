package riidoaiserver

import (
	"net/http"
	"strings"
)

func (s Server) handleTaskMessageReaction(w http.ResponseWriter, r *http.Request, taskID, suffix string) {
	parts := strings.Split(suffix, "/")
	if len(parts) < 5 || len(parts) > 6 || parts[0] != "threads" || parts[2] != "messages" || parts[4] != "reactions" || parts[1] == "" || parts[3] == "" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	remove := r.Method == http.MethodDelete
	if (!remove && (r.Method != http.MethodPost || len(parts) != 5)) || (remove && len(parts) != 6) {
		writeMethodNotAllowed(w)
		return
	}
	principal, ok := s.authorizeAIAgentClient(w, r, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionCreate, TaskID: taskID})
	if !ok {
		return
	}
	var input struct {
		Emoji string `json:"emoji"`
	}
	if remove {
		input.Emoji = parts[5]
	} else if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	emoji, err := normalizeMessageReactionEmoji(input.Emoji)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	store, ok := s.aiAgent.(taskMessageReactionStore)
	if !ok {
		writeError(w, http.StatusServiceUnavailable, "message reactions are unavailable")
		return
	}
	response, err := store.MutateTaskMessageReaction(r.Context(), principal, taskID, parts[1], parts[3], emoji, remove)
	if err != nil {
		writeAIAgentClientError(w, err)
		return
	}
	s.notifyMemberMessageReaction(r.Context(), principal, response, emoji, remove)
	writeJSON(w, http.StatusOK, response)
}
