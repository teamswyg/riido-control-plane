package riidoaiserver

import (
	"context"
	"errors"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type AIAgentClientStore interface {
	BootstrapAIAgentClient(ctx context.Context, principal AuthorizationResult, clientKind ClientKind) (ClientBootstrapResponse, error)
	ListAIAgentDevices(ctx context.Context, principal AuthorizationResult) (DeviceRuntimeListResponse, error)
	ListAIAgentTaskAssignableAgents(ctx context.Context, principal AuthorizationResult, taskID string) (AgentClientListResponse, error)
	GetAIAgentTaskThreads(ctx context.Context, principal AuthorizationResult, taskID string) (AIAgentTaskThreadCollectionResponse, error)
	SubmitAIAgentTaskComment(ctx context.Context, principal AuthorizationResult, taskID string, req SubmitAIAgentTaskCommentRequest) (AIAgentTaskActionResponse, error)
	StopAIAgentTask(ctx context.Context, principal AuthorizationResult, taskID string, req StopAIAgentTaskRequest) (AIAgentTaskActionResponse, error)
	GetAIAgentEditability(ctx context.Context, principal AuthorizationResult, agentID string) (AgentEditabilityResponse, error)
	UpdateAIAgentConfiguration(ctx context.Context, principal AuthorizationResult, agentID string, req UpdateAgentConfigurationRequest) (AgentClientRecordResponse, error)
	DeleteAIAgent(ctx context.Context, principal AuthorizationResult, agentID string) (DeleteAgentResponse, error)
	AIAgentClientEvents(ctx context.Context, principal AuthorizationResult) ([]ClientStreamEvent, error)
}

type AIAgentClientEventSubscriber interface {
	SubscribeAIAgentClientEvents(ctx context.Context, principal AuthorizationResult) ([]ClientStreamEvent, <-chan ClientStreamEvent, func(), error)
}

type AIAgentTaskThreadEventSubscriber interface {
	SubscribeAIAgentTaskThreadEvents(ctx context.Context, principal AuthorizationResult, taskID, threadID string) ([]AgentThreadProgressEvent, <-chan AgentThreadProgressEvent, func(), bool, error)
}

type AIAgentThreadProgressRecorder interface {
	RecordAIAgentThreadProgress(ctx context.Context, agentID string, req AgentThreadProgressBatchRequest) (AgentThreadProgressBatchResponse, error)
}

type MockAIAgentClientStore struct {
	mu                     sync.Mutex
	workspaceID            string
	devices                []DeviceRecord
	agents                 map[string]AgentClientRecord
	taskThreads            map[string][]AIAgentTaskThreadRecord
	threadEvents           map[string][]AgentThreadProgressEvent
	events                 []ClientStreamEvent
	subscribers            map[int]aiAgentClientSubscriber
	threadSubscribers      map[int]aiAgentTaskThreadSubscriber
	nextSubscriberID       int
	nextThreadSubscriberID int
}

type aiAgentClientSubscriber struct {
	principal AuthorizationResult
	events    chan ClientStreamEvent
}

type aiAgentTaskThreadSubscriber struct {
	principal AuthorizationResult
	taskID    string
	threadID  string
	events    chan AgentThreadProgressEvent
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
		workspaceID:       "workspace-mock-riid",
		devices:           []DeviceRecord{device},
		agents:            agents,
		taskThreads:       seedAIAgentTaskThreads(now),
		threadEvents:      seedAIAgentTaskThreadEvents(now),
		subscribers:       map[int]aiAgentClientSubscriber{},
		threadSubscribers: map[int]aiAgentTaskThreadSubscriber{},
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

func seedAIAgentTaskThreads(now time.Time) map[string][]AIAgentTaskThreadRecord {
	return map[string][]AIAgentTaskThreadRecord{
		"task-mock-1": {
			{
				ThreadID:        "thread-mock-completed",
				TaskID:          "task-mock-1",
				AgentID:         "agent-owned-codex",
				AssignmentID:    "asn-mock-completed",
				RunID:           "run-mock-completed",
				StreamState:     AIAgentTaskThreadStreamStateTerminal,
				WorkStatus:      AgentWorkStatusCompleted,
				AssignmentState: AgentAssignmentStateCompleted,
				StartedAt:       now.Add(-20 * time.Minute),
				EndedAt:         now.Add(-15 * time.Minute),
				Entries: []AIAgentTaskThreadEntry{
					{
						EntryID:     "entry-mock-completed-1",
						CommentKind: AgentTaskCommentTaskCompleted,
						Message:     "이전 AI Agent 작업이 완료되었습니다.",
						CreatedAt:   now.Add(-15 * time.Minute),
					},
				},
			},
			{
				ThreadID:        "thread-mock-active",
				TaskID:          "task-mock-1",
				AgentID:         "agent-owned-codex",
				AssignmentID:    "asn-mock-active",
				RunID:           "run-mock-active",
				StreamState:     AIAgentTaskThreadStreamStateActive,
				WorkStatus:      AgentWorkStatusRunning,
				AssignmentState: AgentAssignmentStateRunning,
				StartedAt:       now.Add(-2 * time.Minute),
				Entries: []AIAgentTaskThreadEntry{
					{
						EntryID:     "entry-mock-active-1",
						CommentKind: AgentTaskCommentRuntimeProgress,
						Message:     "팀 프로젝트 수집 중",
						CreatedAt:   now.Add(-90 * time.Second),
						Lines: []AgentThreadProgressLine{
							{Seq: 1, Message: "팀 프로젝트 수집 중", ObservedAt: now.Add(-90 * time.Second)},
						},
					},
				},
			},
		},
		"task-mock-completed-only": {
			{
				ThreadID:        "thread-mock-completed-only",
				TaskID:          "task-mock-completed-only",
				AgentID:         "agent-owned-codex",
				AssignmentID:    "asn-mock-completed-only",
				RunID:           "run-mock-completed-only",
				StreamState:     AIAgentTaskThreadStreamStateTerminal,
				WorkStatus:      AgentWorkStatusCompleted,
				AssignmentState: AgentAssignmentStateCompleted,
				StartedAt:       now.Add(-10 * time.Minute),
				EndedAt:         now.Add(-5 * time.Minute),
				Entries: []AIAgentTaskThreadEntry{
					{
						EntryID:     "entry-mock-completed-only-1",
						CommentKind: AgentTaskCommentTaskCompleted,
						Message:     "AI Agent 작업이 완료되었습니다.",
						CreatedAt:   now.Add(-5 * time.Minute),
					},
				},
			},
		},
	}
}

func seedAIAgentTaskThreadEvents(now time.Time) map[string][]AgentThreadProgressEvent {
	return map[string][]AgentThreadProgressEvent{
		"thread-mock-active": {
			{
				EventType:       AgentClientEventThreadProgress,
				SchemaVersion:   SchemaVersion,
				AgentID:         "agent-owned-codex",
				TaskID:          "task-mock-1",
				ThreadID:        "thread-mock-active",
				RunID:           "run-mock-active",
				WorkStatus:      AgentWorkStatusRunning,
				AssignmentState: AgentAssignmentStateRunning,
				CommentKind:     AgentTaskCommentRuntimeProgress,
				BatchStartedAt:  now.Add(-90 * time.Second),
				BatchEndedAt:    now.Add(-90 * time.Second),
				Lines: []AgentThreadProgressLine{
					{Seq: 1, Message: "팀 프로젝트 수집 중", ObservedAt: now.Add(-90 * time.Second)},
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

func (s *MockAIAgentClientStore) GetAIAgentTaskThreads(ctx context.Context, principal AuthorizationResult, taskID string) (AIAgentTaskThreadCollectionResponse, error) {
	if err := ctx.Err(); err != nil {
		return AIAgentTaskThreadCollectionResponse{}, err
	}
	taskID = strings.TrimSpace(taskID)
	if taskID == "" {
		return AIAgentTaskThreadCollectionResponse{}, errors.New("task_id is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	threads, activeThreadID, err := s.visibleTaskThreadsLocked(principal, taskID)
	if err != nil {
		return AIAgentTaskThreadCollectionResponse{}, err
	}
	response := AIAgentTaskThreadCollectionResponse{
		SchemaVersion: SchemaVersion,
		TaskID:        taskID,
		Threads:       threads,
		Links:         AIAgentTaskThreadLinks{},
	}
	if activeThreadID != "" {
		response.ActiveThreadID = activeThreadID
		response.Links.ActiveStream = &AIAgentTaskThreadStreamLink{
			Rel:         "active_stream",
			Href:        "/v1/client/ai-agent/tasks/" + taskID + "/threads/" + activeThreadID + "/events",
			Method:      "GET",
			ContentType: "text/event-stream",
			EventType:   AgentClientEventThreadProgress,
			ThreadID:    activeThreadID,
		}
	}
	return response, nil
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
	s.upsertTaskThreadActionLocked(response)
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
	s.stopActiveTaskThreadLocked(response)
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

func (s *MockAIAgentClientStore) SubscribeAIAgentTaskThreadEvents(ctx context.Context, principal AuthorizationResult, taskID, threadID string) ([]AgentThreadProgressEvent, <-chan AgentThreadProgressEvent, func(), bool, error) {
	if err := ctx.Err(); err != nil {
		return nil, nil, nil, false, err
	}
	taskID = strings.TrimSpace(taskID)
	threadID = strings.TrimSpace(threadID)
	if taskID == "" {
		return nil, nil, nil, false, errors.New("task_id is required")
	}
	if threadID == "" {
		return nil, nil, nil, false, errors.New("thread_id is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	thread, ok := s.visibleTaskThreadLocked(principal, taskID, threadID)
	if !ok {
		return nil, nil, nil, false, ErrAIAgentThreadNotFound
	}
	history := s.threadProgressEventsForPrincipalLocked(principal, taskID, threadID)
	liveActive := thread.StreamState == AIAgentTaskThreadStreamStateActive
	if !liveActive {
		return history, nil, func() {}, false, nil
	}
	s.nextThreadSubscriberID++
	id := s.nextThreadSubscriberID
	events := make(chan AgentThreadProgressEvent, 32)
	s.threadSubscribers[id] = aiAgentTaskThreadSubscriber{principal: principal, taskID: taskID, threadID: threadID, events: events}
	cancel := func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		if subscriber, ok := s.threadSubscribers[id]; ok {
			delete(s.threadSubscribers, id)
			close(subscriber.events)
		}
	}
	return history, events, cancel, true, nil
}

func (s *MockAIAgentClientStore) RecordAIAgentThreadProgress(ctx context.Context, agentID string, req AgentThreadProgressBatchRequest) (AgentThreadProgressBatchResponse, error) {
	if err := ctx.Err(); err != nil {
		return AgentThreadProgressBatchResponse{}, err
	}
	agentID = strings.TrimSpace(agentID)
	req.TaskID = strings.TrimSpace(req.TaskID)
	req.AssignmentID = strings.TrimSpace(req.AssignmentID)
	req.ThreadID = strings.TrimSpace(req.ThreadID)
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
	threadID := taskThreadIDForProgress(req)
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
	entry := taskThreadEntryFromProgress(threadID, req, lines)
	s.upsertTaskThreadProgressLocked(agentID, threadID, req, entry)
	event := AgentThreadProgressEvent{
		EventType:       AgentClientEventThreadProgress,
		SchemaVersion:   SchemaVersion,
		AgentID:         agentID,
		TaskID:          req.TaskID,
		ThreadID:        threadID,
		RunID:           req.RunID,
		WorkStatus:      AgentWorkStatusRunning,
		AssignmentState: AgentAssignmentStateRunning,
		CommentKind:     AgentTaskCommentRuntimeProgress,
		BatchStartedAt:  req.BatchStartedAt,
		BatchEndedAt:    req.BatchEndedAt,
		Lines:           lines,
	}
	s.threadEvents[threadID] = append(s.threadEvents[threadID], copyAgentThreadProgressEvent(event))
	s.appendClientEventLocked(event.EventType, event)
	s.appendTaskThreadEventLocked(event)
	return AgentThreadProgressBatchResponse{
		SchemaVersion: SchemaVersion,
		AcceptedLines: len(lines),
		Event:         event,
	}, nil
}

var (
	ErrAIAgentNotFound       = errors.New("ai agent not found")
	ErrAIAgentThreadNotFound = errors.New("ai agent task thread not found")
	ErrAIAgentAssigned       = errors.New("ai agent has assigned tasks")
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

func (s *MockAIAgentClientStore) visibleTaskThreadsLocked(principal AuthorizationResult, taskID string) ([]AIAgentTaskThreadRecord, string, error) {
	source := s.taskThreads[taskID]
	threads := make([]AIAgentTaskThreadRecord, 0, len(source))
	activeThreadID := ""
	for _, thread := range source {
		agent, ok := s.agents[thread.AgentID]
		if !ok || !aiAgentVisibleTo(principal, agent) {
			continue
		}
		if thread.StreamState == AIAgentTaskThreadStreamStateActive {
			if activeThreadID != "" {
				return nil, "", errors.New("task has multiple active AI Agent threads")
			}
			activeThreadID = thread.ThreadID
		}
		threads = append(threads, copyAIAgentTaskThread(thread))
	}
	sort.SliceStable(threads, func(i, j int) bool {
		if !threads[i].StartedAt.Equal(threads[j].StartedAt) {
			return threads[i].StartedAt.Before(threads[j].StartedAt)
		}
		return threads[i].ThreadID < threads[j].ThreadID
	})
	return threads, activeThreadID, nil
}

func (s *MockAIAgentClientStore) visibleTaskThreadLocked(principal AuthorizationResult, taskID, threadID string) (AIAgentTaskThreadRecord, bool) {
	for _, thread := range s.taskThreads[taskID] {
		if thread.ThreadID != threadID {
			continue
		}
		agent, ok := s.agents[thread.AgentID]
		if !ok || !aiAgentVisibleTo(principal, agent) {
			return AIAgentTaskThreadRecord{}, false
		}
		return copyAIAgentTaskThread(thread), true
	}
	return AIAgentTaskThreadRecord{}, false
}

func (s *MockAIAgentClientStore) threadProgressEventsForPrincipalLocked(principal AuthorizationResult, taskID, threadID string) []AgentThreadProgressEvent {
	source := s.threadEvents[threadID]
	events := make([]AgentThreadProgressEvent, 0, len(source))
	for _, event := range source {
		if event.TaskID != taskID || event.ThreadID != threadID {
			continue
		}
		agent, ok := s.agents[event.AgentID]
		if !ok || !aiAgentVisibleTo(principal, agent) {
			continue
		}
		events = append(events, copyAgentThreadProgressEvent(event))
	}
	return events
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

func (s *MockAIAgentClientStore) appendTaskThreadEventLocked(event AgentThreadProgressEvent) {
	for _, subscriber := range s.threadSubscribers {
		if subscriber.taskID != event.TaskID || subscriber.threadID != event.ThreadID {
			continue
		}
		agent, ok := s.agents[event.AgentID]
		if !ok || !aiAgentVisibleTo(subscriber.principal, agent) {
			continue
		}
		select {
		case subscriber.events <- copyAgentThreadProgressEvent(event):
		default:
		}
	}
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

func taskThreadIDForProgress(req AgentThreadProgressBatchRequest) string {
	if strings.TrimSpace(req.ThreadID) != "" {
		return strings.TrimSpace(req.ThreadID)
	}
	return strings.TrimSpace(req.AssignmentID)
}

func taskThreadEntryFromProgress(threadID string, req AgentThreadProgressBatchRequest, lines []AgentThreadProgressLine) AIAgentTaskThreadEntry {
	createdAt := req.BatchEndedAt
	if createdAt.IsZero() {
		createdAt = req.BatchStartedAt
	}
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}
	message := lines[len(lines)-1].Message
	return AIAgentTaskThreadEntry{
		EntryID:     threadID + "-entry-" + strconv.Itoa(lines[0].Seq),
		CommentKind: AgentTaskCommentRuntimeProgress,
		Message:     message,
		CreatedAt:   createdAt,
		Lines:       append([]AgentThreadProgressLine(nil), lines...),
	}
}

func (s *MockAIAgentClientStore) upsertTaskThreadProgressLocked(agentID, threadID string, req AgentThreadProgressBatchRequest, entry AIAgentTaskThreadEntry) {
	threads := s.taskThreads[req.TaskID]
	for i := range threads {
		if threads[i].ThreadID == threadID {
			threads[i].AgentID = agentID
			threads[i].AssignmentID = req.AssignmentID
			threads[i].RunID = req.RunID
			threads[i].StreamState = AIAgentTaskThreadStreamStateActive
			threads[i].WorkStatus = AgentWorkStatusRunning
			threads[i].AssignmentState = AgentAssignmentStateRunning
			threads[i].Entries = append(threads[i].Entries, entry)
			s.taskThreads[req.TaskID] = threads
			return
		}
		if threads[i].StreamState == AIAgentTaskThreadStreamStateActive {
			threads[i].StreamState = AIAgentTaskThreadStreamStateTerminal
			threads[i].EndedAt = entry.CreatedAt
		}
	}
	threads = append(threads, AIAgentTaskThreadRecord{
		ThreadID:        threadID,
		TaskID:          req.TaskID,
		AgentID:         agentID,
		AssignmentID:    req.AssignmentID,
		RunID:           req.RunID,
		StreamState:     AIAgentTaskThreadStreamStateActive,
		WorkStatus:      AgentWorkStatusRunning,
		AssignmentState: AgentAssignmentStateRunning,
		StartedAt:       entry.CreatedAt,
		Entries:         []AIAgentTaskThreadEntry{entry},
	})
	s.taskThreads[req.TaskID] = threads
}

func (s *MockAIAgentClientStore) upsertTaskThreadActionLocked(response AIAgentTaskActionResponse) {
	threadID := response.RunID
	now := time.Now().UTC()
	entry := AIAgentTaskThreadEntry{
		EntryID:     threadID + "-entry-action",
		CommentKind: response.CommentKind,
		Message:     response.Message,
		CreatedAt:   now,
	}
	streamState := AIAgentTaskThreadStreamStateCold
	if response.WorkStatus == AgentWorkStatusRunning || response.WorkStatus == AgentWorkStatusWaitingForUser {
		streamState = AIAgentTaskThreadStreamStateActive
	}
	threads := s.taskThreads[response.TaskID]
	if streamState == AIAgentTaskThreadStreamStateActive {
		for i := range threads {
			if threads[i].StreamState == AIAgentTaskThreadStreamStateActive && threads[i].ThreadID != threadID {
				threads[i].StreamState = AIAgentTaskThreadStreamStateTerminal
				threads[i].EndedAt = now
			}
		}
	}
	for i := range threads {
		if threads[i].ThreadID == threadID {
			threads[i].WorkStatus = response.WorkStatus
			threads[i].AssignmentState = response.AssignmentState
			threads[i].StreamState = streamState
			threads[i].Entries = append(threads[i].Entries, entry)
			s.taskThreads[response.TaskID] = threads
			return
		}
	}
	threads = append(threads, AIAgentTaskThreadRecord{
		ThreadID:        threadID,
		TaskID:          response.TaskID,
		AgentID:         response.AgentID,
		AssignmentID:    response.RunID,
		RunID:           response.RunID,
		StreamState:     streamState,
		WorkStatus:      response.WorkStatus,
		AssignmentState: response.AssignmentState,
		StartedAt:       now,
		Entries:         []AIAgentTaskThreadEntry{entry},
	})
	s.taskThreads[response.TaskID] = threads
}

func (s *MockAIAgentClientStore) stopActiveTaskThreadLocked(response AIAgentTaskActionResponse) {
	now := time.Now().UTC()
	threads := s.taskThreads[response.TaskID]
	for i := range threads {
		if threads[i].AgentID != response.AgentID || threads[i].StreamState != AIAgentTaskThreadStreamStateActive {
			continue
		}
		threads[i].StreamState = AIAgentTaskThreadStreamStateTerminal
		threads[i].WorkStatus = response.WorkStatus
		threads[i].AssignmentState = response.AssignmentState
		threads[i].EndedAt = now
		threads[i].Entries = append(threads[i].Entries, AIAgentTaskThreadEntry{
			EntryID:     threads[i].ThreadID + "-entry-stop",
			CommentKind: response.CommentKind,
			Message:     response.Message,
			CreatedAt:   now,
		})
	}
	s.taskThreads[response.TaskID] = threads
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

func copyAIAgentTaskThread(thread AIAgentTaskThreadRecord) AIAgentTaskThreadRecord {
	thread.Entries = append([]AIAgentTaskThreadEntry(nil), thread.Entries...)
	for i := range thread.Entries {
		thread.Entries[i].Lines = append([]AgentThreadProgressLine(nil), thread.Entries[i].Lines...)
	}
	return thread
}

func copyAgentThreadProgressEvent(event AgentThreadProgressEvent) AgentThreadProgressEvent {
	event.Lines = append([]AgentThreadProgressLine(nil), event.Lines...)
	return event
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
	case AgentThreadProgressEvent:
		return event.AgentID, true
	default:
		return "", false
	}
}
