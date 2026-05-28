package riidoaiserver

import (
	"context"
	"errors"
	"sort"
	"strings"
	"sync"
	"time"
)

type AIAgentClientStore interface {
	BootstrapAIAgentClient(ctx context.Context, principal AuthorizationResult, clientKind ClientKind) (ClientBootstrapResponse, error)
	ListAIAgentDevices(ctx context.Context, principal AuthorizationResult) (DeviceRuntimeListResponse, error)
	ListAIAgentTaskAssignableAgents(ctx context.Context, principal AuthorizationResult, taskID string) (AgentClientListResponse, error)
	SubmitAIAgentTaskComment(ctx context.Context, principal AuthorizationResult, taskID string, req SubmitAIAgentTaskCommentRequest) (AIAgentTaskActionResponse, error)
	StopAIAgentTask(ctx context.Context, principal AuthorizationResult, taskID string, req StopAIAgentTaskRequest) (AIAgentTaskActionResponse, error)
	GetAIAgentEditability(ctx context.Context, principal AuthorizationResult, agentID string) (AgentEditabilityResponse, error)
	UpdateAIAgentConfiguration(ctx context.Context, principal AuthorizationResult, agentID string, req UpdateAgentConfigurationRequest) (AgentClientRecordResponse, error)
	DeleteAIAgent(ctx context.Context, principal AuthorizationResult, agentID string) (DeleteAgentResponse, error)
	AIAgentClientEvents(ctx context.Context, principal AuthorizationResult) ([]ClientStreamEvent, error)
}

type MockAIAgentClientStore struct {
	mu          sync.Mutex
	workspaceID string
	devices     []DeviceRecord
	agents      map[string]AgentClientRecord
	events      []ClientStreamEvent
}

func NewMockAIAgentClientStore() *MockAIAgentClientStore {
	now := time.Date(2026, 5, 28, 6, 0, 0, 0, time.UTC)
	device := DeviceRecord{
		DeviceID:         "device-mock-macbook",
		OwnerPrincipalID: "user-1",
		DisplayName:      "Mock MacBook Pro",
		DaemonLastSeenAt: now,
		Runtimes: []RuntimeRecord{
			{
				RuntimeID:        "runtime-codex-mock",
				DeviceID:         "device-mock-macbook",
				Kind:             RuntimeKindCodex,
				Availability:     RuntimeAvailabilityOnline,
				DetectionState:   RuntimeDetectionStateDetected,
				OwnerPrincipalID: "user-1",
				LastDetectedAt:   now,
				HasAssignedAgent: true,
			},
			{
				RuntimeID:        "runtime-claude-code-mock",
				DeviceID:         "device-mock-macbook",
				Kind:             RuntimeKindClaudeCode,
				Availability:     RuntimeAvailabilityOffline,
				DetectionState:   RuntimeDetectionStateMissing,
				OwnerPrincipalID: "user-1",
				LastDetectedAt:   now.Add(-30 * time.Second),
				HasAssignedAgent: true,
			},
			{
				RuntimeID:        "runtime-cursor-mock",
				DeviceID:         "device-mock-macbook",
				Kind:             RuntimeKindCursor,
				Availability:     RuntimeAvailabilityOnline,
				DetectionState:   RuntimeDetectionStateDetected,
				OwnerPrincipalID: "user-1",
				LastDetectedAt:   now,
				HasAssignedAgent: false,
			},
		},
	}
	agents := map[string]AgentClientRecord{
		"agent-owned-codex": {
			AgentID:           "agent-owned-codex",
			OwnerPrincipalID:  "user-1",
			Name:              "Codex 리뷰어",
			Visibility:        AgentVisibilityPrivate,
			RuntimeID:         "runtime-codex-mock",
			RuntimeKind:       RuntimeKindCodex,
			WorkStatus:        AgentWorkStatusRunning,
			Editability:       AgentEditabilityBlockedAssignedTasks,
			AssignedTaskCount: 1,
		},
		"agent-owned-claude": {
			AgentID:           "agent-owned-claude",
			OwnerPrincipalID:  "user-1",
			Name:              "Claude 설계 보조",
			Visibility:        AgentVisibilityPrivate,
			RuntimeID:         "runtime-claude-code-mock",
			RuntimeKind:       RuntimeKindClaudeCode,
			WorkStatus:        AgentWorkStatusOffline,
			Editability:       AgentEditabilityEditable,
			AssignedTaskCount: 0,
		},
		"agent-public-openclaw": {
			AgentID:           "agent-public-openclaw",
			OwnerPrincipalID:  "user-2",
			Name:              "OpenClaw 공개 에이전트",
			Visibility:        AgentVisibilityPublic,
			RuntimeID:         "runtime-openclaw-remote",
			RuntimeKind:       RuntimeKindOpenClaw,
			WorkStatus:        AgentWorkStatusIdle,
			Editability:       AgentEditabilityEditable,
			AssignedTaskCount: 0,
		},
		"agent-private-cursor": {
			AgentID:           "agent-private-cursor",
			OwnerPrincipalID:  "user-2",
			Name:              "Cursor 비공개 에이전트",
			Visibility:        AgentVisibilityPrivate,
			RuntimeID:         "runtime-cursor-mock",
			RuntimeKind:       RuntimeKindCursor,
			WorkStatus:        AgentWorkStatusIdle,
			Editability:       AgentEditabilityEditable,
			AssignedTaskCount: 0,
		},
	}
	return &MockAIAgentClientStore{
		workspaceID: "workspace-mock-riid",
		devices:     []DeviceRecord{device},
		agents:      agents,
		events: []ClientStreamEvent{
			{
				Seq:       1,
				EventType: "device_runtime_snapshot",
				Payload: DeviceRuntimeSnapshotEvent{
					EventType:     "device_runtime_snapshot",
					SchemaVersion: SchemaVersion,
					Device:        device,
				},
			},
			{
				Seq:       2,
				EventType: "agent_work_status_changed",
				Payload: AgentWorkStatusChangedEvent{
					EventType:       "agent_work_status_changed",
					SchemaVersion:   SchemaVersion,
					AgentID:         "agent-owned-codex",
					TaskID:          "task-mock-1",
					RunID:           "run-mock-1",
					WorkStatus:      AgentWorkStatusQueued,
					AssignmentState: AgentAssignmentStateQueued,
					CommentKind:     AgentTaskCommentQueuedByBusyAgent,
				},
			},
		},
	}
}

func (s *MockAIAgentClientStore) BootstrapAIAgentClient(ctx context.Context, principal AuthorizationResult, clientKind ClientKind) (ClientBootstrapResponse, error) {
	if err := ctx.Err(); err != nil {
		return ClientBootstrapResponse{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return ClientBootstrapResponse{
		SchemaVersion: SchemaVersion,
		ClientKind:    normalizeClientKind(clientKind),
		WorkspaceID:   s.workspaceID,
		Agents:        s.visibleAgents(principal),
		Devices:       copyDevices(s.devices),
	}, nil
}

func (s *MockAIAgentClientStore) ListAIAgentDevices(ctx context.Context, principal AuthorizationResult) (DeviceRuntimeListResponse, error) {
	if err := ctx.Err(); err != nil {
		return DeviceRuntimeListResponse{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return DeviceRuntimeListResponse{SchemaVersion: SchemaVersion, Devices: filterDevicesForPrincipal(s.devices, principal)}, nil
}

func (s *MockAIAgentClientStore) ListAIAgentTaskAssignableAgents(ctx context.Context, principal AuthorizationResult, taskID string) (AgentClientListResponse, error) {
	if err := ctx.Err(); err != nil {
		return AgentClientListResponse{}, err
	}
	if strings.TrimSpace(taskID) == "" {
		return AgentClientListResponse{}, errors.New("task_id is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return AgentClientListResponse{SchemaVersion: SchemaVersion, Agents: s.visibleAgents(principal)}, nil
}

func (s *MockAIAgentClientStore) SubmitAIAgentTaskComment(ctx context.Context, principal AuthorizationResult, taskID string, req SubmitAIAgentTaskCommentRequest) (AIAgentTaskActionResponse, error) {
	if err := ctx.Err(); err != nil {
		return AIAgentTaskActionResponse{}, err
	}
	taskID = strings.TrimSpace(taskID)
	req.AgentID = strings.TrimSpace(req.AgentID)
	req.Body = strings.TrimSpace(req.Body)
	if taskID == "" {
		return AIAgentTaskActionResponse{}, errors.New("task_id is required")
	}
	if req.AgentID == "" {
		return AIAgentTaskActionResponse{}, errors.New("agent_id is required")
	}
	if req.Body == "" {
		return AIAgentTaskActionResponse{}, errors.New("body is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	agent, ok := s.visibleAgent(principal, req.AgentID)
	if !ok {
		return AIAgentTaskActionResponse{}, ErrAIAgentNotFound
	}
	response := AIAgentTaskActionResponse{
		SchemaVersion:   SchemaVersion,
		TaskID:          taskID,
		AgentID:         agent.AgentID,
		RunID:           "run-mock-comment-" + taskID,
		WorkStatus:      AgentWorkStatusRunning,
		AssignmentState: AgentAssignmentStateRunning,
		CommentKind:     AgentTaskCommentRuntimeProgress,
		Message:         "agent work started from task comment",
	}
	if agent.WorkStatus == AgentWorkStatusRunning || agent.WorkStatus == AgentWorkStatusWaitingForUser || agent.WorkStatus == AgentWorkStatusQueued {
		response.WorkStatus = AgentWorkStatusQueued
		response.AssignmentState = AgentAssignmentStateQueued
		response.CommentKind = AgentTaskCommentQueuedByBusyAgent
		response.Message = "agent is busy; task comment was queued"
	}
	s.appendAgentTaskActionEvent(response)
	return response, nil
}

func (s *MockAIAgentClientStore) StopAIAgentTask(ctx context.Context, principal AuthorizationResult, taskID string, req StopAIAgentTaskRequest) (AIAgentTaskActionResponse, error) {
	if err := ctx.Err(); err != nil {
		return AIAgentTaskActionResponse{}, err
	}
	taskID = strings.TrimSpace(taskID)
	req.AgentID = strings.TrimSpace(req.AgentID)
	if taskID == "" {
		return AIAgentTaskActionResponse{}, errors.New("task_id is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	agent, ok := s.agentForTaskStop(principal, req.AgentID)
	if !ok {
		return AIAgentTaskActionResponse{}, ErrAIAgentNotFound
	}
	response := AIAgentTaskActionResponse{
		SchemaVersion:   SchemaVersion,
		TaskID:          taskID,
		AgentID:         agent.AgentID,
		RunID:           "run-mock-stop-" + taskID,
		WorkStatus:      AgentWorkStatusIdle,
		AssignmentState: AgentAssignmentStateStopped,
		CommentKind:     AgentTaskCommentStoppedByUserRequest,
		Message:         "agent work was stopped by user request",
	}
	s.appendAgentTaskActionEvent(response)
	return response, nil
}

func (s *MockAIAgentClientStore) GetAIAgentEditability(ctx context.Context, principal AuthorizationResult, agentID string) (AgentEditabilityResponse, error) {
	if err := ctx.Err(); err != nil {
		return AgentEditabilityResponse{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	agent, ok := s.visibleAgent(principal, agentID)
	if !ok {
		return AgentEditabilityResponse{}, ErrAIAgentNotFound
	}
	response := AgentEditabilityResponse{
		SchemaVersion:     SchemaVersion,
		AgentID:           agent.AgentID,
		Editability:       agent.Editability,
		AssignedTaskCount: agent.AssignedTaskCount,
	}
	if agent.Editability == AgentEditabilityBlockedAssignedTasks {
		response.Reason = "assigned_task_count must be zero before editing"
	}
	return response, nil
}

func (s *MockAIAgentClientStore) UpdateAIAgentConfiguration(ctx context.Context, principal AuthorizationResult, agentID string, req UpdateAgentConfigurationRequest) (AgentClientRecordResponse, error) {
	if err := ctx.Err(); err != nil {
		return AgentClientRecordResponse{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	agent, ok := s.agentForMutation(principal, agentID)
	if !ok {
		return AgentClientRecordResponse{}, ErrAIAgentNotFound
	}
	if agent.AssignedTaskCount > 0 {
		return AgentClientRecordResponse{}, ErrAIAgentAssigned
	}
	if strings.TrimSpace(req.Name) != "" {
		agent.Name = strings.TrimSpace(req.Name)
	}
	if req.Visibility != "" {
		if req.Visibility != AgentVisibilityPublic && req.Visibility != AgentVisibilityPrivate {
			return AgentClientRecordResponse{}, errors.New("visibility must be public or private")
		}
		agent.Visibility = req.Visibility
	}
	if strings.TrimSpace(req.RuntimeID) != "" {
		agent.RuntimeID = strings.TrimSpace(req.RuntimeID)
		if kind, ok := runtimeKindByID(s.devices, agent.RuntimeID); ok {
			agent.RuntimeKind = kind
		}
	}
	agent.Editability = editabilityForAssignedTasks(agent.AssignedTaskCount)
	s.agents[agent.AgentID] = agent
	agent.IsOwnedByViewer = agent.OwnerPrincipalID == principal.PrincipalID
	return AgentClientRecordResponse{SchemaVersion: SchemaVersion, Agent: agent}, nil
}

func (s *MockAIAgentClientStore) DeleteAIAgent(ctx context.Context, principal AuthorizationResult, agentID string) (DeleteAgentResponse, error) {
	if err := ctx.Err(); err != nil {
		return DeleteAgentResponse{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	agent, ok := s.agentForMutation(principal, agentID)
	if !ok {
		return DeleteAgentResponse{}, ErrAIAgentNotFound
	}
	queued, running := 0, 0
	switch agent.WorkStatus {
	case AgentWorkStatusQueued:
		queued = agent.AssignedTaskCount
	case AgentWorkStatusRunning, AgentWorkStatusWaitingForUser:
		running = agent.AssignedTaskCount
	}
	delete(s.agents, agent.AgentID)
	s.events = append(s.events, ClientStreamEvent{
		Seq:       int64(len(s.events) + 1),
		EventType: "agent_work_status_changed",
		Payload: AgentWorkStatusChangedEvent{
			EventType:       "agent_work_status_changed",
			SchemaVersion:   SchemaVersion,
			AgentID:         agent.AgentID,
			TaskID:          "task-mock-1",
			WorkStatus:      AgentWorkStatusOffline,
			AssignmentState: AgentAssignmentStateStopped,
			CommentKind:     AgentTaskCommentStoppedByAgentDeleted,
		},
	})
	return DeleteAgentResponse{
		SchemaVersion:            SchemaVersion,
		AgentID:                  agent.AgentID,
		QueuedTasksUnassigned:    queued,
		RunningTasksForceStopped: running,
	}, nil
}

func (s *MockAIAgentClientStore) AIAgentClientEvents(ctx context.Context, principal AuthorizationResult) ([]ClientStreamEvent, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	events := make([]ClientStreamEvent, 0, len(s.events))
	visible := s.visibleAgentIDs(principal)
	for _, event := range s.events {
		if agentID, ok := eventAgentID(event.Payload); ok {
			if _, isVisible := visible[agentID]; !isVisible {
				continue
			}
		}
		events = append(events, event)
	}
	return events, nil
}

var (
	ErrAIAgentNotFound = errors.New("ai agent not found")
	ErrAIAgentAssigned = errors.New("ai agent has assigned tasks")
)

func (s *MockAIAgentClientStore) visibleAgent(principal AuthorizationResult, agentID string) (AgentClientRecord, bool) {
	agentID = strings.TrimSpace(agentID)
	if agentID == "" {
		return AgentClientRecord{}, false
	}
	agent, ok := s.agents[agentID]
	if !ok || !aiAgentVisibleTo(principal, agent) {
		return AgentClientRecord{}, false
	}
	agent.IsOwnedByViewer = agent.OwnerPrincipalID == principal.PrincipalID
	return agent, true
}

func (s *MockAIAgentClientStore) visibleAgents(principal AuthorizationResult) []AgentClientRecord {
	agents := make([]AgentClientRecord, 0, len(s.agents))
	for _, agent := range s.agents {
		if !aiAgentVisibleTo(principal, agent) {
			continue
		}
		agent.IsOwnedByViewer = agent.OwnerPrincipalID == principal.PrincipalID
		agent.Editability = editabilityForAssignedTasks(agent.AssignedTaskCount)
		agents = append(agents, agent)
	}
	sort.SliceStable(agents, func(i, j int) bool {
		if agents[i].IsOwnedByViewer != agents[j].IsOwnedByViewer {
			return agents[i].IsOwnedByViewer
		}
		if agents[i].Name != agents[j].Name {
			return agents[i].Name < agents[j].Name
		}
		return agents[i].AgentID < agents[j].AgentID
	})
	return agents
}

func (s *MockAIAgentClientStore) visibleAgentIDs(principal AuthorizationResult) map[string]struct{} {
	out := map[string]struct{}{}
	for _, agent := range s.visibleAgents(principal) {
		out[agent.AgentID] = struct{}{}
	}
	return out
}

func (s *MockAIAgentClientStore) agentForMutation(principal AuthorizationResult, agentID string) (AgentClientRecord, bool) {
	agentID = strings.TrimSpace(agentID)
	if agentID == "" {
		return AgentClientRecord{}, false
	}
	agent, ok := s.agents[agentID]
	if !ok || !aiAgentMutableBy(principal, agent) {
		return AgentClientRecord{}, false
	}
	return agent, true
}

func (s *MockAIAgentClientStore) agentForTaskStop(principal AuthorizationResult, agentID string) (AgentClientRecord, bool) {
	if strings.TrimSpace(agentID) != "" {
		return s.visibleAgent(principal, agentID)
	}
	for _, agent := range s.visibleAgents(principal) {
		if agent.WorkStatus == AgentWorkStatusRunning || agent.WorkStatus == AgentWorkStatusWaitingForUser {
			return agent, true
		}
	}
	return AgentClientRecord{}, false
}

func (s *MockAIAgentClientStore) appendAgentTaskActionEvent(response AIAgentTaskActionResponse) {
	s.events = append(s.events, ClientStreamEvent{
		Seq:       int64(len(s.events) + 1),
		EventType: "agent_work_status_changed",
		Payload: AgentWorkStatusChangedEvent{
			EventType:       "agent_work_status_changed",
			SchemaVersion:   SchemaVersion,
			AgentID:         response.AgentID,
			TaskID:          response.TaskID,
			RunID:           response.RunID,
			WorkStatus:      response.WorkStatus,
			AssignmentState: response.AssignmentState,
			CommentKind:     response.CommentKind,
		},
	})
}

func aiAgentVisibleTo(principal AuthorizationResult, agent AgentClientRecord) bool {
	if aiAgentIsAdmin(principal) || agent.OwnerPrincipalID == principal.PrincipalID {
		return true
	}
	return agent.Visibility == AgentVisibilityPublic
}

func aiAgentMutableBy(principal AuthorizationResult, agent AgentClientRecord) bool {
	return aiAgentIsAdmin(principal) || agent.OwnerPrincipalID == principal.PrincipalID
}

func aiAgentIsAdmin(principal AuthorizationResult) bool {
	for _, role := range principal.Roles {
		if role == AgentCatalogRoleAdmin {
			return true
		}
	}
	return false
}

func normalizeClientKind(kind ClientKind) ClientKind {
	switch kind {
	case ClientKindDesktopWebview:
		return ClientKindDesktopWebview
	default:
		return ClientKindWeb
	}
}

func editabilityForAssignedTasks(count int) AgentEditability {
	if count > 0 {
		return AgentEditabilityBlockedAssignedTasks
	}
	return AgentEditabilityEditable
}

func filterDevicesForPrincipal(devices []DeviceRecord, principal AuthorizationResult) []DeviceRecord {
	if aiAgentIsAdmin(principal) {
		return copyDevices(devices)
	}
	var out []DeviceRecord
	for _, device := range devices {
		if device.OwnerPrincipalID == principal.PrincipalID {
			out = append(out, copyDevice(device))
		}
	}
	return out
}

func copyDevices(devices []DeviceRecord) []DeviceRecord {
	out := make([]DeviceRecord, 0, len(devices))
	for _, device := range devices {
		out = append(out, copyDevice(device))
	}
	return out
}

func copyDevice(device DeviceRecord) DeviceRecord {
	device.Runtimes = append([]RuntimeRecord(nil), device.Runtimes...)
	return device
}

func runtimeKindByID(devices []DeviceRecord, runtimeID string) (RuntimeKind, bool) {
	for _, device := range devices {
		for _, runtime := range device.Runtimes {
			if runtime.RuntimeID == runtimeID {
				return runtime.Kind, true
			}
		}
	}
	return "", false
}

func eventAgentID(payload any) (string, bool) {
	switch event := payload.(type) {
	case AgentEditabilityChangedEvent:
		return event.AgentID, true
	case AgentWorkStatusChangedEvent:
		return event.AgentID, true
	default:
		return "", false
	}
}
