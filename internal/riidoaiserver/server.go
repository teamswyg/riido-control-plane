package riidoaiserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/http"
	"strings"
	"time"
)

const (
	aiAgentTokenHeader = "X-Riido-Ai-Agent-Token"
	deviceIDHeader     = "X-Riido-Device-Id"
	deviceSecretHeader = "X-Riido-Device-Secret"
)

// sseKeepaliveInterval is how often an idle SSE stream emits a comment line to
// keep the connection alive through ALB/proxy idle timeouts.
const sseKeepaliveInterval = 15 * time.Second

type ServerConfig struct {
	Authorizer               RequestAuthorizer
	AgentCatalogStore        AgentCatalogStore
	AIAgentClient            AIAgentClientStore
	AIAgentProfileThumbnails AIAgentProfileThumbnailUploadService
	DeviceCredentials        DeviceCredentialStore
	Assignment               AssignmentStore
	TaskContext              AIAgentTaskContextReader
	ProviderStatus           ProviderStatusStore
	ProviderRead             ProviderStatusReader
	WebAllowedOrigins        []string
	// LongPollMaxHold caps how long a daemon claim poll (PollRequest.WaitMs) may
	// be held open. Zero applies the default (25s). Must stay well under the ALB
	// idle timeout (60s default) and the http.Server write/idle timeouts (unset).
	LongPollMaxHold time.Duration
	// LongPollTick is the fallback re-evaluation interval during a held poll. It
	// bounds cross-instance discovery latency (an assignment queued on another
	// control-plane instance). Zero applies the default (2s).
	LongPollTick time.Duration
}

type Server struct {
	assignment               AssignmentStore
	agentCatalog             AgentCatalogStore
	aiAgent                  AIAgentClientStore
	aiAgentProfileThumbnails AIAgentProfileThumbnailUploadService
	daemonRuntime            AIAgentDaemonRuntimeStore
	taskContext              AIAgentTaskContextReader
	provider                 ProviderStatusStore
	providerRead             ProviderStatusReader
	devices                  DeviceCredentialStore
	config                   ServerConfig
}

type aiAgentWorkspaceIDContextKey struct{}

func NewServer(config ServerConfig) Server {
	config.WebAllowedOrigins = normalizeWebAllowedOrigins(config.WebAllowedOrigins)
	if config.LongPollMaxHold <= 0 {
		config.LongPollMaxHold = 25 * time.Second
	}
	if config.LongPollTick <= 0 {
		config.LongPollTick = 2 * time.Second
	}
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
	devices := config.DeviceCredentials
	if devices == nil {
		if store, ok := config.AIAgentClient.(DeviceCredentialStore); ok {
			devices = store
		}
	}
	var daemonRuntime AIAgentDaemonRuntimeStore
	if store, ok := config.AIAgentClient.(AIAgentDaemonRuntimeStore); ok {
		daemonRuntime = store
	}
	return Server{
		assignment:               config.Assignment,
		agentCatalog:             agentCatalog,
		aiAgent:                  config.AIAgentClient,
		aiAgentProfileThumbnails: config.AIAgentProfileThumbnails,
		daemonRuntime:            daemonRuntime,
		taskContext:              config.TaskContext,
		provider:                 provider,
		providerRead:             providerRead,
		devices:                  devices,
		config:                   config,
	}
}

func (s Server) reconcileAIAgentTaskThreadProjections(ctx context.Context, principal AuthorizationResult, taskID string) error {
	reader, ok := s.assignment.(AssignmentProjectionReader)
	if !ok {
		return nil
	}
	reconciler, ok := s.aiAgent.(AIAgentTaskThreadProjectionReconciler)
	if !ok {
		return nil
	}
	_, err := reconciler.ReconcileAIAgentActiveThreadProjections(ctx, principal, taskID, reader)
	return err
}

func (s Server) cancelAIAgentAssignmentFromAction(ctx context.Context, response AIAgentTaskActionResponse, reason string) error {
	canceller, ok := s.assignment.(AssignmentCancellationStore)
	if !ok {
		return nil
	}
	taskID := strings.TrimSpace(response.TaskID)
	agentID := strings.TrimSpace(response.AgentID)
	assignmentID := strings.TrimSpace(response.AssignmentID)
	if taskID == "" || agentID == "" || assignmentID == "" {
		return nil
	}
	_, err := canceller.CancelAssignment(ctx, taskID, CancelAssignmentRequest{
		AgentID:      agentID,
		AssignmentID: assignmentID,
		Reason:       strings.TrimSpace(reason),
	})
	return err
}

func (s Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc("/readyz", s.handleReady)
	mux.HandleFunc("/metrics", s.handleMetrics)
	mux.HandleFunc("/v2/desktop/workspaces/", s.handleDesktopWorkspaceRoutes)
	mux.HandleFunc("/v2/client/workspaces/", s.handleAIAgentClientWorkspaceRoutes)
	mux.HandleFunc("/v1/client/ai-agent/bootstrap", s.handleAIAgentClientBootstrap)
	mux.HandleFunc("/v1/client/ai-agent/devices", s.handleAIAgentClientDevices)
	mux.HandleFunc("/v1/client/ai-agent/devices/", s.handleAIAgentClientDeviceRoutes)
	mux.HandleFunc("/v1/client/ai-agent/onboarding/fixtures", s.handleAIAgentClientOnboardingFixtures)
	mux.HandleFunc("/v1/client/ai-agent/onboarding/fixtures/", s.handleAIAgentClientOnboardingFixtures)
	mux.HandleFunc("/v1/client/ai-agent/profile-thumbnails/uploads", s.handleAIAgentClientProfileThumbnailUpload)
	mux.HandleFunc("/v1/client/ai-agent/tasks/", s.handleAIAgentClientTasks)
	mux.HandleFunc("/v1/client/ai-agent/agents", s.handleAIAgentClientAgents)
	mux.HandleFunc("/v1/client/ai-agent/agents/", s.handleAIAgentClientAgents)
	mux.HandleFunc("/v1/client/ai-agent/events", s.handleAIAgentClientEvents)
	mux.HandleFunc("/v1/daemon/runtime-snapshot", s.handleDaemonRuntimeSnapshot)
	mux.HandleFunc("/v1/daemon/agent-bindings", s.handleDaemonAgentBindings)
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

func (s Server) handleDesktopWorkspaceRoutes(w http.ResponseWriter, r *http.Request) {
	workspaceID, suffix, ok := splitNestedResourcePath(r.URL.Path, "/v2/desktop/workspaces/")
	if !ok || strings.TrimSpace(workspaceID) == "" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	suffix = strings.Trim(suffix, "/")
	switch {
	case suffix == "devices/enroll" && r.Method == http.MethodPost:
		s.handleDesktopDeviceEnroll(w, r, workspaceID)
	case suffix == "devices/connect" && r.Method == http.MethodPost:
		s.handleDesktopDeviceConnect(w, r, workspaceID)
	default:
		writeError(w, http.StatusNotFound, "not found")
	}
}

// handleDesktopDeviceConnect connects the caller's machine device to this
// workspace so every member of the workspace can see the device and its
// runtimes. Authorized by the user token (workspace membership); identifies the
// device by machine_id. Does not issue/rotate the device secret.
func (s Server) handleDesktopDeviceConnect(w http.ResponseWriter, r *http.Request, workspaceID string) {
	if s.daemonRuntime == nil {
		writeError(w, http.StatusServiceUnavailable, "daemon runtime store is not configured")
		return
	}
	r = requestWithAIAgentWorkspaceIDAndPath(r, workspaceID, r.URL.Path)
	principal, ok := s.authorizeAIAgentClient(w, r, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionDeviceRead})
	if !ok {
		return
	}
	var req struct {
		MachineID string `json:"machine_id"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if strings.TrimSpace(req.MachineID) == "" {
		writeError(w, http.StatusBadRequest, "machine_id is required")
		return
	}
	device, err := s.daemonRuntime.ConnectAIAgentDevice(r.Context(), principal, req.MachineID)
	if err != nil {
		if errors.Is(err, ErrAuthorizationForbidden) {
			writeError(w, http.StatusForbidden, "forbidden")
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"schema_version":          SchemaVersion,
		"device_id":               device.DeviceID,
		"connected_workspace_ids": device.ConnectedWorkspaceIDs,
	})
}

func (s Server) handleDesktopDeviceEnroll(w http.ResponseWriter, r *http.Request, workspaceID string) {
	if s.devices == nil {
		writeError(w, http.StatusServiceUnavailable, "device credential store is not configured")
		return
	}
	r = requestWithAIAgentWorkspaceIDAndPath(r, workspaceID, r.URL.Path)
	principal, ok := s.authorizeAIAgentClient(w, r, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionCreate})
	if !ok {
		return
	}
	var req EnrollDeviceRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	response, err := s.devices.EnrollDeviceCredential(r.Context(), principal, workspaceID, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, response)
}

func (s Server) handleDaemonRuntimeSnapshot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w)
		return
	}
	if s.daemonRuntime == nil {
		writeError(w, http.StatusServiceUnavailable, "daemon runtime store is not configured")
		return
	}
	deviceID := strings.TrimSpace(r.Header.Get(deviceIDHeader))
	if deviceID == "" {
		writeUnauthorized(w)
		return
	}
	principal, ok := s.authorizeRequest(w, r, AuthorizationRequest{
		Resource: AuthorizationResourceAgent,
		Action:   AuthorizationActionProviderStatusWrite,
		DeviceID: deviceID,
	})
	if !ok {
		return
	}
	var req DeviceRuntimeSnapshotSyncRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	req.DeviceID = strings.TrimSpace(req.DeviceID)
	if req.DeviceID == "" {
		req.DeviceID = deviceID
	}
	if req.DeviceID != deviceID {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}
	response, err := s.daemonRuntime.SyncAIAgentDaemonRuntimeSnapshot(r.Context(), principal, req)
	if err != nil {
		if errors.Is(err, ErrAuthorizationForbidden) {
			writeError(w, http.StatusForbidden, "forbidden")
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusAccepted, response)
}

func (s Server) handleDaemonAgentBindings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	if s.daemonRuntime == nil {
		writeError(w, http.StatusServiceUnavailable, "daemon runtime store is not configured")
		return
	}
	deviceID := strings.TrimSpace(r.Header.Get(deviceIDHeader))
	if deviceID == "" {
		writeUnauthorized(w)
		return
	}
	principal, ok := s.authorizeRequest(w, r, AuthorizationRequest{
		Resource: AuthorizationResourceAgent,
		Action:   AuthorizationActionRead,
		DeviceID: deviceID,
	})
	if !ok {
		return
	}
	response, err := s.daemonRuntime.ListAIAgentDaemonAgentBindings(r.Context(), principal, deviceID)
	if err != nil {
		if errors.Is(err, ErrAuthorizationForbidden) {
			writeError(w, http.StatusForbidden, "forbidden")
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (s Server) handleAIAgentClientWorkspaceRoutes(w http.ResponseWriter, r *http.Request) {
	workspaceID, v1Path, ok := splitAIAgentClientWorkspacePath(r.URL.Path)
	if !ok {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	r = requestWithAIAgentWorkspaceIDAndPath(r, workspaceID, v1Path)
	switch {
	case v1Path == "/v1/client/ai-agent/bootstrap":
		s.handleAIAgentClientBootstrap(w, r)
	case v1Path == "/v1/client/ai-agent/devices":
		s.handleAIAgentClientDevices(w, r)
	case strings.HasPrefix(v1Path, "/v1/client/ai-agent/devices/"):
		s.handleAIAgentClientDeviceRoutes(w, r)
	case v1Path == "/v1/client/ai-agent/onboarding/fixtures" || strings.HasPrefix(v1Path, "/v1/client/ai-agent/onboarding/fixtures/"):
		s.handleAIAgentClientOnboardingFixtures(w, r)
	case v1Path == "/v1/client/ai-agent/profile-thumbnails/uploads":
		s.handleAIAgentClientProfileThumbnailUpload(w, r)
	case v1Path == "/v1/client/ai-agent/tasks/assigned-agent-profiles":
		s.handleAIAgentClientWorkspaceAssignedAgentProfiles(w, r)
	case strings.HasPrefix(v1Path, "/v1/client/ai-agent/tasks/"):
		s.handleAIAgentClientTasks(w, r)
	case v1Path == "/v1/client/ai-agent/agents" || strings.HasPrefix(v1Path, "/v1/client/ai-agent/agents/"):
		s.handleAIAgentClientAgents(w, r)
	case v1Path == "/v1/client/ai-agent/events":
		s.handleAIAgentClientEvents(w, r)
	default:
		writeError(w, http.StatusNotFound, "not found")
	}
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
	if err := s.reconcileAIAgentTaskThreadProjections(r.Context(), principal, ""); err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	assignmentReq, err := s.assignRequestFromAIAgentClientTask(r.Context(), principal, bearerToken, taskID, req)
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
	if err := s.reconcileAIAgentTaskThreadProjections(r.Context(), principal, ""); err != nil {
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
	req.AssignmentID = assignment.ID
	response, err := s.aiAgent.CreateAIAgentTaskAgentAssignment(r.Context(), principal, taskID, req)
	if err != nil {
		writeAIAgentClientError(w, err)
		return
	}
	writeJSON(w, http.StatusAccepted, response)
}

// agentResponseGuidelines is appended to every assignment's agent instruction
// (which the daemon turns into the runtime system prompt). It makes the agent
// answer in Korean and, when there is no codebase/workspace or tool use is
// restricted, respond helpfully instead of hard-failing: answer questions
// directly, and for work that would need a workspace, ask the user where to do
// it (path/repository) or propose concrete next steps rather than spamming
// tools or ending with a failure.
const agentResponseGuidelines = `[응답 지침]
- 모든 답변은 한국어로 작성하세요.
- 작업 디렉터리에 코드/리포지터리가 없거나 도구 사용이 제한되어 실제 파일 작업을 할 수 없더라도, "작업 공간이 없어 실패했다"로 끝내지 마세요.
  - 질문이거나 설명/조언으로 충분한 요청이면 도구 없이 바로 한국어로 답하세요.
  - 파일·코드 작업이 꼭 필요하면 도구를 무리하게 반복 호출하지 말고, 사용자에게 "어느 경로 또는 리포지터리에서 작업할까요?"처럼 필요한 정보를 물어보거나, 이후 진행해야 할 일을 단계로 제시하며 마무리하세요.`

// augmentAgentInstruction appends the shared response guidelines to an agent's
// own instruction before it becomes the runtime system prompt.
func augmentAgentInstruction(instruction string) string {
	instruction = strings.TrimSpace(instruction)
	if instruction == "" {
		return agentResponseGuidelines
	}
	return instruction + "\n\n" + agentResponseGuidelines
}

func (s Server) assignRequestFromAIAgentClientTask(ctx context.Context, principal AuthorizationResult, bearerToken, taskID string, req AssignAIAgentTaskRequest) (AssignRequest, error) {
	taskID = strings.TrimSpace(taskID)
	req.AgentID = strings.TrimSpace(req.AgentID)
	if taskID == "" {
		return AssignRequest{}, errors.New("task_id is required")
	}
	if req.AgentID == "" {
		return AssignRequest{}, errors.New("agent_id is required")
	}
	agents, err := s.aiAgent.ListAIAgentTaskAssignableAgents(ctx, principal, taskID)
	if err != nil {
		return AssignRequest{}, err
	}
	var selected AgentClientRecord
	for _, agent := range agents.Agents {
		if agent.AgentID == req.AgentID {
			selected = agent
			break
		}
	}
	if selected.AgentID == "" {
		return AssignRequest{}, ErrAIAgentNotFound
	}
	var binding AgentRuntimeBinding
	var runtime RuntimeRecord
	if factRegistry, ok := s.aiAgent.(AgentRuntimeFactRegistry); ok {
		var found bool
		binding, runtime, found = factRegistry.LookupAgentRuntimeFact(selected.AgentID)
		if !found {
			return AssignRequest{}, errors.New("ai agent runtime binding is not configured")
		}
	} else {
		registry, ok := s.aiAgent.(AgentRegistry)
		if !ok {
			return AssignRequest{}, errors.New("ai agent runtime registry is not configured")
		}
		var found bool
		binding, found = registry.LookupAgent(selected.AgentID)
		if !found {
			return AssignRequest{}, errors.New("ai agent runtime binding is not configured")
		}
	}
	assignmentReq := AssignRequest{
		ComponentID:              taskID,
		AgentID:                  selected.AgentID,
		RuntimeProvider:          binding.RuntimeProvider,
		ModelID:                  selected.ModelID,
		AgentInstruction:         augmentAgentInstruction(selected.Instruction),
		AllowExperimentalRuntime: runtime.RequiresExperimentalOptIn,
		CreatedBy:                strings.TrimSpace(principal.PrincipalID),
	}
	return s.assignRequestWithTaskContextPromptForClient(ctx, taskID, assignmentReq, AIAgentTaskContextRequest{
		ComponentID: taskID,
		WorkspaceID: principal.WorkspaceID,
		BearerToken: bearerToken,
	})
}

func (s Server) assignRequestFromAIAgentTaskThreadMessage(ctx context.Context, principal AuthorizationResult, bearerToken, taskID, threadID string, req CreateAIAgentTaskThreadMessageRequest) (AssignRequest, error) {
	taskID = strings.TrimSpace(taskID)
	threadID = strings.TrimSpace(threadID)
	req.Body = strings.TrimSpace(req.Body)
	if taskID == "" {
		return AssignRequest{}, errors.New("task_id is required")
	}
	if threadID == "" {
		return AssignRequest{}, errors.New("thread_id is required")
	}
	if req.Body == "" {
		return AssignRequest{}, errors.New("body is required")
	}
	if err := s.reconcileAIAgentTaskThreadProjections(ctx, principal, taskID); err != nil {
		return AssignRequest{}, err
	}
	threads, err := s.aiAgent.ListAIAgentTaskThreads(ctx, principal, taskID)
	if err != nil {
		return AssignRequest{}, err
	}
	if threads.ActiveStream != nil && threads.ActiveStream.ThreadID != threadID {
		return AssignRequest{}, ErrAIAgentTaskThreadConflict
	}
	var selectedThread AIAgentTaskThreadRecord
	for _, thread := range threads.Threads {
		if thread.ThreadID == threadID {
			selectedThread = thread
			break
		}
	}
	if selectedThread.ThreadID == "" {
		return AssignRequest{}, ErrAIAgentNotFound
	}
	assignmentReq, err := s.assignRequestFromAIAgentClientTask(ctx, principal, bearerToken, taskID, AssignAIAgentTaskRequest{
		AgentID: selectedThread.AgentID,
	})
	if err != nil {
		return AssignRequest{}, err
	}
	assignmentReq.Prompt = appendAIAgentTaskThreadMessagePrompt(assignmentReq.Prompt, selectedThread, req)
	return assignmentReq, nil
}

func appendAIAgentTaskThreadMessagePrompt(prompt string, thread AIAgentTaskThreadRecord, req CreateAIAgentTaskThreadMessageRequest) string {
	var b strings.Builder
	b.WriteString(strings.TrimSpace(prompt))
	b.WriteString("\n\n## Follow-up Thread Message\n")
	b.WriteString("- thread_id: ")
	b.WriteString(strings.TrimSpace(thread.ThreadID))
	b.WriteString("\n- previous_run_id: ")
	b.WriteString(strings.TrimSpace(thread.RunID))
	b.WriteString("\n- previous_work_status: ")
	b.WriteString(string(thread.WorkStatus))
	b.WriteString("\n- previous_assignment_state: ")
	b.WriteString(string(thread.AssignmentState))
	if sourceMessageID := strings.TrimSpace(req.SourceMessageID); sourceMessageID != "" {
		b.WriteString("\n- source_message_id: ")
		b.WriteString(sourceMessageID)
	}
	if previousMessage := strings.TrimSpace(thread.Message); previousMessage != "" {
		b.WriteString("\n\n### Previous Thread Message\n")
		b.WriteString(previousMessage)
	}
	b.WriteString("\n\n### New User Instruction\n")
	b.WriteString(strings.TrimSpace(req.Body))
	return strings.TrimSpace(b.String())
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
	if err := s.cancelAIAgentAssignmentFromAction(r.Context(), response, req.Reason); err != nil {
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
	response, err := s.aiAgent.DeleteAIAgentTaskAgentAssignment(r.Context(), principal, taskID, agentID, req)
	if err != nil {
		writeAIAgentClientError(w, err)
		return
	}
	if err := s.cancelAIAgentAssignmentFromAction(r.Context(), response, req.Reason); err != nil {
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
	response, err := s.aiAgent.StopAIAgentTask(r.Context(), principal, taskID, req)
	if err != nil {
		writeAIAgentClientError(w, err)
		return
	}
	if err := s.cancelAIAgentAssignmentFromAction(r.Context(), response, req.Reason); err != nil {
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
	response, err := s.aiAgent.StopAIAgentTaskAgentAssignment(r.Context(), principal, taskID, agentID, req)
	if err != nil {
		writeAIAgentClientError(w, err)
		return
	}
	if err := s.cancelAIAgentAssignmentFromAction(r.Context(), response, req.Reason); err != nil {
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
	return composeAssignRequestWithTaskContext(taskID, componentID, req, contextSnapshot)
}

func (s Server) assignRequestWithTaskContextPromptForClient(ctx context.Context, taskID string, req AssignRequest, contextReq AIAgentTaskContextRequest) (AssignRequest, error) {
	componentID := strings.TrimSpace(req.ComponentID)
	if componentID == "" {
		componentID = strings.TrimSpace(taskID)
	}
	contextReq.ComponentID = strings.TrimSpace(contextReq.ComponentID)
	if contextReq.ComponentID == "" {
		contextReq.ComponentID = componentID
	}
	contextSnapshot, err := s.getAIAgentTaskContextForRequest(ctx, contextReq)
	if err != nil {
		return AssignRequest{}, err
	}
	return composeAssignRequestWithTaskContext(taskID, componentID, req, contextSnapshot)
}

func (s Server) getAIAgentTaskContextForRequest(ctx context.Context, req AIAgentTaskContextRequest) (AIAgentTaskContext, error) {
	if s.taskContext == nil {
		return AIAgentTaskContext{}, errors.New("task context reader is not configured")
	}
	if reader, ok := s.taskContext.(AIAgentTaskContextRequestReader); ok {
		return reader.GetAIAgentTaskContextForRequest(ctx, req)
	}
	return s.taskContext.GetAIAgentTaskContext(ctx, req.ComponentID)
}

func composeAssignRequestWithTaskContext(taskID, componentID string, req AssignRequest, contextSnapshot AIAgentTaskContext) (AssignRequest, error) {
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
		metadata["thread_progress_seq"] = fmt.Sprint(line.Seq)
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
	req.WorkspaceID = strings.TrimSpace(aiAgentWorkspaceIDFromRequest(r))
	result, ok := s.authorizeRequest(w, r, req)
	if !ok {
		return AuthorizationResult{}, false
	}
	if result.WorkspaceID == "" {
		result.WorkspaceID = req.WorkspaceID
	}
	if strings.TrimSpace(result.PrincipalID) == "" {
		writeError(w, http.StatusForbidden, "forbidden")
		return AuthorizationResult{}, false
	}
	return result, true
}

func (s Server) authorizeRequest(w http.ResponseWriter, r *http.Request, req AuthorizationRequest) (AuthorizationResult, bool) {
	if req.Resource == AuthorizationResourceAgent {
		if result, handled := s.authorizeDeviceCredential(w, r, req); handled {
			if strings.TrimSpace(result.PrincipalID) == "" {
				return AuthorizationResult{}, false
			}
			return result, true
		}
	}
	if s.config.Authorizer == nil {
		writeError(w, http.StatusServiceUnavailable, "scoped request authorizer is not configured")
		return AuthorizationResult{}, false
	}
	token, ok := requestToken(r)
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

func (s Server) authorizeDeviceCredential(w http.ResponseWriter, r *http.Request, req AuthorizationRequest) (AuthorizationResult, bool) {
	deviceID := strings.TrimSpace(r.Header.Get(deviceIDHeader))
	deviceSecret := strings.TrimSpace(r.Header.Get(deviceSecretHeader))
	if deviceID == "" && deviceSecret == "" {
		return AuthorizationResult{}, false
	}
	if deviceID == "" || deviceSecret == "" {
		writeUnauthorized(w)
		return AuthorizationResult{}, true
	}
	if s.devices == nil {
		writeError(w, http.StatusServiceUnavailable, "device credential store is not configured")
		return AuthorizationResult{}, true
	}
	if req.DeviceID == "" {
		req.DeviceID = deviceID
	}
	result, err := s.devices.AuthorizeDeviceCredential(r.Context(), deviceID, deviceSecret, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrAuthorizationForbidden):
			writeError(w, http.StatusForbidden, "forbidden")
		case errors.Is(err, ErrAuthorizationUnauthenticated):
			writeUnauthorized(w)
		default:
			writeError(w, http.StatusServiceUnavailable, err.Error())
		}
		return AuthorizationResult{}, true
	}
	return result, true
}

func requestToken(r *http.Request) (string, bool) {
	if token := strings.TrimSpace(r.Header.Get(aiAgentTokenHeader)); token != "" {
		return token, true
	}
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

func aiAgentWorkspaceIDFromRequest(r *http.Request) string {
	if r == nil {
		return ""
	}
	if value, ok := r.Context().Value(aiAgentWorkspaceIDContextKey{}).(string); ok {
		return strings.TrimSpace(value)
	}
	return ""
}

func requestWithAIAgentWorkspaceIDAndPath(r *http.Request, workspaceID, path string) *http.Request {
	urlCopy := *r.URL
	urlCopy.Path = path
	next := r.Clone(context.WithValue(r.Context(), aiAgentWorkspaceIDContextKey{}, strings.TrimSpace(workspaceID)))
	next.URL = &urlCopy
	return next
}

func splitAIAgentClientWorkspacePath(path string) (string, string, bool) {
	workspaceID, suffix, ok := splitNestedResourcePath(path, "/v2/client/workspaces/")
	if !ok || strings.TrimSpace(workspaceID) == "" {
		return "", "", false
	}
	suffix = strings.Trim(suffix, "/")
	switch {
	case suffix == "ai-agent":
		return workspaceID, "/v1/client/ai-agent", true
	case strings.HasPrefix(suffix, "ai-agent/"):
		return workspaceID, "/v1/client/ai-agent/" + strings.TrimPrefix(suffix, "ai-agent/"), true
	default:
		return "", "", false
	}
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

func splitNestedResourcePath(path, prefix string) (string, string, bool) {
	if !strings.HasPrefix(path, prefix) {
		return "", "", false
	}
	rest := strings.Trim(strings.TrimPrefix(path, prefix), "/")
	parts := strings.SplitN(rest, "/", 2)
	if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
		return "", "", false
	}
	return parts[0], parts[1], true
}

func threadMessageSuffixThreadID(suffix string) (string, bool) {
	parts := strings.Split(strings.Trim(suffix, "/"), "/")
	if len(parts) != 3 || parts[0] != "threads" || strings.TrimSpace(parts[1]) == "" || parts[2] != "messages" {
		return "", false
	}
	return parts[1], true
}

func agentAssignmentSuffixAgentID(suffix string) (string, bool) {
	parts := strings.Split(strings.Trim(suffix, "/"), "/")
	if len(parts) != 2 || parts[0] != "agent-assignments" || strings.TrimSpace(parts[1]) == "" {
		return "", false
	}
	return parts[1], true
}

func agentAssignmentStopSuffixAgentID(suffix string) (string, bool) {
	parts := strings.Split(strings.Trim(suffix, "/"), "/")
	if len(parts) != 3 || parts[0] != "agent-assignments" || strings.TrimSpace(parts[1]) == "" || parts[2] != "stop" {
		return "", false
	}
	return parts[1], true
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
	if err := json.NewEncoder(w).Encode(value); err != nil {
		return
	}
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
	maps.Copy(out, in)
	return out
}

func writeAIAgentClientError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrAIAgentNotFound):
		writeError(w, http.StatusNotFound, "not found")
	case errors.Is(err, ErrAIAgentAssigned):
		writeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, ErrAIAgentTaskThreadConflict):
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
	keepalive := time.NewTicker(sseKeepaliveInterval)
	defer keepalive.Stop()
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
		case <-keepalive.C:
			if _, err := io.WriteString(w, ": keepalive\n\n"); err != nil {
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
