package riidoaiserver

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/teamswyg/riido-contracts/metadatakeys"
)

const (
	aiAgentTokenHeader = "X-Riido-Ai-Agent-Token"
	deviceIDHeader     = "X-Riido-Device-Id"
	deviceSecretHeader = "X-Riido-Device-Secret"
)

// sseKeepaliveInterval is how often an idle SSE stream emits a comment line to
// keep the connection alive through ALB/proxy idle timeouts.
const sseKeepaliveInterval = 15 * time.Second

func (s Server) handleAIAgentClientBootstrap(w http.ResponseWriter, r *http.Request) {
	if s.aiAgent == nil {
		writeError(w, http.StatusServiceUnavailable, "ai agent client store is not configured")
		return
	}
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	principal, ok := s.authorizeAIAgentClient(w, r, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionRead})
	if !ok {
		return
	}
	if err := s.reconcileAIAgentTaskThreadProjections(r.Context(), principal, ""); err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	response, err := s.aiAgent.BootstrapAIAgentClient(r.Context(), principal, ClientKind(strings.TrimSpace(r.URL.Query().Get("client_kind"))))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (s Server) handleAIAgentClientDevices(w http.ResponseWriter, r *http.Request) {
	if s.aiAgent == nil {
		writeError(w, http.StatusServiceUnavailable, "ai agent client store is not configured")
		return
	}
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	principal, ok := s.authorizeAIAgentClient(w, r, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionDeviceRead})
	if !ok {
		return
	}
	response, err := s.aiAgent.ListAIAgentDevices(r.Context(), principal)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (s Server) handleAIAgentClientOnboardingFixtures(w http.ResponseWriter, r *http.Request) {
	if s.aiAgent == nil {
		writeError(w, http.StatusServiceUnavailable, "ai agent client store is not configured")
		return
	}
	if r.URL.Path == "/v1/client/ai-agent/onboarding/fixtures" || r.URL.Path == "/v1/client/ai-agent/onboarding/fixtures/" {
		if r.Method != http.MethodGet {
			writeMethodNotAllowed(w)
			return
		}
		principal, ok := s.authorizeAIAgentClient(w, r, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionRead})
		if !ok {
			return
		}
		response, err := s.aiAgent.ListAIAgentOnboardingFixtures(r.Context(), principal)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, response)
		return
	}
	fixtureID, suffix, ok := splitNestedResourcePath(r.URL.Path, "/v1/client/ai-agent/onboarding/fixtures/")
	if !ok {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	switch {
	case suffix == "agents" && r.Method == http.MethodPost:
		s.handleAIAgentClientCreateFromOnboardingFixture(w, r, fixtureID)
	default:
		writeError(w, http.StatusNotFound, "not found")
	}
}

func (s Server) handleAIAgentClientCreateFromOnboardingFixture(w http.ResponseWriter, r *http.Request, fixtureID string) {
	principal, ok := s.authorizeAIAgentClient(w, r, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionCreate})
	if !ok {
		return
	}
	var req CreateAgentConfigurationRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	response, err := s.aiAgent.CreateAIAgentFromOnboardingFixture(r.Context(), principal, fixtureID, req)
	if err != nil {
		writeAIAgentClientError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, response)
}

func (s Server) handleAIAgentClientProfileThumbnailUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w)
		return
	}
	if s.aiAgentProfileThumbnails == nil {
		writeError(w, http.StatusServiceUnavailable, "profile thumbnail upload service is not configured")
		return
	}
	principal, ok := s.authorizeAIAgentClient(w, r, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionCreate})
	if !ok {
		return
	}
	var req CreateAgentProfileThumbnailUploadRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	response, err := s.aiAgentProfileThumbnails.CreateAIAgentProfileThumbnailUpload(r.Context(), principal, req)
	if err != nil {
		writeAIAgentClientError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, response)
}

func (s Server) handleAIAgentClientDeviceRoutes(w http.ResponseWriter, r *http.Request) {
	if s.aiAgent == nil {
		writeError(w, http.StatusServiceUnavailable, "ai agent client store is not configured")
		return
	}
	deviceID, suffix, ok := splitAIAgentClientDevicePath(r.URL.Path)
	if !ok {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	switch {
	case suffix == "":
		writeMethodNotAllowed(w)
	case suffix == "daemon" && r.Method == http.MethodGet:
		s.handleAIAgentClientDeviceDaemon(w, r, deviceID)
	case strings.HasPrefix(suffix, "daemon/") && r.Method == http.MethodPost:
		s.handleAIAgentClientDeviceDaemonControl(w, r, deviceID, strings.TrimPrefix(suffix, "daemon/"))
	default:
		writeError(w, http.StatusNotFound, "not found")
	}
}

func (s Server) handleAIAgentClientDeviceDaemon(w http.ResponseWriter, r *http.Request, deviceID string) {
	principal, ok := s.authorizeAIAgentClient(w, r, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionDeviceRead, DeviceID: deviceID})
	if !ok {
		return
	}
	response, err := s.aiAgent.GetAIAgentDeviceDaemon(r.Context(), principal, deviceID)
	if err != nil {
		writeAIAgentClientError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (s Server) handleAIAgentClientDeviceDaemonControl(w http.ResponseWriter, r *http.Request, deviceID, actionValue string) {
	var action DaemonControlAction
	switch actionValue {
	case string(DaemonControlActionStart):
		action = DaemonControlActionStart
	case string(DaemonControlActionRestart):
		action = DaemonControlActionRestart
	case string(DaemonControlActionStop):
		action = DaemonControlActionStop
	default:
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	principal, ok := s.authorizeAIAgentClient(w, r, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionDeviceControl, DeviceID: deviceID})
	if !ok {
		return
	}
	var req ControlDeviceDaemonRequest
	if r.Body != nil && r.ContentLength != 0 {
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	response, err := s.aiAgent.ControlAIAgentDeviceDaemon(r.Context(), principal, deviceID, action, req)
	if err != nil {
		writeAIAgentClientError(w, err)
		return
	}
	writeJSON(w, http.StatusAccepted, response)
}

func (s Server) handleAIAgentClientAgentDaemon(w http.ResponseWriter, r *http.Request, agentID string) {
	principal, ok := s.authorizeAIAgentClient(w, r, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionDeviceRead, AgentID: agentID})
	if !ok {
		return
	}
	response, err := s.aiAgent.GetAIAgentDaemon(r.Context(), principal, agentID)
	if err != nil {
		writeAIAgentClientError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (s Server) handleAIAgentClientAgentDaemonControl(w http.ResponseWriter, r *http.Request, agentID, actionValue string) {
	var action DaemonControlAction
	switch actionValue {
	case string(DaemonControlActionStart):
		action = DaemonControlActionStart
	case string(DaemonControlActionRestart):
		action = DaemonControlActionRestart
	case string(DaemonControlActionStop):
		action = DaemonControlActionStop
	default:
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	principal, ok := s.authorizeAIAgentClient(w, r, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionDeviceControl, AgentID: agentID})
	if !ok {
		return
	}
	var req ControlDeviceDaemonRequest
	if r.Body != nil && r.ContentLength != 0 {
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	response, err := s.aiAgent.ControlAIAgentDaemon(r.Context(), principal, agentID, action, req)
	if err != nil {
		writeAIAgentClientError(w, err)
		return
	}
	writeJSON(w, http.StatusAccepted, response)
}

func (s Server) handleAIAgentClientTasks(w http.ResponseWriter, r *http.Request) {
	if s.aiAgent == nil {
		writeError(w, http.StatusServiceUnavailable, "ai agent client store is not configured")
		return
	}
	taskID, suffix, ok := splitNestedResourcePath(r.URL.Path, "/v1/client/ai-agent/tasks/")
	if !ok {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	switch {
	case suffix == "assignable-agents" && r.Method == http.MethodGet:
		s.handleAIAgentClientTaskAssignableAgents(w, r, taskID)
	case suffix == "assignment" && r.Method == http.MethodPost:
		s.handleAIAgentClientAssignTask(w, r, taskID)
	case suffix == "assignment" && r.Method == http.MethodDelete:
		s.handleAIAgentClientUnassignTask(w, r, taskID)
	case suffix == "agent-assignments" && r.Method == http.MethodPost:
		s.handleAIAgentClientCreateTaskAgentAssignment(w, r, taskID)
	case strings.HasPrefix(suffix, "agent-assignments/") && strings.HasSuffix(suffix, "/stop") && r.Method == http.MethodPost:
		agentID, ok := agentAssignmentStopSuffixAgentID(suffix)
		if !ok {
			writeError(w, http.StatusNotFound, "not found")
			return
		}
		s.handleAIAgentClientStopTaskAgentAssignment(w, r, taskID, agentID)
	case strings.HasPrefix(suffix, "agent-assignments/") && r.Method == http.MethodDelete:
		agentID, ok := agentAssignmentSuffixAgentID(suffix)
		if !ok {
			writeError(w, http.StatusNotFound, "not found")
			return
		}
		s.handleAIAgentClientDeleteTaskAgentAssignment(w, r, taskID, agentID)
	case suffix == "threads" && r.Method == http.MethodGet:
		s.handleAIAgentClientTaskThreads(w, r, taskID)
	case suffix == "tool-approvals" && r.Method == http.MethodGet:
		s.handleAIAgentClientTaskToolApprovals(w, r, taskID)
	case strings.HasPrefix(suffix, "tool-approvals/") && strings.HasSuffix(suffix, "/decision") && r.Method == http.MethodPost:
		approvalID, ok := toolApprovalDecisionSuffixApprovalID(suffix)
		if !ok {
			writeError(w, http.StatusNotFound, "not found")
			return
		}
		s.handleAIAgentClientTaskToolApprovalDecision(w, r, taskID, approvalID)
	case suffix == "thread-stream-subscription" && r.Method == http.MethodGet:
		s.handleAIAgentClientTaskThreadStreamSubscription(w, r, taskID)
	case suffix == "comments" && r.Method == http.MethodPost:
		s.handleAIAgentClientSubmitTaskComment(w, r, taskID)
	case strings.HasPrefix(suffix, "threads/") && strings.HasSuffix(suffix, "/messages") && r.Method == http.MethodPost:
		threadID, ok := threadMessageSuffixThreadID(suffix)
		if !ok {
			writeError(w, http.StatusNotFound, "not found")
			return
		}
		s.handleAIAgentClientCreateTaskThreadMessage(w, r, taskID, threadID)
	case suffix == "stop" && r.Method == http.MethodPost:
		s.handleAIAgentClientStopTask(w, r, taskID)
	default:
		writeMethodNotAllowed(w)
	}
}

func (s Server) handleAIAgentClientTaskAssignableAgents(w http.ResponseWriter, r *http.Request, taskID string) {
	principal, ok := s.authorizeAIAgentClient(w, r, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionRead, TaskID: taskID})
	if !ok {
		return
	}
	response, err := s.aiAgent.ListAIAgentTaskAssignableAgents(r.Context(), principal, taskID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (s Server) handleAIAgentClientWorkspaceAssignedAgentProfiles(w http.ResponseWriter, r *http.Request) {
	if s.aiAgent == nil {
		writeError(w, http.StatusServiceUnavailable, "ai agent client store is not configured")
		return
	}
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	principal, ok := s.authorizeAIAgentClient(w, r, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionRead})
	if !ok {
		return
	}
	if err := s.reconcileAIAgentTaskThreadProjections(r.Context(), principal, ""); err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	response, err := s.aiAgent.ListWorkspaceAssignedAgentProfiles(r.Context(), principal)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (s Server) handleAIAgentClientAssignTask(w http.ResponseWriter, r *http.Request, taskID string) {
	if s.assignment == nil {
		writeError(w, http.StatusServiceUnavailable, "assignment store is not configured")
		return
	}
	principal, ok := s.authorizeAIAgentClient(w, r, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionAssign, TaskID: taskID})
	if !ok {
		return
	}
	bearerToken, _ := requestToken(r)
	var req AssignAIAgentTaskRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.reconcileAIAgentTaskThreadProjections(r.Context(), principal, taskID); err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	assignmentReq, err := s.assignRequestFromAIAgentClientTask(r.Context(), principal, bearerToken, taskID, req)
	if err != nil {
		writeAIAgentClientError(w, err)
		return
	}
	assignment, err := s.assignment.AssignTaskReplacement(r.Context(), taskID, assignmentReq)
	if err != nil {
		writeAIAgentClientError(w, err)
		return
	}
	req.AssignmentID, req.durableState = assignment.ID, assignmentClientResponseDurableState(assignment)
	response, err := s.aiAgent.AssignAIAgentTask(r.Context(), principal, taskID, req)
	if err != nil {
		writeAIAgentClientError(w, err)
		return
	}
	writeJSON(w, http.StatusAccepted, response)
}

func (s Server) handleAIAgentClientCreateTaskAgentAssignment(w http.ResponseWriter, r *http.Request, taskID string) {
	if aiAgentWorkspaceIDFromRequest(r) == "" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	if s.assignment == nil {
		writeError(w, http.StatusServiceUnavailable, "assignment store is not configured")
		return
	}
	principal, ok := s.authorizeAIAgentClient(w, r, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionAssign, TaskID: taskID})
	if !ok {
		return
	}
	bearerToken, _ := requestToken(r)
	var req AssignAIAgentTaskRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.reconcileAIAgentTaskThreadProjections(r.Context(), principal, taskID); err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	assignmentReq, err := s.assignRequestFromAIAgentClientTask(r.Context(), principal, bearerToken, taskID, req)
	if err != nil {
		writeAIAgentClientError(w, err)
		return
	}
	assignment, err := s.assignment.AssignTaskAdditive(r.Context(), taskID, assignmentReq)
	if err != nil {
		writeAIAgentClientError(w, err)
		return
	}
	req.AssignmentID, req.durableState = assignment.ID, assignmentClientResponseDurableState(assignment)
	response, err := s.aiAgent.CreateAIAgentTaskAgentAssignment(r.Context(), principal, taskID, req)
	if err != nil {
		writeAIAgentClientError(w, err)
		return
	}
	writeJSON(w, http.StatusAccepted, response)
}

func (s Server) handleAIAgentClientUnassignTask(w http.ResponseWriter, r *http.Request, taskID string) {
	principal, ok := s.authorizeAIAgentClient(w, r, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionStop, TaskID: taskID})
	if !ok {
		return
	}
	var req UnassignAIAgentTaskRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if target, ok, err := s.cancelAIAgentAssignmentBeforeAction(r.Context(), principal, taskID, req.AgentID, req.AssignmentID, req.Reason); err != nil {
		writeAIAgentClientError(w, err)
		return
	} else if ok {
		req.AgentID = target.AgentID
		req.AssignmentID = target.AssignmentID
	}
	response, err := s.aiAgent.UnassignAIAgentTask(r.Context(), principal, taskID, req)
	if err != nil {
		writeAIAgentClientError(w, err)
		return
	}
	writeJSON(w, http.StatusAccepted, response)
}

func (s Server) handleAIAgentClientDeleteTaskAgentAssignment(w http.ResponseWriter, r *http.Request, taskID, agentID string) {
	if aiAgentWorkspaceIDFromRequest(r) == "" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	principal, ok := s.authorizeAIAgentClient(w, r, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionStop, TaskID: taskID})
	if !ok {
		return
	}
	var req AgentAssignmentActionRequest
	if r.Body != nil && r.ContentLength != 0 {
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	if target, ok, err := s.cancelAIAgentAssignmentBeforeAction(r.Context(), principal, taskID, agentID, req.AssignmentID, req.Reason); err != nil {
		writeAIAgentClientError(w, err)
		return
	} else if ok {
		req.AssignmentID = target.AssignmentID
	}
	response, err := s.aiAgent.DeleteAIAgentTaskAgentAssignment(r.Context(), principal, taskID, agentID, req)
	if err != nil {
		writeAIAgentClientError(w, err)
		return
	}
	writeJSON(w, http.StatusAccepted, response)
}

func (s Server) handleAIAgentClientTaskThreads(w http.ResponseWriter, r *http.Request, taskID string) {
	principal, ok := s.authorizeAIAgentClient(w, r, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionRead, TaskID: taskID})
	if !ok {
		return
	}
	if err := s.reconcileAIAgentTaskThreadProjections(r.Context(), principal, taskID); err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	response, err := s.aiAgent.ListAIAgentTaskThreads(r.Context(), principal, taskID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (s Server) handleAIAgentClientTaskThreadStreamSubscription(w http.ResponseWriter, r *http.Request, taskID string) {
	if aiAgentWorkspaceIDFromRequest(r) == "" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	principal, ok := s.authorizeAIAgentClient(w, r, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionRead, TaskID: taskID})
	if !ok {
		return
	}
	if err := s.reconcileAIAgentTaskThreadProjections(r.Context(), principal, taskID); err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	response, err := s.aiAgent.GetAIAgentTaskThreadStreamSubscription(r.Context(), principal, taskID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (s Server) handleAIAgentClientSubmitTaskComment(w http.ResponseWriter, r *http.Request, taskID string) {
	principal, ok := s.authorizeAIAgentClient(w, r, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionCreate, TaskID: taskID})
	if !ok {
		return
	}
	var req SubmitAIAgentTaskCommentRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	response, err := s.aiAgent.SubmitAIAgentTaskComment(r.Context(), principal, taskID, req)
	if err != nil {
		writeAIAgentClientError(w, err)
		return
	}
	writeJSON(w, http.StatusAccepted, response)
}

func (s Server) handleAIAgentClientCreateTaskThreadMessage(w http.ResponseWriter, r *http.Request, taskID, threadID string) {
	if s.assignment == nil {
		writeError(w, http.StatusServiceUnavailable, "assignment store is not configured")
		return
	}
	principal, ok := s.authorizeAIAgentClient(w, r, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionCreate, TaskID: taskID})
	if !ok {
		return
	}
	bearerToken, _ := requestToken(r)
	var req CreateAIAgentTaskThreadMessageRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	assignmentReq, err := s.assignRequestFromAIAgentTaskThreadMessage(r.Context(), principal, bearerToken, taskID, threadID, req)
	if err != nil {
		writeAIAgentClientError(w, err)
		return
	}
	assignment, err := s.assignment.AssignTask(r.Context(), taskID, assignmentReq)
	if err != nil {
		writeAIAgentClientError(w, err)
		return
	}
	req.AssignmentID = assignment.ID
	req.durableState = assignment.State
	response, err := s.aiAgent.CreateAIAgentTaskThreadMessage(r.Context(), principal, taskID, threadID, req)
	if err != nil {
		writeAIAgentClientError(w, err)
		return
	}
	writeJSON(w, http.StatusAccepted, response)
}

func (s Server) handleAIAgentClientStopTask(w http.ResponseWriter, r *http.Request, taskID string) {
	principal, ok := s.authorizeAIAgentClient(w, r, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionStop, TaskID: taskID})
	if !ok {
		return
	}
	var req StopAIAgentTaskRequest
	if r.Body != nil && r.ContentLength != 0 {
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	if target, ok, err := s.cancelAIAgentAssignmentBeforeAction(r.Context(), principal, taskID, req.AgentID, req.AssignmentID, req.Reason); err != nil {
		writeAIAgentClientError(w, err)
		return
	} else if ok {
		req.AgentID = target.AgentID
		req.AssignmentID = target.AssignmentID
		req.durableState = target.State
	}
	response, err := s.aiAgent.StopAIAgentTask(r.Context(), principal, taskID, req)
	if err != nil {
		writeAIAgentClientError(w, err)
		return
	}
	writeJSON(w, http.StatusAccepted, response)
}

func (s Server) handleAIAgentClientStopTaskAgentAssignment(w http.ResponseWriter, r *http.Request, taskID, agentID string) {
	if aiAgentWorkspaceIDFromRequest(r) == "" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	principal, ok := s.authorizeAIAgentClient(w, r, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionStop, TaskID: taskID})
	if !ok {
		return
	}
	var req AgentAssignmentActionRequest
	if r.Body != nil && r.ContentLength != 0 {
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	if target, ok, err := s.cancelAIAgentAssignmentBeforeAction(r.Context(), principal, taskID, agentID, req.AssignmentID, req.Reason); err != nil {
		writeAIAgentClientError(w, err)
		return
	} else if ok {
		req.AssignmentID = target.AssignmentID
		req.durableState = target.State
	}
	response, err := s.aiAgent.StopAIAgentTaskAgentAssignment(r.Context(), principal, taskID, agentID, req)
	if err != nil {
		writeAIAgentClientError(w, err)
		return
	}
	writeJSON(w, http.StatusAccepted, response)
}

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

func (s Server) handleAIAgentClientCreate(w http.ResponseWriter, r *http.Request) {
	principal, ok := s.authorizeAIAgentClient(w, r, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionCreate})
	if !ok {
		return
	}
	var req CreateAgentConfigurationRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	response, err := s.aiAgent.CreateAIAgent(r.Context(), principal, req)
	if err != nil {
		writeAIAgentClientError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, response)
}

func (s Server) handleAIAgentClientEditability(w http.ResponseWriter, r *http.Request, agentID string) {
	principal, ok := s.authorizeAIAgentClient(w, r, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionRead, AgentID: agentID})
	if !ok {
		return
	}
	response, err := s.aiAgent.GetAIAgentEditability(r.Context(), principal, agentID)
	if err != nil {
		writeAIAgentClientError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (s Server) handleAIAgentClientUpdate(w http.ResponseWriter, r *http.Request, agentID string) {
	principal, ok := s.authorizeAIAgentClient(w, r, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionUpdate, AgentID: agentID})
	if !ok {
		return
	}
	var req UpdateAgentConfigurationRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	response, err := s.aiAgent.UpdateAIAgentConfiguration(r.Context(), principal, agentID, req)
	if err != nil {
		writeAIAgentClientError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (s Server) handleAIAgentClientDelete(w http.ResponseWriter, r *http.Request, agentID string) {
	principal, ok := s.authorizeAIAgentClient(w, r, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionDelete, AgentID: agentID})
	if !ok {
		return
	}
	response, err := s.aiAgent.DeleteAIAgent(r.Context(), principal, agentID)
	if err != nil {
		writeAIAgentClientError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (s Server) handleAIAgentClientEvents(w http.ResponseWriter, r *http.Request) {
	if s.aiAgent == nil {
		writeError(w, http.StatusServiceUnavailable, "ai agent client store is not configured")
		return
	}
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	principal, ok := s.authorizeAIAgentClient(w, r, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionStream})
	if !ok {
		return
	}
	var (
		events []ClientStreamEvent
		live   <-chan ClientStreamEvent
		cancel func()
		err    error
	)
	if subscriber, ok := s.aiAgent.(AIAgentClientEventSubscriber); ok {
		events, live, cancel, err = subscriber.SubscribeAIAgentClientEvents(r.Context(), principal)
		if cancel != nil {
			defer cancel()
		}
	} else {
		events, err = s.aiAgent.AIAgentClientEvents(r.Context(), principal)
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	for _, event := range events {
		if err := writeAIAgentClientSSE(w, event); err != nil {
			return
		}
	}
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
	if r.URL.Query().Get("replay") == "1" {
		return
	}
	keepalive := time.NewTicker(sseKeepaliveInterval)
	defer keepalive.Stop()
	for {
		select {
		case event, ok := <-live:
			if !ok {
				return
			}
			if err := writeAIAgentClientSSE(w, event); err != nil {
				return
			}
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
		case <-keepalive.C:
			// Periodic SSE comment keeps the connection alive through idle gaps so
			// ALB/proxy idle timeouts don't drop the stream (choppy reconnects).
			if _, err := io.WriteString(w, ": keepalive\n\n"); err != nil {
				return
			}
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
		case <-r.Context().Done():
			return
		}
	}
}

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

func (s Server) handleComponentTasks(w http.ResponseWriter, r *http.Request) {
	taskID, suffix, ok := splitResourcePath(r.URL.Path, "/v1/component-tasks/")
	if !ok {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	switch {
	case suffix == "assignment" && r.Method == http.MethodPost:
		s.handleComponentTaskAssign(w, r, taskID)
	case suffix == "events" && r.Method == http.MethodGet:
		s.handleComponentTaskEvents(w, r, taskID)
	default:
		writeError(w, http.StatusNotFound, "not found")
	}
}

func (s Server) handleComponentTaskAssign(w http.ResponseWriter, r *http.Request, taskID string) {
	if s.assignment == nil {
		writeError(w, http.StatusServiceUnavailable, "assignment store is not configured")
		return
	}
	if _, ok := s.authorizeRequest(w, r, AuthorizationRequest{Resource: AuthorizationResourceComponentTask, Action: AuthorizationActionAssign, TaskID: taskID}); !ok {
		return
	}
	var req AssignRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if strings.TrimSpace(req.Prompt) == "" && s.taskContext != nil {
		composedReq, err := s.assignRequestWithTaskContextPrompt(r.Context(), taskID, req)
		if err != nil {
			writeError(w, http.StatusBadGateway, err.Error())
			return
		}
		req = composedReq
	}
	assignment, err := s.assignment.AssignTask(r.Context(), taskID, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, struct {
		SchemaVersion string     `json:"schema_version"`
		Assignment    Assignment `json:"assignment"`
	}{SchemaVersion: SchemaVersion, Assignment: assignment})
}

func (s Server) handleComponentTaskEvents(w http.ResponseWriter, r *http.Request, taskID string) {
	if s.assignment == nil {
		writeError(w, http.StatusServiceUnavailable, "assignment store is not configured")
		return
	}
	if _, ok := s.authorizeRequest(w, r, AuthorizationRequest{Resource: AuthorizationResourceComponentTaskEvents, Action: AuthorizationActionEventsRead, TaskID: taskID}); !ok {
		return
	}
	s.streamTaskEvents(w, r, taskID)
}

func (s Server) handleAgentPoll(w http.ResponseWriter, r *http.Request, agentID string) {
	if s.assignment == nil {
		writeError(w, http.StatusServiceUnavailable, "assignment store is not configured")
		return
	}
	if _, ok := s.authorizeRequest(w, r, AuthorizationRequest{Resource: AuthorizationResourceAgent, Action: AuthorizationActionPoll, AgentID: agentID}); !ok {
		return
	}
	var req PollRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	var response PollResponse
	var err error
	if longPoll, ok := s.assignment.(AssignmentLongPollStore); ok && req.WaitMs > 0 {
		hold := time.Duration(req.WaitMs) * time.Millisecond
		if maxHold := s.config.LongPollMaxHold; maxHold > 0 && hold > maxHold {
			hold = maxHold
		}
		response, err = longPoll.WaitForAssignment(r.Context(), agentID, req, hold, s.config.LongPollTick)
	} else {
		response, err = s.assignment.PollAgent(r.Context(), agentID, req)
	}
	if err != nil {
		logAgentPollRejected(agentID, err)
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (s Server) handleAgentHeartbeat(w http.ResponseWriter, r *http.Request, agentID string) {
	if s.assignment == nil {
		writeError(w, http.StatusServiceUnavailable, "assignment store is not configured")
		return
	}
	if _, ok := s.authorizeRequest(w, r, AuthorizationRequest{Resource: AuthorizationResourceAgent, Action: AuthorizationActionHeartbeat, AgentID: agentID}); !ok {
		return
	}
	var req AgentHeartbeatRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	var response AgentHeartbeatResponse
	var events []TaskEvent
	var err error
	if heartbeatStore, ok := s.assignment.(AssignmentHeartbeatEventStore); ok {
		response, events, err = heartbeatStore.HeartbeatAgentWithEvents(r.Context(), agentID, req)
	} else {
		response, err = s.assignment.HeartbeatAgent(r.Context(), agentID, req)
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if recorder, ok := s.aiAgent.(AIAgentAssignmentEventRecorder); ok {
		for _, event := range events {
			_ = recorder.RecordAIAgentAssignmentEvent(r.Context(), agentID, AgentEventRequest{
				AssignmentID: event.AssignmentID,
				TaskID:       event.TaskID,
				DaemonID:     req.DaemonID,
				DeviceID:     req.DeviceID,
				RuntimeID:    req.RuntimeID,
				State:        event.State,
				EventType:    event.Type,
				Message:      event.Message,
				Metadata:     event.Metadata,
			}, event)
		}
	}
	writeJSON(w, http.StatusOK, response)
}

func (s Server) handleAgentEvent(w http.ResponseWriter, r *http.Request, agentID string) {
	if s.assignment == nil {
		writeError(w, http.StatusServiceUnavailable, "assignment store is not configured")
		return
	}
	if _, ok := s.authorizeRequest(w, r, AuthorizationRequest{Resource: AuthorizationResourceAgent, Action: AuthorizationActionEventsWrite, AgentID: agentID}); !ok {
		return
	}
	var req AgentEventRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	response, err := s.assignment.RecordAgentEvent(r.Context(), agentID, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if recorder, ok := s.aiAgent.(AIAgentAssignmentEventRecorder); ok {
		_ = recorder.RecordAIAgentAssignmentEvent(r.Context(), agentID, req, response.Event)
	}
	writeJSON(w, http.StatusOK, response)
}

func (s Server) handleAgentThreadProgress(w http.ResponseWriter, r *http.Request, agentID string) {
	if s.assignment == nil {
		writeError(w, http.StatusServiceUnavailable, "assignment store is not configured")
		return
	}
	if _, ok := s.authorizeRequest(w, r, AuthorizationRequest{Resource: AuthorizationResourceAgent, Action: AuthorizationActionEventsWrite, AgentID: agentID}); !ok {
		return
	}
	var req AgentThreadProgressBatchRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	req.AssignmentID = strings.TrimSpace(req.AssignmentID)
	req.TaskID = strings.TrimSpace(req.TaskID)
	req.ThreadID = strings.TrimSpace(req.ThreadID)
	req.RunID = strings.TrimSpace(req.RunID)
	if req.AssignmentID == "" {
		writeError(w, http.StatusBadRequest, "assignment_id is required")
		return
	}
	if req.TaskID == "" {
		writeError(w, http.StatusBadRequest, "task_id is required")
		return
	}
	if req.RunID == "" {
		req.RunID = "run-" + req.AssignmentID
	}
	lines := normalizeProgressLines(req.Lines)
	if len(lines) == 0 {
		writeError(w, http.StatusBadRequest, "lines are required")
		return
	}
	req.Lines = lines
	for _, line := range lines {
		metadata := copyStringMap(req.Metadata)
		metadata[metadatakeys.ThreadProgressSeq.String()] = fmt.Sprint(line.Seq)
		metadata = addProgressLineMetadata(metadata, line)
		if _, err := s.assignment.RecordAgentEvent(r.Context(), agentID, AgentEventRequest{
			AssignmentID: req.AssignmentID,
			TaskID:       req.TaskID,
			DaemonID:     req.DaemonID,
			DeviceID:     req.DeviceID,
			RuntimeID:    req.RuntimeID,
			State:        AssignmentRunning,
			EventType:    EventRiidoLog,
			Message:      line.Message,
			Metadata:     metadata,
		}); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	if recorder, ok := s.aiAgent.(AIAgentThreadProgressRecorder); ok {
		response, err := recorder.RecordAIAgentThreadProgress(r.Context(), agentID, req)
		if err != nil {
			writeAIAgentClientError(w, err)
			return
		}
		writeJSON(w, http.StatusAccepted, response)
		return
	}
	if req.ThreadID == "" {
		req.ThreadID = threadIDForRun(req.TaskID, agentID, req.RunID)
	}
	writeJSON(w, http.StatusAccepted, AgentThreadProgressBatchResponse{
		SchemaVersion: SchemaVersion,
		AcceptedLines: len(lines),
		Event: AgentThreadProgressEvent{
			EventType:       AgentClientEventThreadProgress,
			SchemaVersion:   SchemaVersion,
			AgentID:         agentID,
			TaskID:          req.TaskID,
			ThreadID:        req.ThreadID,
			RunID:           req.RunID,
			WorkStatus:      AgentWorkStatusRunning,
			AssignmentState: AgentAssignmentStateRunning,
			CommentKind:     AgentTaskCommentRuntimeProgress,
			BatchStartedAt:  req.BatchStartedAt,
			BatchEndedAt:    req.BatchEndedAt,
			Lines:           lines,
		},
	})
}

func (s Server) handleProviderStatusSync(w http.ResponseWriter, r *http.Request, agentID string) {
	if s.provider == nil {
		writeError(w, http.StatusServiceUnavailable, "provider status store is not configured")
		return
	}
	if _, ok := s.authorizeRequest(w, r, AuthorizationRequest{Resource: AuthorizationResourceAgent, Action: AuthorizationActionProviderStatusWrite, AgentID: agentID}); !ok {
		return
	}
	var req ProviderStatusSyncRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	response, err := s.provider.SyncProviderStatus(r.Context(), agentID, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (s Server) handleProviderStatusRead(w http.ResponseWriter, r *http.Request, agentID string) {
	if s.providerRead == nil {
		writeError(w, http.StatusServiceUnavailable, "provider status reader is not configured")
		return
	}
	if _, ok := s.authorizeRequest(w, r, AuthorizationRequest{Resource: AuthorizationResourceAgent, Action: AuthorizationActionProviderStatusRead, AgentID: agentID}); !ok {
		return
	}
	response, found, err := s.providerRead.GetProviderStatus(r.Context(), agentID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if !found {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	writeJSON(w, http.StatusOK, response)
}
