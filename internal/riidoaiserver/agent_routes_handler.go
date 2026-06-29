package riidoaiserver

import (
	"net/http"
	"strings"
)

func (s Server) handleAgents(w http.ResponseWriter, r *http.Request) {
	agentID, suffix, ok := splitNestedResourcePath(r.URL.Path, "/v1/agents/")
	if !ok {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	switch {
	case suffix == "poll" && r.Method == http.MethodPost:
		s.handleAgentPoll(w, r, agentID)
	case suffix == "heartbeat" && r.Method == http.MethodPost:
		s.handleAgentHeartbeat(w, r, agentID)
	case suffix == "thread-progress" && r.Method == http.MethodPost:
		s.handleAgentThreadProgress(w, r, agentID)
	case suffix == "events" && r.Method == http.MethodPost:
		s.handleAgentEvent(w, r, agentID)
	case suffix == "tool-approvals" && r.Method == http.MethodPost:
		s.handleAgentToolApprovalCreate(w, r, agentID)
	case strings.HasPrefix(suffix, "tool-approvals/") && strings.HasSuffix(suffix, "/wait") && r.Method == http.MethodPost:
		approvalID, ok := toolApprovalWaitSuffixApprovalID(suffix)
		if !ok {
			writeError(w, http.StatusNotFound, "not found")
			return
		}
		s.handleAgentToolApprovalWait(w, r, agentID, approvalID)
	case suffix == "provider-status" && r.Method == http.MethodPost:
		s.handleProviderStatusSync(w, r, agentID)
	case suffix == "provider-status" && r.Method == http.MethodGet:
		s.handleProviderStatusRead(w, r, agentID)
	default:
		writeError(w, http.StatusNotFound, "not found")
	}
}
