package riidoaiserver

import (
	"net/http"
)

func (s Server) handleAIAgentClientThreads(w http.ResponseWriter, r *http.Request) {
	if s.aiAgent == nil {
		writeError(w, http.StatusServiceUnavailable, "ai agent client store is not configured")
		return
	}
	threadID, suffix, ok := splitNestedResourcePath(r.URL.Path, "/v1/client/ai-agent/threads/")
	if !ok {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	switch {
	case suffix == "messages" && r.Method == http.MethodPost:
		s.handleAIAgentClientCreateThreadMessage(w, r, threadID)
	case suffix == "messages":
		writeMethodNotAllowed(w)
	default:
		writeError(w, http.StatusNotFound, "not found")
	}
}

func (s Server) handleAIAgentClientCreateThreadMessage(w http.ResponseWriter, r *http.Request, threadID string) {
	thread, err := s.aiAgent.FindAIAgentTaskThreadByID(r.Context(), aiAgentWorkspaceIDFromRequest(r), threadID)
	if err != nil {
		writeAIAgentClientError(w, err)
		return
	}
	s.handleAIAgentClientCreateTaskThreadMessage(w, r, thread.TaskID, thread.ThreadID)
}
