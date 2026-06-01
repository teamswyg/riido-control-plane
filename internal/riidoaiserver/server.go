package riidoaiserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type ServerConfig struct {
	Authorizer        RequestAuthorizer
	AgentCatalogStore AgentCatalogStore
	AIAgentClient     AIAgentClientStore
	Assignment        AssignmentStore
	TaskContext       AIAgentTaskContextReader
	ProviderStatus    ProviderStatusStore
	ProviderRead      ProviderStatusReader
	WebAllowedOrigins []string
}

type Server struct {
	assignment   AssignmentStore
	agentCatalog AgentCatalogStore
	aiAgent      AIAgentClientStore
	taskContext  AIAgentTaskContextReader
	provider     ProviderStatusStore
	providerRead ProviderStatusReader
	config       ServerConfig
}

func NewServer(config ServerConfig) Server {
	config.WebAllowedOrigins = normalizeWebAllowedOrigins(config.WebAllowedOrigins)
	agentCatalog := config.AgentCatalogStore
	if agentCatalog == nil {
		if store, ok := config.Assignment.(AgentCatalogStore); ok {
			agentCatalog = store
		}
	}
	provider := config.ProviderStatus
	if provider == nil {
		if store, ok := config.Assignment.(ProviderStatusStore); ok {
			provider = store
		}
	}
	providerRead := config.ProviderRead
	if providerRead == nil {
		if reader, ok := provider.(ProviderStatusReader); ok {
			providerRead = reader
		}
	}
	return Server{
		assignment:   config.Assignment,
		agentCatalog: agentCatalog,
		aiAgent:      config.AIAgentClient,
		taskContext:  config.TaskContext,
		provider:     provider,
		providerRead: providerRead,
		config:       config,
	}
}

func (s Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc("/readyz", s.handleReady)
	mux.HandleFunc("/metrics", s.handleMetrics)
	mux.HandleFunc("/v1/client/ai-agent/bootstrap", s.handleAIAgentClientBootstrap)
	mux.HandleFunc("/v1/client/ai-agent/devices", s.handleAIAgentClientDevices)
	mux.HandleFunc("/v1/client/ai-agent/devices/", s.handleAIAgentClientDeviceRoutes)
	mux.HandleFunc("/v1/client/ai-agent/tasks/", s.handleAIAgentClientTasks)
	mux.HandleFunc("/v1/client/ai-agent/agents", s.handleAIAgentClientAgents)
	mux.HandleFunc("/v1/client/ai-agent/agents/", s.handleAIAgentClientAgents)
	mux.HandleFunc("/v1/client/ai-agent/events", s.handleAIAgentClientEvents)
	mux.HandleFunc("/v1/agent-catalog", s.handleAgentCatalog)
	mux.HandleFunc("/v1/agent-catalog/", s.handleAgentCatalog)
	mux.HandleFunc("/v1/component-tasks/", s.handleComponentTasks)
	mux.HandleFunc("/v1/agents/", s.handleAgents)
	var handler http.Handler = mux
	if len(s.config.WebAllowedOrigins) > 0 {
		handler = s.withWebFrontendCORS(handler)
	}
	return handler
}

func (s Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusOK, Health{SchemaVersion: SchemaVersion, Status: "ok"})
}

func (s Server) handleReady(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusOK, Health{SchemaVersion: SchemaVersion, Status: "ready"})
}

func (s Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	if s.assignment == nil {
		writeError(w, http.StatusServiceUnavailable, "assignment store is not configured")
		return
	}
	if _, ok := s.authorizeRequest(w, r, AuthorizationRequest{Resource: AuthorizationResourceMetrics, Action: AuthorizationActionRead}); !ok {
		return
	}
	snapshot, err := s.assignment.Metrics(r.Context())
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, snapshot)
}

func (s Server) handleAgentCatalog(w http.ResponseWriter, r *http.Request) {
	if s.agentCatalog == nil {
		writeError(w, http.StatusServiceUnavailable, "agent catalog store is not configured")
		return
	}
	agentID, hasAgentID, ok := splitOptionalResourcePath(r.URL.Path, "/v1/agent-catalog")
	if !ok {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	switch {
	case !hasAgentID && r.Method == http.MethodGet:
		s.handleAgentCatalogList(w, r)
	case !hasAgentID && r.Method == http.MethodPost:
		s.handleAgentCatalogCreate(w, r)
	case hasAgentID && r.Method == http.MethodGet:
		s.handleAgentCatalogGet(w, r, agentID)
	case hasAgentID && r.Method == http.MethodPatch:
		s.handleAgentCatalogUpdate(w, r, agentID)
	case hasAgentID && r.Method == http.MethodDelete:
		s.handleAgentCatalogDelete(w, r, agentID)
	default:
		writeMethodNotAllowed(w)
	}
}

func (s Server) handleAIAgentClientBootstrap(w http.ResponseWriter, r *http.Request) {
	if s.aiAgent == nil {
		writeError(w, http.StatusServiceUnavailable, "ai agent client mock is not configured")
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
	response, err := s.aiAgent.BootstrapAIAgentClient(r.Context(), principal, ClientKind(strings.TrimSpace(r.URL.Query().Get("client_kind"))))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (s Server) handleAIAgentClientDevices(w http.ResponseWriter, r *http.Request) {
	if s.aiAgent == nil {
		writeError(w, http.StatusServiceUnavailable, "ai agent client mock is not configured")
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

func (s Server) handleAIAgentClientDeviceRoutes(w http.ResponseWriter, r *http.Request) {
	if s.aiAgent == nil {
		writeError(w, http.StatusServiceUnavailable, "ai agent client mock is not configured")
		return
	}
	_, suffix, ok := splitAIAgentClientDevicePath(r.URL.Path)
	if !ok {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	switch {
	case suffix == "":
		writeMethodNotAllowed(w)
	default:
		writeError(w, http.StatusNotFound, "not found")
	}
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
		writeError(w, http.StatusServiceUnavailable, "ai agent client mock is not configured")
		return
	}
	taskID, suffix, ok := splitResourcePath(r.URL.Path, "/v1/client/ai-agent/tasks/")
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
	case suffix == "threads" && r.Method == http.MethodGet:
		s.handleAIAgentClientTaskThreads(w, r, taskID)
	case suffix == "comments" && r.Method == http.MethodPost:
		s.handleAIAgentClientSubmitTaskComment(w, r, taskID)
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

func (s Server) handleAIAgentClientAssignTask(w http.ResponseWriter, r *http.Request, taskID string) {
	principal, ok := s.authorizeAIAgentClient(w, r, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionAssign, TaskID: taskID})
	if !ok {
		return
	}
	var req AssignAIAgentTaskRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	response, err := s.aiAgent.AssignAIAgentTask(r.Context(), principal, taskID, req)
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
	response, err := s.aiAgent.UnassignAIAgentTask(r.Context(), principal, taskID, req)
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
	response, err := s.aiAgent.ListAIAgentTaskThreads(r.Context(), principal, taskID)
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
	response, err := s.aiAgent.StopAIAgentTask(r.Context(), principal, taskID, req)
	if err != nil {
		writeAIAgentClientError(w, err)
		return
	}
	writeJSON(w, http.StatusAccepted, response)
}

func (s Server) handleAIAgentClientAgents(w http.ResponseWriter, r *http.Request) {
	if s.aiAgent == nil {
		writeError(w, http.StatusServiceUnavailable, "ai agent client mock is not configured")
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
		writeError(w, http.StatusServiceUnavailable, "ai agent client mock is not configured")
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
		case <-r.Context().Done():
			return
		}
	}
}

func (s Server) handleAgentCatalogList(w http.ResponseWriter, r *http.Request) {
	principal, ok := s.authorizeAgentCatalog(w, r, AuthorizationRequest{Resource: AuthorizationResourceAgentCatalog, Action: AuthorizationActionRead})
	if !ok {
		return
	}
	records, err := s.agentCatalog.ListAgentCatalog(r.Context())
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, AgentCatalogListResponse{
		SchemaVersion: SchemaVersion,
		Agents:        VisibleAgentCatalogRecords(principal, records),
	})
}

func (s Server) handleAgentCatalogCreate(w http.ResponseWriter, r *http.Request) {
	principal, ok := s.authorizeAgentCatalog(w, r, AuthorizationRequest{Resource: AuthorizationResourceAgentCatalog, Action: AuthorizationActionCreate})
	if !ok {
		return
	}
	var req CreateAgentCatalogRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	record := AgentCatalogRecord{
		AgentID:          req.AgentID,
		OwnerPrincipalID: principal.PrincipalID,
		Visibility:       req.Visibility,
	}
	saved, err := s.agentCatalog.SaveAgentCatalog(r.Context(), record)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, AgentCatalogRecordResponse{SchemaVersion: SchemaVersion, Agent: saved})
}

func (s Server) handleAgentCatalogGet(w http.ResponseWriter, r *http.Request, agentID string) {
	principal, ok := s.authorizeAgentCatalog(w, r, AuthorizationRequest{Resource: AuthorizationResourceAgentCatalog, Action: AuthorizationActionRead, AgentID: agentID})
	if !ok {
		return
	}
	record, found, err := s.agentCatalog.GetAgentCatalog(r.Context(), agentID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if !found {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	decision := EvaluateAgentCatalogAccess(principal, record, AgentCatalogActionRead)
	if !decision.Allowed {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}
	writeJSON(w, http.StatusOK, AgentCatalogRecordResponse{SchemaVersion: SchemaVersion, Agent: record})
}

func (s Server) handleAgentCatalogUpdate(w http.ResponseWriter, r *http.Request, agentID string) {
	principal, ok := s.authorizeAgentCatalog(w, r, AuthorizationRequest{Resource: AuthorizationResourceAgentCatalog, Action: AuthorizationActionUpdate, AgentID: agentID})
	if !ok {
		return
	}
	record, found, err := s.agentCatalog.GetAgentCatalog(r.Context(), agentID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if !found {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	decision := EvaluateAgentCatalogAccess(principal, record, AgentCatalogActionUpdate)
	if !decision.Allowed {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}
	var req UpdateAgentCatalogRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	record.Visibility = req.Visibility
	saved, err := s.agentCatalog.SaveAgentCatalog(r.Context(), record)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, AgentCatalogRecordResponse{SchemaVersion: SchemaVersion, Agent: saved})
}

func (s Server) handleAgentCatalogDelete(w http.ResponseWriter, r *http.Request, agentID string) {
	principal, ok := s.authorizeAgentCatalog(w, r, AuthorizationRequest{Resource: AuthorizationResourceAgentCatalog, Action: AuthorizationActionDelete, AgentID: agentID})
	if !ok {
		return
	}
	record, found, err := s.agentCatalog.GetAgentCatalog(r.Context(), agentID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if !found {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	decision := EvaluateAgentCatalogAccess(principal, record, AgentCatalogActionDelete)
	if !decision.Allowed {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}
	deleted, err := s.agentCatalog.DeleteAgentCatalog(r.Context(), agentID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, struct {
		SchemaVersion string `json:"schema_version"`
		Deleted       bool   `json:"deleted"`
	}{SchemaVersion: SchemaVersion, Deleted: deleted})
}

func (s Server) handleAgents(w http.ResponseWriter, r *http.Request) {
	agentID, suffix, ok := splitResourcePath(r.URL.Path, "/v1/agents/")
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

func (s Server) assignRequestWithTaskContextPrompt(ctx context.Context, taskID string, req AssignRequest) (AssignRequest, error) {
	componentID := strings.TrimSpace(req.ComponentID)
	if componentID == "" {
		componentID = strings.TrimSpace(taskID)
	}
	contextSnapshot, err := s.taskContext.GetAIAgentTaskContext(ctx, componentID)
	if err != nil {
		return AssignRequest{}, err
	}
	composed, err := ComposeAIAgentAssignmentPrompt(AIAgentAssignmentPromptInput{
		TaskID:  taskID,
		Context: contextSnapshot,
	})
	if err != nil {
		return AssignRequest{}, err
	}
	req.Prompt = composed.Prompt
	if strings.TrimSpace(req.ComponentID) == "" {
		req.ComponentID = strings.TrimSpace(contextSnapshot.Component.ID)
		if req.ComponentID == "" {
			req.ComponentID = componentID
		}
	}
	return req, nil
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
	response, err := s.assignment.PollAgent(r.Context(), agentID, req)
	if err != nil {
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
	response, err := s.assignment.HeartbeatAgent(r.Context(), agentID, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
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
	if req.ThreadID == "" {
		req.ThreadID = threadIDForRun(req.TaskID, agentID, req.RunID)
	}
	lines := normalizeProgressLines(req.Lines)
	if len(lines) == 0 {
		writeError(w, http.StatusBadRequest, "lines are required")
		return
	}
	req.Lines = lines
	for _, line := range lines {
		metadata := copyStringMap(req.Metadata)
		metadata["thread_progress_seq"] = fmt.Sprint(line.Seq)
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

func (s Server) authorizeAgentCatalog(w http.ResponseWriter, r *http.Request, req AuthorizationRequest) (AgentCatalogPrincipal, bool) {
	result, ok := s.authorizeRequest(w, r, req)
	if !ok {
		return AgentCatalogPrincipal{}, false
	}
	principal := AgentCatalogPrincipalFromAuthorization(result)
	if err := principal.Validate(); err != nil {
		writeError(w, http.StatusForbidden, "forbidden")
		return AgentCatalogPrincipal{}, false
	}
	return principal, true
}

func (s Server) authorizeAIAgentClient(w http.ResponseWriter, r *http.Request, req AuthorizationRequest) (AuthorizationResult, bool) {
	result, ok := s.authorizeRequest(w, r, req)
	if !ok {
		return AuthorizationResult{}, false
	}
	if strings.TrimSpace(result.PrincipalID) == "" {
		writeError(w, http.StatusForbidden, "forbidden")
		return AuthorizationResult{}, false
	}
	return result, true
}

func (s Server) authorizeRequest(w http.ResponseWriter, r *http.Request, req AuthorizationRequest) (AuthorizationResult, bool) {
	if s.config.Authorizer == nil {
		writeError(w, http.StatusServiceUnavailable, "scoped request authorizer is not configured")
		return AuthorizationResult{}, false
	}
	token, ok := bearerToken(r)
	if !ok {
		writeUnauthorized(w)
		return AuthorizationResult{}, false
	}
	result, err := s.config.Authorizer.Authorize(r.Context(), token, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrAuthorizationForbidden):
			writeError(w, http.StatusForbidden, "forbidden")
		case errors.Is(err, ErrAuthorizationUnauthenticated):
			writeUnauthorized(w)
		default:
			writeError(w, http.StatusServiceUnavailable, err.Error())
		}
		return AuthorizationResult{}, false
	}
	return result, true
}

func bearerToken(r *http.Request) (string, bool) {
	got := strings.TrimSpace(r.Header.Get("Authorization"))
	if !strings.HasPrefix(got, "Bearer ") {
		return "", false
	}
	token := strings.TrimSpace(strings.TrimPrefix(got, "Bearer "))
	if token == "" {
		return "", false
	}
	return token, true
}

func splitOptionalResourcePath(path, prefix string) (string, bool, bool) {
	if path == prefix || path == prefix+"/" {
		return "", false, true
	}
	withSlash := prefix + "/"
	if !strings.HasPrefix(path, withSlash) {
		return "", false, false
	}
	rest := strings.Trim(strings.TrimPrefix(path, withSlash), "/")
	if rest == "" || strings.Contains(rest, "/") {
		return "", false, false
	}
	return rest, true, true
}

func splitResourcePath(path, prefix string) (string, string, bool) {
	if !strings.HasPrefix(path, prefix) {
		return "", "", false
	}
	rest := strings.Trim(strings.TrimPrefix(path, prefix), "/")
	parts := strings.Split(rest, "/")
	if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
		return "", "", false
	}
	return parts[0], parts[1], true
}

func splitAIAgentClientDevicePath(path string) (string, string, bool) {
	const prefix = "/v1/client/ai-agent/devices/"
	if !strings.HasPrefix(path, prefix) {
		return "", "", false
	}
	rest := strings.Trim(strings.TrimPrefix(path, prefix), "/")
	parts := strings.Split(rest, "/")
	if len(parts) < 2 || len(parts) > 3 || strings.TrimSpace(parts[0]) == "" {
		return "", "", false
	}
	for _, part := range parts[1:] {
		if strings.TrimSpace(part) == "" {
			return "", "", false
		}
	}
	return parts[0], strings.Join(parts[1:], "/"), true
}

func splitAIAgentClientAgentPath(path string) (string, string, bool) {
	const prefix = "/v1/client/ai-agent/agents/"
	if !strings.HasPrefix(path, prefix) {
		return "", "", false
	}
	rest := strings.Trim(strings.TrimPrefix(path, prefix), "/")
	if rest == "" {
		return "", "", false
	}
	parts := strings.Split(rest, "/")
	if len(parts) < 1 || len(parts) > 3 {
		return "", "", false
	}
	for _, part := range parts {
		if strings.TrimSpace(part) == "" {
			return "", "", false
		}
	}
	switch len(parts) {
	case 1:
		if strings.TrimSpace(parts[0]) == "" {
			return "", "", false
		}
		return parts[0], "", true
	default:
		return parts[0], strings.Join(parts[1:], "/"), true
	}
}

func decodeJSON(r *http.Request, out any) error {
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(out); err != nil {
		return fmt.Errorf("decode json: %w", err)
	}
	var extra struct{}
	if err := dec.Decode(&extra); err != io.EOF {
		return errors.New("decode json: trailing data")
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, struct {
		SchemaVersion string `json:"schema_version"`
		Error         string `json:"error"`
	}{SchemaVersion: SchemaVersion, Error: message})
}

func writeMethodNotAllowed(w http.ResponseWriter) {
	writeError(w, http.StatusMethodNotAllowed, "method not allowed")
}

func writeUnauthorized(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Bearer realm="riido_ai_server"`)
	writeError(w, http.StatusUnauthorized, "unauthorized")
}

func copyStringMap(in map[string]string) map[string]string {
	out := make(map[string]string, len(in)+1)
	for k, v := range in {
		out[k] = v
	}
	return out
}

func writeAIAgentClientError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrAIAgentNotFound):
		writeError(w, http.StatusNotFound, "not found")
	case errors.Is(err, ErrAIAgentAssigned):
		writeError(w, http.StatusConflict, err.Error())
	default:
		writeError(w, http.StatusBadRequest, err.Error())
	}
}

func (s Server) streamTaskEvents(w http.ResponseWriter, r *http.Request, taskID string) {
	history, events, cancel, err := s.assignment.SubscribeTask(r.Context(), taskID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer cancel()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	flusher, _ := w.(http.Flusher)
	for _, event := range history {
		if err := writeSSE(w, event); err != nil {
			return
		}
	}
	if flusher != nil {
		flusher.Flush()
	}
	if r.URL.Query().Get("replay") == "1" {
		return
	}
	for {
		select {
		case event, ok := <-events:
			if !ok {
				return
			}
			if err := writeSSE(w, event); err != nil {
				return
			}
			if flusher != nil {
				flusher.Flush()
			}
		case <-r.Context().Done():
			return
		}
	}
}

func writeSSE(w http.ResponseWriter, event TaskEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "id: %d\nevent: %s\ndata: %s\n\n", event.Seq, event.Type, data)
	return err
}

func writeAIAgentClientSSE(w http.ResponseWriter, event ClientStreamEvent) error {
	data, err := json.Marshal(event.Payload)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "id: %d\nevent: %s\ndata: %s\n\n", event.Seq, event.EventType, data)
	return err
}
