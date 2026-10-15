package riidoaiserver

import (
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
	history, err := reader.LoadAgentDashboardHistory(r.Context(), agentID)
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
		_, err := s.config.Authorizer.Authorize(r.Context(), token, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionRead, WorkspaceID: principal.WorkspaceID, TaskID: assignment.TaskID})
		if errors.Is(err, ErrAuthorizationForbidden) {
			continue
		}
		if err != nil {
			writeError(w, http.StatusServiceUnavailable, "task access could not be verified")
			return
		}
		// Require the viewer's credentials for metadata as well; workspace service keys
		// must never expand access for a shared agent's history.
		contextReader, ok := s.taskContext.(AIAgentTaskContextRequestReader)
		if !ok {
			writeError(w, http.StatusServiceUnavailable, "request-scoped task context is unavailable")
			return
		}
		request := AIAgentTaskContextRequest{ComponentID: assignment.TaskID, WorkspaceID: principal.WorkspaceID, BearerToken: token}
		readTask := contextReader.GetAIAgentTaskContextForRequest
		if metadata, ok := s.taskContext.(agentDashboardTaskReader); ok {
			readTask = metadata.GetAgentDashboardTask
		}
		context, err := readTask(r.Context(), request)
		if errors.Is(err, ErrAuthorizationForbidden) || errors.Is(err, errDashboardTaskUnavailable) {
			continue
		}
		if err != nil {
			writeError(w, http.StatusServiceUnavailable, "task context could not be verified")
			return
		}
		if context.Component.ID != assignment.TaskID {
			continue
		}
		tasks[assignment.TaskID] = context.Component
	}
	writeJSON(w, http.StatusOK, buildAgentDashboard(agent, history, tasks, page, time.Now().UTC()))
}

var errDashboardTaskUnavailable = errors.New("dashboard task unavailable")
