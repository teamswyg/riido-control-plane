package riidoaiserver

import (
	"net/http"
	"strconv"
	"time"
)

// User reads never create a synthetic Assignment or claim a mention precheck.
// Only the dedicated task-bound credential can enter this public projection.
func (s Server) handleCLIAgentDirectory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	target, ok := cliAgentReadTargetFromRequest(r)
	if !ok || !hasCLIAgentReadCredential(r) {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}
	principal, ok := s.authorizeRequest(w, r, AuthorizationRequest{Resource: AuthorizationResourceCLIAgentRead, Action: AuthorizationActionRead, WorkspaceID: target.workspace, TaskID: target.task, AgentID: target.agent})
	if !ok {
		return
	}
	if s.aiAgent == nil {
		writeError(w, http.StatusServiceUnavailable, "agent directory is unavailable")
		return
	}
	if err := s.reconcileAIAgentTaskThreadProjections(r.Context(), principal, ""); err != nil {
		writeError(w, http.StatusServiceUnavailable, "agent directory is unavailable")
		return
	}
	list, err := s.aiAgent.ListAIAgentTaskAssignableAgents(r.Context(), principal, target.task)
	if err != nil || r.Context().Err() != nil {
		writeError(w, http.StatusServiceUnavailable, "agent directory is unavailable")
		return
	}
	entries := make([]AgentDirectoryEntry, 0, len(list.Agents))
	for _, agent := range list.Agents {
		entry := s.agentDirectoryEntry(agent, "")
		entry.Runtime.LastSeenAt = list.agentDaemonLastSeenAt[agent.AgentID]
		entries = append(entries, entry)
	}
	s.writeCLIAgentDirectory(w, r, target, entries)
}

func (s Server) writeCLIAgentDirectory(w http.ResponseWriter, r *http.Request, target cliAgentReadTarget, entries []AgentDirectoryEntry) {
	if r.Context().Err() != nil {
		return
	}
	if target.agent != "" {
		for _, entry := range entries {
			if entry.AgentID == target.agent {
				writeJSON(w, http.StatusOK, AgentDirectoryDetail{Agent: entry, FetchedAt: time.Now().UTC()})
				return
			}
		}
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}
	out := AgentDirectoryPage{Agents: []AgentDirectoryEntry{}, Total: len(entries), Cursor: strconv.Itoa(target.page), FetchedAt: time.Now().UTC()}
	if target.page < (len(entries)+agentDirectoryPageSize-1)/agentDirectoryPageSize {
		start := target.page * agentDirectoryPageSize
		end := min(start+agentDirectoryPageSize, len(entries))
		out.Agents = entries[start:end]
		if end < len(entries) {
			out.NextCursor = strconv.Itoa(target.page + 1)
		}
	}
	writeJSON(w, http.StatusOK, out)
}
