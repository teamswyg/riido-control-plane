package riidoaiserver

import (
	"context"
	"errors"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"
)

type AIAgentClientStore interface {
	BootstrapAIAgentClient(ctx context.Context, principal AuthorizationResult, clientKind ClientKind) (ClientBootstrapResponse, error)
	ListAIAgentDevices(ctx context.Context, principal AuthorizationResult) (DeviceRuntimeListResponse, error)
	ListAIAgentTaskAssignableAgents(ctx context.Context, principal AuthorizationResult, taskID string) (AgentClientListResponse, error)
	SubmitAIAgentTaskComment(ctx context.Context, principal AuthorizationResult, taskID string, req SubmitAIAgentTaskCommentRequest) (AIAgentTaskActionResponse, error)
	StopAIAgentTask(ctx context.Context, principal AuthorizationResult, taskID string, req StopAIAgentTaskRequest) (AIAgentTaskActionResponse, error)
	CreateAIAgent(ctx context.Context, principal AuthorizationResult, req CreateAgentConfigurationRequest) (AgentClientRecordResponse, error)
	GetAIAgentEditability(ctx context.Context, principal AuthorizationResult, agentID string) (AgentEditabilityResponse, error)
	UpdateAIAgentConfiguration(ctx context.Context, principal AuthorizationResult, agentID string, req UpdateAgentConfigurationRequest) (AgentClientRecordResponse, error)
	DeleteAIAgent(ctx context.Context, principal AuthorizationResult, agentID string) (DeleteAgentResponse, error)
	AIAgentClientEvents(ctx context.Context, principal AuthorizationResult) ([]ClientStreamEvent, error)
}

type AIAgentClientEventSubscriber interface {
	SubscribeAIAgentClientEvents(ctx context.Context, principal AuthorizationResult) ([]ClientStreamEvent, <-chan ClientStreamEvent, func(), error)
}

type AIAgentThreadProgressRecorder interface {
	RecordAIAgentThreadProgress(ctx context.Context, agentID string, req AgentThreadProgressBatchRequest) (AgentThreadProgressBatchResponse, error)
}

type MockAIAgentClientStore struct {
	mu               sync.Mutex
	workspaceID      string
	devices          []DeviceRecord
	agents           map[string]AgentClientRecord
	events           []ClientStreamEvent
	subscribers      map[int]aiAgentClientSubscriber
	nextSubscriberID int
}

type aiAgentClientSubscriber struct {
	principal AuthorizationResult
	events    chan ClientStreamEvent
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
			AgentID:             "agent-owned-codex",
			OwnerPrincipalID:    "user-1",
			Name:                "Codex 리뷰어",
			ProfileThumbnailURL: "https://cdn.riido.io/mock/ai-agents/codex-reviewer.png",
			Description:         "코드 변경 위험을 먼저 보는 리뷰 에이전트",
			Instruction:         "코드 변경의 위험과 검증 근거를 우선 확인합니다.",
			Visibility:          AgentVisibilityPrivate,
			RuntimeID:           "runtime-codex-mock",
			RuntimeKind:         RuntimeKindCodex,
			WorkStatus:          AgentWorkStatusRunning,
			Editability:         AgentEditabilityBlockedAssignedTasks,
			AssignedTaskCount:   1,
			UpdatedAt:           now.Add(-6 * time.Hour),
		},
		"agent-owned-claude": {
			AgentID:             "agent-owned-claude",
			OwnerPrincipalID:    "user-1",
			Name:                "Claude 설계 보조",
			ProfileThumbnailURL: "https://cdn.riido.io/mock/ai-agents/claude-designer.png",
			Description:         "기획 의도를 구현 범위로 정리하는 설계 에이전트",
			Instruction:         "기획 의도와 도메인 정책을 먼저 정리한 뒤 구현 범위를 제안합니다.",
			Visibility:          AgentVisibilityPrivate,
			RuntimeID:           "runtime-claude-code-mock",
			RuntimeKind:         RuntimeKindClaudeCode,
			WorkStatus:          AgentWorkStatusOffline,
			Editability:         AgentEditabilityEditable,
			AssignedTaskCount:   0,
			UpdatedAt:           now.Add(-5 * time.Hour),
		},
		"agent-public-openclaw": {
			AgentID:             "agent-public-openclaw",
			OwnerPrincipalID:    "user-2",
			Name:                "OpenClaw 공개 에이전트",
			ProfileThumbnailURL: "https://cdn.riido.io/mock/ai-agents/openclaw-public.png",
			Description:         "공개 워크스페이스 반복 작업 에이전트",
			Instruction:         "공개 워크스페이스에서 반복 가능한 보조 작업을 수행합니다.",
			Visibility:          AgentVisibilityPublic,
			RuntimeID:           "runtime-openclaw-remote",
			RuntimeKind:         RuntimeKindOpenClaw,
			WorkStatus:          AgentWorkStatusIdle,
			Editability:         AgentEditabilityEditable,
			AssignedTaskCount:   0,
			UpdatedAt:           now.Add(-4 * time.Hour),
		},
		"agent-private-cursor": {
			AgentID:             "agent-private-cursor",
			OwnerPrincipalID:    "user-2",
			Name:                "Cursor 비공개 에이전트",
			ProfileThumbnailURL: "https://cdn.riido.io/mock/ai-agents/cursor-private.png",
			Description:         "소유자 전용 Cursor 코드 탐색 에이전트",
			Instruction:         "소유자 전용 Cursor 기반 코드 탐색을 수행합니다.",
			Visibility:          AgentVisibilityPrivate,
			RuntimeID:           "runtime-cursor-mock",
			RuntimeKind:         RuntimeKindCursor,
			WorkStatus:          AgentWorkStatusIdle,
			Editability:         AgentEditabilityEditable,
			AssignedTaskCount:   0,
			UpdatedAt:           now.Add(-3 * time.Hour),
		},
	}
	return &MockAIAgentClientStore{
		workspaceID: "workspace-mock-riid",
		devices:     []DeviceRecord{device},
		agents:      agents,
		subscribers: map[int]aiAgentClientSubscriber{},
		events: []ClientStreamEvent{
			{
				Seq:       1,
				EventType: AgentClientEventDeviceRuntimeSnapshot,
				Payload: DeviceRuntimeSnapshotEvent{
					EventType:     AgentClientEventDeviceRuntimeSnapshot,
					SchemaVersion: SchemaVersion,
					Device:        device,
				},
			},
			{
				Seq:       2,
				EventType: AgentClientEventWorkStatusChanged,
				Payload: AgentWorkStatusChangedEvent{
					EventType:       AgentClientEventWorkStatusChanged,
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

func (s *MockAIAgentClientStore) CreateAIAgent(ctx context.Context, principal AuthorizationResult, req CreateAgentConfigurationRequest) (AgentClientRecordResponse, error) {
	if err := ctx.Err(); err != nil {
		return AgentClientRecordResponse{}, err
	}
	name := strings.TrimSpace(req.Name)
	runtimeID := strings.TrimSpace(req.RuntimeID)
	if name == "" {
		return AgentClientRecordResponse{}, errors.New("name is required")
	}
	if runtimeID == "" {
		return AgentClientRecordResponse{}, errors.New("runtime_id is required")
	}
	if req.Visibility != AgentVisibilityPublic && req.Visibility != AgentVisibilityPrivate {
		return AgentClientRecordResponse{}, errors.New("visibility must be public or private")
	}
	thumbnailURL := ""
	if req.ProfileThumbnailURL != nil {
		var err error
		thumbnailURL, err = normalizeAgentProfileThumbnailURL(*req.ProfileThumbnailURL)
		if err != nil {
			return AgentClientRecordResponse{}, err
		}
	}
	description := ""
	if req.Description != nil {
		if err := validateAgentDescription(*req.Description); err != nil {
			return AgentClientRecordResponse{}, err
		}
		description = *req.Description
	}
	instruction := ""
	if req.Instruction != nil {
		if err := validateAgentInstruction(*req.Instruction); err != nil {
			return AgentClientRecordResponse{}, err
		}
		instruction = *req.Instruction
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	runtimeKind, ok := runtimeKindByIDForPrincipal(s.devices, principal, runtimeID)
	if !ok {
		return AgentClientRecordResponse{}, errors.New("runtime_id is not available")
	}
	now := time.Now().UTC()
	agentID := uniqueAIAgentIDLocked(s.agents, "agent-"+principal.PrincipalID+"-"+runtimeID)
	agent := AgentClientRecord{
		AgentID:             agentID,
		OwnerPrincipalID:    principal.PrincipalID,
		IsOwnedByViewer:     true,
		Name:                name,
		ProfileThumbnailURL: thumbnailURL,
		Description:         description,
		Instruction:         instruction,
		Visibility:          req.Visibility,
		RuntimeID:           runtimeID,
		RuntimeKind:         runtimeKind,
		WorkStatus:          AgentWorkStatusIdle,
		Editability:         AgentEditabilityEditable,
		AssignedTaskCount:   0,
		UpdatedAt:           now,
	}
	s.agents[agent.AgentID] = agent
	markRuntimeHasAssignedAgentLocked(s.devices, runtimeID, true)
	return AgentClientRecordResponse{SchemaVersion: SchemaVersion, Agent: agent}, nil
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
	if req.ProfileThumbnailURL != nil {
		thumbnailURL, err := normalizeAgentProfileThumbnailURL(*req.ProfileThumbnailURL)
		if err != nil {
			return AgentClientRecordResponse{}, err
		}
		agent.ProfileThumbnailURL = thumbnailURL
	}
	if req.Description != nil {
		if err := validateAgentDescription(*req.Description); err != nil {
			return AgentClientRecordResponse{}, err
		}
		agent.Description = *req.Description
	}
	if req.Instruction != nil {
		if err := validateAgentInstruction(*req.Instruction); err != nil {
			return AgentClientRecordResponse{}, err
		}
		agent.Instruction = *req.Instruction
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
	agent.UpdatedAt = time.Now().UTC()
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
	s.appendClientEventLocked(AgentClientEventWorkStatusChanged, AgentWorkStatusChangedEvent{
		EventType:       AgentClientEventWorkStatusChanged,
		SchemaVersion:   SchemaVersion,
		AgentID:         agent.AgentID,
		TaskID:          "task-mock-1",
		WorkStatus:      AgentWorkStatusOffline,
		AssignmentState: AgentAssignmentStateStopped,
		CommentKind:     AgentTaskCommentStoppedByAgentDeleted,
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
	return s.clientEventsForPrincipalLocked(principal), nil
}

func (s *MockAIAgentClientStore) SubscribeAIAgentClientEvents(ctx context.Context, principal AuthorizationResult) ([]ClientStreamEvent, <-chan ClientStreamEvent, func(), error) {
	if err := ctx.Err(); err != nil {
		return nil, nil, nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	history := s.clientEventsForPrincipalLocked(principal)
	s.nextSubscriberID++
	id := s.nextSubscriberID
	events := make(chan ClientStreamEvent, 32)
	s.subscribers[id] = aiAgentClientSubscriber{principal: principal, events: events}
	cancel := func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		if subscriber, ok := s.subscribers[id]; ok {
			delete(s.subscribers, id)
			close(subscriber.events)
		}
	}
	return history, events, cancel, nil
}

func (s *MockAIAgentClientStore) RecordAIAgentThreadProgress(ctx context.Context, agentID string, req AgentThreadProgressBatchRequest) (AgentThreadProgressBatchResponse, error) {
	if err := ctx.Err(); err != nil {
		return AgentThreadProgressBatchResponse{}, err
	}
	agentID = strings.TrimSpace(agentID)
	req.TaskID = strings.TrimSpace(req.TaskID)
	req.AssignmentID = strings.TrimSpace(req.AssignmentID)
	req.RunID = strings.TrimSpace(req.RunID)
	if agentID == "" {
		return AgentThreadProgressBatchResponse{}, errors.New("agent_id is required")
	}
	if req.TaskID == "" {
		return AgentThreadProgressBatchResponse{}, errors.New("task_id is required")
	}
	if req.AssignmentID == "" {
		return AgentThreadProgressBatchResponse{}, errors.New("assignment_id is required")
	}
	lines := normalizeProgressLines(req.Lines)
	if len(lines) == 0 {
		return AgentThreadProgressBatchResponse{}, errors.New("lines are required")
	}
	if req.RunID == "" {
		req.RunID = "run-" + req.AssignmentID
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	agent, ok := s.agents[agentID]
	if !ok {
		return AgentThreadProgressBatchResponse{}, ErrAIAgentNotFound
	}
	agent.WorkStatus = AgentWorkStatusRunning
	agent.Editability = AgentEditabilityBlockedAssignedTasks
	if agent.AssignedTaskCount == 0 {
		agent.AssignedTaskCount = 1
	}
	s.agents[agentID] = agent
	event := AgentThreadProgressEvent{
		EventType:       AgentClientEventThreadProgress,
		SchemaVersion:   SchemaVersion,
		AgentID:         agentID,
		TaskID:          req.TaskID,
		RunID:           req.RunID,
		WorkStatus:      AgentWorkStatusRunning,
		AssignmentState: AgentAssignmentStateRunning,
		CommentKind:     AgentTaskCommentRuntimeProgress,
		BatchStartedAt:  req.BatchStartedAt,
		BatchEndedAt:    req.BatchEndedAt,
		Lines:           lines,
	}
	s.appendClientEventLocked(event.EventType, event)
	return AgentThreadProgressBatchResponse{
		SchemaVersion: SchemaVersion,
		AcceptedLines: len(lines),
		Event:         event,
	}, nil
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
	s.appendClientEventLocked(AgentClientEventWorkStatusChanged, AgentWorkStatusChangedEvent{
		EventType:       AgentClientEventWorkStatusChanged,
		SchemaVersion:   SchemaVersion,
		AgentID:         response.AgentID,
		TaskID:          response.TaskID,
		RunID:           response.RunID,
		WorkStatus:      response.WorkStatus,
		AssignmentState: response.AssignmentState,
		CommentKind:     response.CommentKind,
	})
}

func (s *MockAIAgentClientStore) clientEventsForPrincipalLocked(principal AuthorizationResult) []ClientStreamEvent {
	events := make([]ClientStreamEvent, 0, len(s.events))
	for _, event := range s.events {
		if !clientEventVisibleToLocked(s, principal, event) {
			continue
		}
		events = append(events, event)
	}
	return events
}

func (s *MockAIAgentClientStore) appendClientEventLocked(eventType string, payload any) ClientStreamEvent {
	event := ClientStreamEvent{
		Seq:       int64(len(s.events) + 1),
		EventType: eventType,
		Payload:   payload,
	}
	s.events = append(s.events, event)
	for _, subscriber := range s.subscribers {
		if !clientEventVisibleToLocked(s, subscriber.principal, event) {
			continue
		}
		select {
		case subscriber.events <- event:
		default:
		}
	}
	return event
}

func clientEventVisibleToLocked(s *MockAIAgentClientStore, principal AuthorizationResult, event ClientStreamEvent) bool {
	agentID, ok := eventAgentID(event.Payload)
	if !ok {
		return true
	}
	agent, exists := s.agents[agentID]
	if !exists {
		return aiAgentIsAdmin(principal)
	}
	return aiAgentVisibleTo(principal, agent)
}

func normalizeProgressLines(lines []AgentThreadProgressLine) []AgentThreadProgressLine {
	out := make([]AgentThreadProgressLine, 0, len(lines))
	for _, line := range lines {
		line.Message = strings.TrimSpace(line.Message)
		if line.Message == "" {
			continue
		}
		if line.Seq <= 0 {
			line.Seq = len(out) + 1
		}
		out = append(out, line)
	}
	return out
}

func normalizeAgentProfileThumbnailURL(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", nil
	}
	parsed, err := url.Parse(trimmed)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
		return "", errors.New("profile_thumbnail_url must be an https URL")
	}
	return trimmed, nil
}

func validateAgentInstruction(value string) error {
	if utf8.RuneCountInString(value) > AgentInstructionMaxCharacters {
		return errors.New("instruction must be 1000 characters or fewer")
	}
	return nil
}

func validateAgentDescription(value string) error {
	if utf8.RuneCountInString(value) > AgentDescriptionMaxCharacters {
		return errors.New("description must be 160 characters or fewer")
	}
	return nil
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

func runtimeKindByIDForPrincipal(devices []DeviceRecord, principal AuthorizationResult, runtimeID string) (RuntimeKind, bool) {
	for _, device := range filterDevicesForPrincipal(devices, principal) {
		for _, runtime := range device.Runtimes {
			if runtime.RuntimeID == runtimeID {
				return runtime.Kind, true
			}
		}
	}
	return "", false
}

func markRuntimeHasAssignedAgentLocked(devices []DeviceRecord, runtimeID string, value bool) {
	for deviceIndex := range devices {
		for runtimeIndex := range devices[deviceIndex].Runtimes {
			if devices[deviceIndex].Runtimes[runtimeIndex].RuntimeID == runtimeID {
				devices[deviceIndex].Runtimes[runtimeIndex].HasAssignedAgent = value
				return
			}
		}
	}
}

func uniqueAIAgentIDLocked(agents map[string]AgentClientRecord, seed string) string {
	base := slugAIAgentIDComponent(seed)
	if base == "" {
		base = "agent"
	}
	if _, exists := agents[base]; !exists {
		return base
	}
	for i := 2; ; i++ {
		candidate := base + "-" + strconv.Itoa(i)
		if _, exists := agents[candidate]; !exists {
			return candidate
		}
	}
}

func slugAIAgentIDComponent(value string) string {
	var b strings.Builder
	previousDash := false
	for _, r := range strings.ToLower(strings.TrimSpace(value)) {
		isAllowed := unicode.IsLetter(r) || unicode.IsDigit(r)
		if isAllowed {
			b.WriteRune(r)
			previousDash = false
			continue
		}
		if !previousDash {
			b.WriteByte('-')
			previousDash = true
		}
	}
	return strings.Trim(b.String(), "-")
}

func eventAgentID(payload any) (string, bool) {
	switch event := payload.(type) {
	case AgentEditabilityChangedEvent:
		return event.AgentID, true
	case AgentWorkStatusChangedEvent:
		return event.AgentID, true
	case AgentThreadProgressEvent:
		return event.AgentID, true
	default:
		return "", false
	}
}
