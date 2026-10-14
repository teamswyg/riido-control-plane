package riidoaiserver

import (
	"net/http"
	"strings"
)

func (s Server) handleAIAgentClientAgents(w http.ResponseWriter, r *http.Request) {
	if s.aiAgent == nil {
		writeError(w, http.StatusServiceUnavailable, "ai agent client store is not configured")
		return
	}
	if r.URL.Path == "/v1/client/ai-agent/agents" || r.URL.Path == "/v1/client/ai-agent/agents/" {
		if r.Method == http.MethodPost {
			s.handleAIAgentClientCreate(w, r)
			return
		}
		writeMethodNotAllowed(w)
		return
	}
	agentID, suffix, ok := splitAIAgentClientAgentPath(r.URL.Path)
	if !ok {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	switch {
	case suffix == "daemon" && r.Method == http.MethodGet:
		s.handleAIAgentClientAgentDaemon(w, r, agentID)
	case strings.HasPrefix(suffix, "daemon/") && r.Method == http.MethodPost:
		action := strings.TrimPrefix(suffix, "daemon/")
		s.handleAIAgentClientAgentDaemonControl(w, r, agentID, action)
	case suffix == "editability" && r.Method == http.MethodGet:
		s.handleAIAgentClientEditability(w, r, agentID)
	case suffix == "" && r.Method == http.MethodPatch:
		s.handleAIAgentClientUpdate(w, r, agentID)
	case suffix == "" && r.Method == http.MethodDelete:
		s.handleAIAgentClientDelete(w, r, agentID)
	default:
		writeError(w, http.StatusNotFound, "not found")
	}
}
