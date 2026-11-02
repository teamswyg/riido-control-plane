package riidoaiserver

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"
)

func (s Server) handleAgentDashboard(w http.ResponseWriter, r *http.Request, agentID string) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	principal, ok := s.authorizeAIAgentClient(w, r, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionRead, AgentID: agentID})
	if !ok {
		return
	}
	page := 1
	if raw := r.URL.Query().Get("page"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 1 {
			writeError(w, http.StatusBadRequest, "invalid page")
			return
		}
		page = value
	}
	reader, ok := s.assignment.(AgentDashboardHistoryReader)
	if !ok || s.aiAgent == nil {
		writeError(w, http.StatusServiceUnavailable, "dashboard history is unavailable")
		return
	}
	bootstrap, err := s.aiAgent.BootstrapAIAgentClient(r.Context(), principal, ClientKindWeb)
	if err != nil {
		writeAIAgentClientError(w, err)
		return
	}
	var agent AgentClientRecord
	for _, candidate := range bootstrap.Agents {
		if candidate.AgentID == agentID {
			agent = candidate
			break
		}
	}
	if agent.AgentID == "" {
		writeError(w, http.StatusNotFound, "agent not found")
		return
	}
	load := reader.LoadAgentDashboardHistory
	section := r.URL.Query().Get("section")
	if section != "" && section != "active" {
		writeError(w, http.StatusBadRequest, "invalid dashboard section")
		return
	}
	if section == "active" {
		if active, ok := s.assignment.(interface {
			LoadAgentDashboardActive(context.Context, string) (AgentDashboardHistory, error)
		}); ok {
			load = active.LoadAgentDashboardActive
		} else {
			writeError(w, http.StatusServiceUnavailable, "dashboard history is unavailable")
			return
		}
	}
	history, err := load(r.Context(), agentID)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "dashboard history is unavailable")
		return
	}
	scopedRuns := make([]AgentDashboardRun, 0, len(history.Runs))
	for _, run := range history.Runs {
		if run.Assignment.WorkspaceID == "" || run.Assignment.WorkspaceID == principal.WorkspaceID {
			scopedRuns = append(scopedRuns, run)
		}
	}
	history.Runs = scopedRuns
	token, _ := requestToken(r)
	tasks := map[string]AIAgentTaskContextComponent{}
	checked := map[string]bool{}
	for _, run := range history.Runs {
		assignment := run.Assignment
		if assignment.WorkspaceID != "" && assignment.WorkspaceID != principal.WorkspaceID {
			continue
		}
		if checked[assignment.TaskID] {
			continue
		}
		checked[assignment.TaskID] = true
		task, err := s.dashboardTaskForViewer(r.Context(), token, principal.WorkspaceID, assignment.TaskID)
		if errors.Is(err, ErrAuthorizationForbidden) {
			continue
		}
		if err != nil {
			writeError(w, http.StatusServiceUnavailable, err.Error())
			return
		}
		tasks[assignment.TaskID] = task
	}
	view := buildAgentDashboard(agent, history, tasks, page, time.Now().UTC())
	if err := history.fillVisibleDetails(r.Context(), agentID, &view); err != nil {
		writeError(w, http.StatusServiceUnavailable, "dashboard history is unavailable")
		return
	}
	writeJSON(w, http.StatusOK, view)
}

var errDashboardTaskUnavailable = errors.New("dashboard task unavailable")
