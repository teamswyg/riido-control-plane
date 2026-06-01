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
	GetAIAgentDaemon(ctx context.Context, principal AuthorizationResult, agentID string) (DeviceDaemonDetailResponse, error)
	ControlAIAgentDaemon(ctx context.Context, principal AuthorizationResult, agentID string, action DaemonControlAction, req ControlDeviceDaemonRequest) (DeviceDaemonCommandResponse, error)
	ListAIAgentTaskAssignableAgents(ctx context.Context, principal AuthorizationResult, taskID string) (AgentClientListResponse, error)
	ListAIAgentTaskThreads(ctx context.Context, principal AuthorizationResult, taskID string) (AIAgentTaskThreadCollectionResponse, error)
	AssignAIAgentTask(ctx context.Context, principal AuthorizationResult, taskID string, req AssignAIAgentTaskRequest) (AIAgentTaskActionResponse, error)
	UnassignAIAgentTask(ctx context.Context, principal AuthorizationResult, taskID string, req UnassignAIAgentTaskRequest) (AIAgentTaskActionResponse, error)
	SubmitAIAgentTaskComment(ctx context.Context, principal AuthorizationResult, taskID string, req SubmitAIAgentTaskCommentRequest) (AIAgentTaskActionResponse, error)
	CreateAIAgentTaskThreadMessage(ctx context.Context, principal AuthorizationResult, taskID, threadID string, req CreateAIAgentTaskThreadMessageRequest) (AIAgentTaskActionResponse, error)
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
	mu                sync.Mutex
	workspaceID       string
	devices           []DeviceRecord
	daemons           map[string]DeviceDaemonRecord
	nextDaemonCommand int
	agents            map[string]AgentClientRecord
	templates         []AgentOnboardingTemplate
	taskThreads       map[string][]AIAgentTaskThreadRecord
	events            []ClientStreamEvent
	subscribers       map[int]aiAgentClientSubscriber
	nextSubscriberID  int
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
				Models: []RuntimeModelRecord{
					{ModelID: "codex-default", Label: "Codex 기본 모델", IsDefault: true},
				},
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
				Models: []RuntimeModelRecord{
					{ModelID: "claude-sonnect-4-6", Label: "Sonnect 4.6 (기본값)", IsDefault: true},
					{ModelID: "claude-opus-4-7", Label: "Opus 4.7", IsDefault: false},
					{ModelID: "claude-haiku-4-5", Label: "Haiku 4.5", IsDefault: false},
					{ModelID: "claude-opus-4-6", Label: "Opus 4.6", IsDefault: false},
					{ModelID: "claude-opus-3", Label: "Opus 3", IsDefault: false},
				},
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
				Models: []RuntimeModelRecord{
					{ModelID: "cursor-auto", Label: "Cursor Auto", IsDefault: true},
					{ModelID: "cursor-fast", Label: "Cursor Fast", IsDefault: false},
				},
			},
		},
	}
	daemon := DeviceDaemonRecord{
		DeviceID:          device.DeviceID,
		OwnerPrincipalID:  device.OwnerPrincipalID,
		DeviceDisplayName: device.DisplayName,
		DaemonID:          "daemon-mock-macbook",
		Profile:           "desktop-api.riido.ai",
		PID:               5111,
		UptimeSeconds:     74 * 60,
		StartedAt:         now.Add(-74 * time.Minute),
		LastSeenAt:        now,
		Availability:      DaemonAvailabilityOnline,
		ControlState:      DaemonControlStateIdle,
		SupportedActions:  []DaemonControlAction{DaemonControlActionRestart, DaemonControlActionStop},
	}
	sharedDevice := DeviceRecord{
		DeviceID:         "device-shared-studio",
		OwnerPrincipalID: "user-2",
		DisplayName:      "Shared Studio Mac",
		DaemonLastSeenAt: now,
		Runtimes: []RuntimeRecord{
			{
				RuntimeID:        "runtime-openclaw-shared",
				DeviceID:         "device-shared-studio",
				Kind:             RuntimeKindOpenClaw,
				Availability:     RuntimeAvailabilityOnline,
				DetectionState:   RuntimeDetectionStateDetected,
				OwnerPrincipalID: "user-2",
				LastDetectedAt:   now,
				HasAssignedAgent: true,
				Models: []RuntimeModelRecord{
					{ModelID: "openclaw-default", Label: "OpenClaw 기본 모델", IsDefault: true},
				},
			},
			{
				RuntimeID:        "runtime-cursor-private",
				DeviceID:         "device-shared-studio",
				Kind:             RuntimeKindCursor,
				Availability:     RuntimeAvailabilityOnline,
				DetectionState:   RuntimeDetectionStateDetected,
				OwnerPrincipalID: "user-2",
				LastDetectedAt:   now,
				HasAssignedAgent: true,
				Models: []RuntimeModelRecord{
					{ModelID: "cursor-auto", Label: "Cursor Auto", IsDefault: true},
					{ModelID: "cursor-fast", Label: "Cursor Fast", IsDefault: false},
				},
			},
		},
	}
	sharedDaemon := DeviceDaemonRecord{
		DeviceID:          sharedDevice.DeviceID,
		OwnerPrincipalID:  sharedDevice.OwnerPrincipalID,
		DeviceDisplayName: sharedDevice.DisplayName,
		DaemonID:          "daemon-shared-studio",
		Profile:           "desktop-api.riido.ai",
		PID:               6111,
		UptimeSeconds:     42 * 60,
		StartedAt:         now.Add(-42 * time.Minute),
		LastSeenAt:        now,
		Availability:      DaemonAvailabilityOnline,
		ControlState:      DaemonControlStateIdle,
		SupportedActions:  []DaemonControlAction{DaemonControlActionRestart, DaemonControlActionStop},
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
			ModelID:             "codex-default",
			ModelLabel:          "Codex 기본 모델",
			WorkStatus:          AgentWorkStatusRunning,
			Editability:         AgentEditabilityBlockedAssignedTasks,
			AssignedTaskCount:   1,
			CreatedAt:           now.Add(-72 * time.Hour),
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
			ModelID:             "claude-sonnect-4-6",
			ModelLabel:          "Sonnect 4.6 (기본값)",
			WorkStatus:          AgentWorkStatusOffline,
			Editability:         AgentEditabilityEditable,
			AssignedTaskCount:   0,
			CreatedAt:           now.Add(-96 * time.Hour),
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
			RuntimeID:           "runtime-openclaw-shared",
			RuntimeKind:         RuntimeKindOpenClaw,
			ModelID:             "openclaw-default",
			ModelLabel:          "OpenClaw 기본 모델",
			WorkStatus:          AgentWorkStatusIdle,
			Editability:         AgentEditabilityEditable,
			AssignedTaskCount:   0,
			CreatedAt:           now.Add(-48 * time.Hour),
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
			RuntimeID:           "runtime-cursor-private",
			RuntimeKind:         RuntimeKindCursor,
			ModelID:             "cursor-auto",
			ModelLabel:          "Cursor Auto",
			WorkStatus:          AgentWorkStatusIdle,
			Editability:         AgentEditabilityEditable,
			AssignedTaskCount:   0,
			CreatedAt:           now.Add(-24 * time.Hour),
			UpdatedAt:           now.Add(-3 * time.Hour),
		},
	}
	taskThreads := map[string][]AIAgentTaskThreadRecord{
		"task-1": {
			{
				ThreadID:        "thread-task-1-claude-1",
				TaskID:          "task-1",
				AgentID:         "agent-owned-claude",
				RunID:           "run-mock-completed-1",
				SourceCommentID: "comment-mock-1",
				WorkStatus:      AgentWorkStatusCompleted,
				AssignmentState: AgentAssignmentStateCompleted,
				CommentKind:     AgentTaskCommentTaskCompleted,
				Message:         "이전 AI Agent 작업이 완료됐어요.",
				StartedAt:       now.Add(-20 * time.Minute),
				CompletedAt:     now.Add(-15 * time.Minute),
				Lines: []AgentThreadProgressLine{
					{Seq: 1, Message: "팀 프로젝트 조회 완료 - 프로젝트 3건의 요약을 가져왔습니다.", ObservedAt: now.Add(-18 * time.Minute)},
				},
			},
			{
				ThreadID:        "thread-task-1-codex-2",
				TaskID:          "task-1",
				AgentID:         "agent-owned-codex",
				RunID:           "run-mock-1",
				SourceCommentID: "comment-mock-2",
				WorkStatus:      AgentWorkStatusRunning,
				AssignmentState: AgentAssignmentStateRunning,
				CommentKind:     AgentTaskCommentRuntimeProgress,
				Message:         "팀 프로젝트 수집 중 - 팀의 프로젝트 목록을 조회 중.",
				StartedAt:       now.Add(-3 * time.Minute),
				Lines: []AgentThreadProgressLine{
					{Seq: 1, Message: "생각 중...", ObservedAt: now.Add(-3 * time.Minute)},
					{Seq: 2, Message: "팀 프로젝트 수집 중 - 팀의 프로젝트 목록, 진행 상태, 우선순위와 담당자 정보를 조회해 요약을 준비 중.", ObservedAt: now.Add(-2 * time.Minute)},
				},
			},
		},
	}
	return &MockAIAgentClientStore{
		workspaceID:       "workspace-mock-riid",
		devices:           []DeviceRecord{device, sharedDevice},
		daemons:           map[string]DeviceDaemonRecord{device.DeviceID: daemon, sharedDevice.DeviceID: sharedDaemon},
		nextDaemonCommand: 1,
		agents:            agents,
		templates: []AgentOnboardingTemplate{
			{
				TemplateID:          "riido_pm",
				Name:                "리도",
				RoleLabel:           "PM Agent",
				ProfileThumbnailURL: "https://cdn.riido.io/mock/ai-agent-templates/riido-pm.png",
				Description:         "문제 정의부터 우선순위, 출시 계획까지 정리합니다.",
				Instruction:         "기능 요청을 문제, 목표, 성공 기준으로 재정의하고 PRD, 우선순위, 로드맵, 출시 계획을 구조화합니다. 아이디어는 가설로 다루며 불확실한 내용은 [확인 필요]로 표시합니다.",
			},
			{
				TemplateID:          "yeongsil_backend",
				Name:                "영실",
				RoleLabel:           "Backend Agent",
				ProfileThumbnailURL: "https://cdn.riido.io/mock/ai-agent-templates/yeongsil-backend.png",
				Description:         "서버 구조를 설계하고, API와 데이터 흐름을 안정적으로 구현합니다.",
				Instruction:         "요구사항을 API, 데이터 흐름, 저장 경계, 실패 처리 기준으로 나누고 안정적인 서버 구현 계획을 제안합니다.",
			},
			{
				TemplateID:          "hongdo_frontend",
				Name:                "홍도",
				RoleLabel:           "Frontend Agent",
				ProfileThumbnailURL: "https://cdn.riido.io/mock/ai-agent-templates/hongdo-frontend.png",
				Description:         "사용자가 보는 화면을 구현하고, 성능과 접근성을 개선합니다.",
				Instruction:         "화면 구조, 상태, 접근성, 성능을 함께 검토하고 사용자에게 자연스러운 프론트엔드 구현을 제안합니다.",
			},
			{
				TemplateID:          "jiwon_research",
				Name:                "지원",
				RoleLabel:           "Research Agent",
				ProfileThumbnailURL: "https://cdn.riido.io/mock/ai-agent-templates/jiwon-research.png",
				Description:         "시장과 경쟁사를 조사하고, 의사결정에 필요한 인사이트를 정리합니다.",
				Instruction:         "시장, 경쟁사, 사용자 맥락을 조사하고 의사결정에 필요한 근거와 확인이 필요한 가정을 분리해 정리합니다.",
			},
		},
		taskThreads: taskThreads,
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
				EventType: AgentClientEventDeviceDaemonStatus,
				Payload: DeviceDaemonStatusEvent{
					EventType:     AgentClientEventDeviceDaemonStatus,
					SchemaVersion: SchemaVersion,
					Daemon:        daemon,
				},
			},
			{
				Seq:       3,
				EventType: AgentClientEventWorkStatusChanged,
				Payload: AgentWorkStatusChangedEvent{
					EventType:       AgentClientEventWorkStatusChanged,
					SchemaVersion:   SchemaVersion,
					AgentID:         "agent-owned-codex",
					TaskID:          "task-1",
					ThreadID:        "thread-task-1-codex-2",
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
		SchemaVersion:  SchemaVersion,
		ClientKind:     normalizeClientKind(clientKind),
		WorkspaceID:    s.workspaceID,
		Agents:         s.visibleAgents(principal),
		Devices:        s.visibleDevicesLocked(principal),
		AgentTemplates: copyAgentTemplates(s.templates),
	}, nil
}

func (s *MockAIAgentClientStore) ListAIAgentDevices(ctx context.Context, principal AuthorizationResult) (DeviceRuntimeListResponse, error) {
	if err := ctx.Err(); err != nil {
		return DeviceRuntimeListResponse{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return DeviceRuntimeListResponse{SchemaVersion: SchemaVersion, Devices: s.visibleDevicesLocked(principal)}, nil
}

func (s *MockAIAgentClientStore) GetAIAgentDaemon(ctx context.Context, principal AuthorizationResult, agentID string) (DeviceDaemonDetailResponse, error) {
	if err := ctx.Err(); err != nil {
		return DeviceDaemonDetailResponse{}, err
	}
	agentID = strings.TrimSpace(agentID)
	if agentID == "" {
		return DeviceDaemonDetailResponse{}, errors.New("agent_id is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	_, daemon, ok := s.deviceDaemonForAgentAccessLocked(principal, agentID)
	if !ok {
		return DeviceDaemonDetailResponse{}, ErrAIAgentNotFound
	}
	return DeviceDaemonDetailResponse{SchemaVersion: SchemaVersion, Daemon: copyDeviceDaemon(daemon)}, nil
}

func (s *MockAIAgentClientStore) ControlAIAgentDaemon(ctx context.Context, principal AuthorizationResult, agentID string, action DaemonControlAction, req ControlDeviceDaemonRequest) (DeviceDaemonCommandResponse, error) {
	if err := ctx.Err(); err != nil {
		return DeviceDaemonCommandResponse{}, err
	}
	agentID = strings.TrimSpace(agentID)
	if agentID == "" {
		return DeviceDaemonCommandResponse{}, errors.New("agent_id is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	agent, daemon, ok := s.deviceDaemonForAgentAccessLocked(principal, agentID)
	if !ok {
		return DeviceDaemonCommandResponse{}, ErrAIAgentNotFound
	}
	if !daemonSupportsAction(daemon, action) {
		return DeviceDaemonCommandResponse{}, errors.New("daemon action is not supported in the current state")
	}
	now := time.Now().UTC()
	commandID := "daemon-command-" + strconv.Itoa(s.nextDaemonCommand)
	s.nextDaemonCommand++
	daemon.LastCommandID = commandID
	daemon.LastCommandAction = action
	daemon.LastCommandRequestedAt = now
	daemon.LastSeenAt = now
	message := "daemon command accepted"
	switch action {
	case DaemonControlActionStart:
		daemon.Availability = DaemonAvailabilityOnline
		daemon.ControlState = DaemonControlStateStarting
		daemon.StartedAt = now
		daemon.PID = 5111
		daemon.UptimeSeconds = 0
		daemon.SupportedActions = []DaemonControlAction{DaemonControlActionRestart, DaemonControlActionStop}
		message = "daemon start command accepted"
	case DaemonControlActionRestart:
		daemon.Availability = DaemonAvailabilityOnline
		daemon.ControlState = DaemonControlStateRestarting
		daemon.StartedAt = now
		daemon.UptimeSeconds = 0
		daemon.SupportedActions = []DaemonControlAction{DaemonControlActionStop}
		message = "daemon restart command accepted"
	case DaemonControlActionStop:
		daemon.Availability = DaemonAvailabilityOffline
		daemon.ControlState = DaemonControlStateStopping
		daemon.PID = 0
		daemon.UptimeSeconds = 0
		daemon.SupportedActions = []DaemonControlAction{DaemonControlActionStart}
		s.markDeviceRuntimesOfflineLocked(daemon.DeviceID, now)
		message = "daemon stop command accepted"
	default:
		return DeviceDaemonCommandResponse{}, errors.New("unsupported daemon action")
	}
	s.daemons[daemon.DeviceID] = daemon
	s.appendClientEventLocked(AgentClientEventDeviceDaemonStatus, DeviceDaemonStatusEvent{
		EventType:     AgentClientEventDeviceDaemonStatus,
		SchemaVersion: SchemaVersion,
		Daemon:        copyDeviceDaemon(daemon),
	})
	if action == DaemonControlActionStop {
		if device, ok := s.deviceByIDLocked(daemon.DeviceID); ok {
			s.appendClientEventLocked(AgentClientEventDeviceRuntimeSnapshot, DeviceRuntimeSnapshotEvent{
				EventType:     AgentClientEventDeviceRuntimeSnapshot,
				SchemaVersion: SchemaVersion,
				Device:        device,
			})
		}
	}
	return DeviceDaemonCommandResponse{
		SchemaVersion: SchemaVersion,
		CommandID:     commandID,
		DeviceID:      daemon.DeviceID,
		Action:        action,
		Availability:  daemon.Availability,
		ControlState:  daemon.ControlState,
		AcceptedAt:    now,
		Message:       message + " for agent " + agent.AgentID,
	}, nil
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

func (s *MockAIAgentClientStore) ListAIAgentTaskThreads(ctx context.Context, principal AuthorizationResult, taskID string) (AIAgentTaskThreadCollectionResponse, error) {
	if err := ctx.Err(); err != nil {
		return AIAgentTaskThreadCollectionResponse{}, err
	}
	taskID = strings.TrimSpace(taskID)
	if taskID == "" {
		return AIAgentTaskThreadCollectionResponse{}, errors.New("task_id is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	threads := s.visibleTaskThreadsLocked(principal, taskID)
	response := AIAgentTaskThreadCollectionResponse{
		SchemaVersion: SchemaVersion,
		TaskID:        taskID,
		Threads:       threads,
	}
	for i := range threads {
		if !taskThreadHasActiveStream(threads[i]) {
			continue
		}
		link := activeStreamLinkForThread(threads[i])
		response.ActiveStream = &link
		break
	}
	return response, nil
}

func (s *MockAIAgentClientStore) AssignAIAgentTask(ctx context.Context, principal AuthorizationResult, taskID string, req AssignAIAgentTaskRequest) (AIAgentTaskActionResponse, error) {
	if err := ctx.Err(); err != nil {
		return AIAgentTaskActionResponse{}, err
	}
	taskID = strings.TrimSpace(taskID)
	req.AgentID = strings.TrimSpace(req.AgentID)
	if taskID == "" {
		return AIAgentTaskActionResponse{}, errors.New("task_id is required")
	}
	if req.AgentID == "" {
		return AIAgentTaskActionResponse{}, errors.New("agent_id is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	agent, ok := s.visibleAgent(principal, req.AgentID)
	if !ok {
		return AIAgentTaskActionResponse{}, ErrAIAgentNotFound
	}
	if thread, ok := s.activeTaskThreadForAgentLocked(taskID, agent.AgentID); ok {
		return actionResponseFromThread(thread), nil
	}
	s.markTaskActiveThreadsStoppedLocked(taskID, AgentTaskCommentStoppedByUserRequest, "agent assignment was replaced by a participant change")
	response := AIAgentTaskActionResponse{
		SchemaVersion:   SchemaVersion,
		TaskID:          taskID,
		AgentID:         agent.AgentID,
		RunID:           "run-mock-assignment-" + taskID + "-" + strconv.Itoa(len(s.taskThreads[taskID])+1),
		WorkStatus:      AgentWorkStatusRunning,
		AssignmentState: AgentAssignmentStateRunning,
		CommentKind:     AgentTaskCommentAssignmentStarted,
		Message:         "agent assignment started from task participant",
	}
	response.ThreadID = threadIDForRun(response.TaskID, response.AgentID, response.RunID)
	if agent.WorkStatus == AgentWorkStatusRunning || agent.WorkStatus == AgentWorkStatusWaitingForUser || agent.WorkStatus == AgentWorkStatusQueued {
		response.WorkStatus = AgentWorkStatusQueued
		response.AssignmentState = AgentAssignmentStateQueued
		response.CommentKind = AgentTaskCommentQueuedByBusyAgent
		response.Message = "agent is busy; task assignment was queued"
	}
	agent.WorkStatus = response.WorkStatus
	agent.Editability = AgentEditabilityBlockedAssignedTasks
	agent.AssignedTaskCount++
	s.agents[agent.AgentID] = agent
	s.upsertTaskThreadFromActionLocked(response, "")
	s.appendAgentTaskActionEvent(response)
	return response, nil
}

func (s *MockAIAgentClientStore) UnassignAIAgentTask(ctx context.Context, principal AuthorizationResult, taskID string, req UnassignAIAgentTaskRequest) (AIAgentTaskActionResponse, error) {
	if err := ctx.Err(); err != nil {
		return AIAgentTaskActionResponse{}, err
	}
	taskID = strings.TrimSpace(taskID)
	req.AgentID = strings.TrimSpace(req.AgentID)
	if taskID == "" {
		return AIAgentTaskActionResponse{}, errors.New("task_id is required")
	}
	if req.AgentID == "" {
		return AIAgentTaskActionResponse{}, errors.New("agent_id is required")
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
		RunID:           "run-mock-unassign-" + taskID,
		WorkStatus:      AgentWorkStatusIdle,
		AssignmentState: AgentAssignmentStateStopped,
		CommentKind:     AgentTaskCommentStoppedByUserRequest,
		Message:         "agent work was stopped by task participant removal",
	}
	if thread, ok := s.activeTaskThreadForAgentLocked(taskID, agent.AgentID); ok {
		response.ThreadID = thread.ThreadID
		response.RunID = thread.RunID
	} else {
		response.ThreadID = threadIDForRun(response.TaskID, response.AgentID, response.RunID)
	}
	s.markTaskAgentThreadsStoppedLocked(taskID, agent.AgentID, AgentTaskCommentStoppedByUserRequest, response.Message)
	agent.WorkStatus = AgentWorkStatusIdle
	if agent.AssignedTaskCount > 0 {
		agent.AssignedTaskCount--
	}
	agent.Editability = editabilityForAssignedTasks(agent.AssignedTaskCount)
	s.agents[agent.AgentID] = agent
	s.upsertTaskThreadFromActionLocked(response, "")
	s.appendAgentTaskActionEvent(response)
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
	response.ThreadID = threadIDForRun(response.TaskID, response.AgentID, response.RunID)
	if agent.WorkStatus == AgentWorkStatusRunning || agent.WorkStatus == AgentWorkStatusWaitingForUser || agent.WorkStatus == AgentWorkStatusQueued {
		response.WorkStatus = AgentWorkStatusQueued
		response.AssignmentState = AgentAssignmentStateQueued
		response.CommentKind = AgentTaskCommentQueuedByBusyAgent
		response.Message = "agent is busy; task comment was queued"
	}
	s.upsertTaskThreadFromActionLocked(response, req.SourceCommentID)
	s.appendAgentTaskActionEvent(response)
	return response, nil
}

func (s *MockAIAgentClientStore) CreateAIAgentTaskThreadMessage(ctx context.Context, principal AuthorizationResult, taskID, threadID string, req CreateAIAgentTaskThreadMessageRequest) (AIAgentTaskActionResponse, error) {
	if err := ctx.Err(); err != nil {
		return AIAgentTaskActionResponse{}, err
	}
	taskID = strings.TrimSpace(taskID)
	threadID = strings.TrimSpace(threadID)
	req.Body = strings.TrimSpace(req.Body)
	if taskID == "" {
		return AIAgentTaskActionResponse{}, errors.New("task_id is required")
	}
	if threadID == "" {
		return AIAgentTaskActionResponse{}, errors.New("thread_id is required")
	}
	if req.Body == "" {
		return AIAgentTaskActionResponse{}, errors.New("body is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	thread, ok := s.visibleTaskThreadLocked(principal, taskID, threadID)
	if !ok {
		return AIAgentTaskActionResponse{}, ErrAIAgentNotFound
	}
	agent, ok := s.visibleAgent(principal, thread.AgentID)
	if !ok {
		return AIAgentTaskActionResponse{}, ErrAIAgentNotFound
	}
	if active, ok := s.activeTaskThreadLocked(taskID); ok && active.ThreadID != threadID {
		return AIAgentTaskActionResponse{}, ErrAIAgentTaskThreadConflict
	}
	response := AIAgentTaskActionResponse{
		SchemaVersion:   SchemaVersion,
		TaskID:          taskID,
		AgentID:         agent.AgentID,
		ThreadID:        threadID,
		RunID:           thread.RunID,
		WorkStatus:      AgentWorkStatusRunning,
		AssignmentState: AgentAssignmentStateRunning,
		CommentKind:     AgentTaskCommentRuntimeProgress,
		Message:         "agent work continued from task thread message",
	}
	threadWasActive := taskThreadHasActiveStream(thread)
	if !threadWasActive {
		response.RunID = "run-mock-message-" + taskID + "-" + threadID
	}
	if !threadWasActive && (agent.WorkStatus == AgentWorkStatusRunning || agent.WorkStatus == AgentWorkStatusWaitingForUser || agent.WorkStatus == AgentWorkStatusQueued) {
		response.WorkStatus = AgentWorkStatusQueued
		response.AssignmentState = AgentAssignmentStateQueued
		response.CommentKind = AgentTaskCommentQueuedByBusyAgent
		response.Message = "agent is busy; task thread message was queued"
	}
	s.upsertTaskThreadMessageFromActionLocked(response, req.SourceMessageID)
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
	if thread, ok := s.activeTaskThreadForAgentLocked(taskID, agent.AgentID); ok {
		response.ThreadID = thread.ThreadID
		response.RunID = thread.RunID
	} else {
		response.ThreadID = threadIDForRun(response.TaskID, response.AgentID, response.RunID)
	}
	s.markTaskAgentThreadsStoppedLocked(taskID, agent.AgentID, AgentTaskCommentStoppedByUserRequest, response.Message)
	s.upsertTaskThreadFromActionLocked(response, "")
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
	runtimeKind, runtimeModel, ok := runtimeSelectionForPrincipal(s.devices, principal, runtimeID, req.ModelID)
	if !ok {
		return AgentClientRecordResponse{}, errors.New("runtime_id or model_id is not available")
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
		ModelID:             runtimeModel.ModelID,
		ModelLabel:          runtimeModel.Label,
		WorkStatus:          AgentWorkStatusIdle,
		Editability:         AgentEditabilityEditable,
		AssignedTaskCount:   0,
		CreatedAt:           now,
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
	nextRuntimeID := agent.RuntimeID
	if strings.TrimSpace(req.RuntimeID) != "" {
		nextRuntimeID = strings.TrimSpace(req.RuntimeID)
	}
	if strings.TrimSpace(req.RuntimeID) != "" || req.ModelID != nil {
		runtimeKind, runtimeModel, ok := runtimeSelectionForPrincipal(s.devices, principal, nextRuntimeID, req.ModelID)
		if !ok {
			return AgentClientRecordResponse{}, errors.New("runtime_id or model_id is not available")
		}
		agent.RuntimeID = nextRuntimeID
		agent.RuntimeKind = runtimeKind
		agent.ModelID = runtimeModel.ModelID
		agent.ModelLabel = runtimeModel.Label
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
	s.markAgentTaskThreadsStoppedLocked(agent.AgentID, AgentTaskCommentStoppedByAgentDeleted, "에이전트가 삭제되어 진행 중이던 작업이 중지됐어요.")
	s.appendClientEventLocked(AgentClientEventWorkStatusChanged, AgentWorkStatusChangedEvent{
		EventType:       AgentClientEventWorkStatusChanged,
		SchemaVersion:   SchemaVersion,
		AgentID:         agent.AgentID,
		TaskID:          "task-1",
		ThreadID:        "thread-task-1-codex-2",
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
	req.ThreadID = strings.TrimSpace(req.ThreadID)
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
	if req.ThreadID == "" {
		req.ThreadID = threadIDForRun(req.TaskID, agentID, req.RunID)
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
		ThreadID:        req.ThreadID,
		RunID:           req.RunID,
		WorkStatus:      AgentWorkStatusRunning,
		AssignmentState: AgentAssignmentStateRunning,
		CommentKind:     AgentTaskCommentRuntimeProgress,
		BatchStartedAt:  req.BatchStartedAt,
		BatchEndedAt:    req.BatchEndedAt,
		Lines:           lines,
	}
	s.appendThreadProgressLocked(event)
	s.appendClientEventLocked(event.EventType, event)
	return AgentThreadProgressBatchResponse{
		SchemaVersion: SchemaVersion,
		AcceptedLines: len(lines),
		Event:         event,
	}, nil
}

var (
	ErrAIAgentNotFound           = errors.New("ai agent not found")
	ErrAIAgentAssigned           = errors.New("ai agent has assigned tasks")
	ErrAIAgentTaskThreadConflict = errors.New("task already has another active ai agent thread")
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

func (s *MockAIAgentClientStore) visibleTaskThreadsLocked(principal AuthorizationResult, taskID string) []AIAgentTaskThreadRecord {
	source := s.taskThreads[taskID]
	out := make([]AIAgentTaskThreadRecord, 0, len(source))
	for _, thread := range source {
		agent, ok := s.agents[thread.AgentID]
		if !ok || !aiAgentVisibleTo(principal, agent) {
			continue
		}
		out = append(out, copyTaskThread(thread))
	}
	return out
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
		return copyTaskThread(thread), true
	}
	return AIAgentTaskThreadRecord{}, false
}

func (s *MockAIAgentClientStore) activeTaskThreadLocked(taskID string) (AIAgentTaskThreadRecord, bool) {
	threads := s.taskThreads[taskID]
	for i := len(threads) - 1; i >= 0; i-- {
		thread := threads[i]
		if taskThreadHasActiveStream(thread) {
			return copyTaskThread(thread), true
		}
	}
	return AIAgentTaskThreadRecord{}, false
}

func (s *MockAIAgentClientStore) activeTaskThreadForAgentLocked(taskID, agentID string) (AIAgentTaskThreadRecord, bool) {
	threads := s.taskThreads[taskID]
	for i := len(threads) - 1; i >= 0; i-- {
		thread := threads[i]
		if thread.AgentID == agentID && taskThreadHasActiveStream(thread) {
			return copyTaskThread(thread), true
		}
	}
	return AIAgentTaskThreadRecord{}, false
}

func actionResponseFromThread(thread AIAgentTaskThreadRecord) AIAgentTaskActionResponse {
	return AIAgentTaskActionResponse{
		SchemaVersion:   SchemaVersion,
		TaskID:          thread.TaskID,
		AgentID:         thread.AgentID,
		ThreadID:        thread.ThreadID,
		RunID:           thread.RunID,
		WorkStatus:      thread.WorkStatus,
		AssignmentState: thread.AssignmentState,
		CommentKind:     thread.CommentKind,
		Message:         thread.Message,
	}
}

func (s *MockAIAgentClientStore) upsertTaskThreadFromActionLocked(response AIAgentTaskActionResponse, sourceCommentID string) {
	now := time.Now().UTC()
	thread := AIAgentTaskThreadRecord{
		ThreadID:        response.ThreadID,
		TaskID:          response.TaskID,
		AgentID:         response.AgentID,
		RunID:           response.RunID,
		SourceCommentID: strings.TrimSpace(sourceCommentID),
		WorkStatus:      response.WorkStatus,
		AssignmentState: response.AssignmentState,
		CommentKind:     response.CommentKind,
		Message:         response.Message,
		StartedAt:       now,
		Lines:           []AgentThreadProgressLine{},
	}
	if !taskThreadHasActiveStream(thread) {
		thread.CompletedAt = now
	}
	threads := s.taskThreads[response.TaskID]
	for i := range threads {
		if threads[i].ThreadID != response.ThreadID {
			continue
		}
		threads[i].WorkStatus = response.WorkStatus
		threads[i].AssignmentState = response.AssignmentState
		threads[i].CommentKind = response.CommentKind
		threads[i].Message = response.Message
		if sourceCommentID != "" {
			threads[i].SourceCommentID = sourceCommentID
		}
		if threads[i].StartedAt.IsZero() {
			threads[i].StartedAt = now
		}
		if !taskThreadHasActiveStream(threads[i]) {
			threads[i].CompletedAt = now
		}
		s.taskThreads[response.TaskID] = threads
		return
	}
	s.taskThreads[response.TaskID] = append(threads, thread)
}

func (s *MockAIAgentClientStore) upsertTaskThreadMessageFromActionLocked(response AIAgentTaskActionResponse, sourceMessageID string) {
	now := time.Now().UTC()
	thread := AIAgentTaskThreadRecord{
		ThreadID:        response.ThreadID,
		TaskID:          response.TaskID,
		AgentID:         response.AgentID,
		RunID:           response.RunID,
		SourceMessageID: strings.TrimSpace(sourceMessageID),
		WorkStatus:      response.WorkStatus,
		AssignmentState: response.AssignmentState,
		CommentKind:     response.CommentKind,
		Message:         response.Message,
		StartedAt:       now,
		Lines:           []AgentThreadProgressLine{},
	}
	if !taskThreadHasActiveStream(thread) {
		thread.CompletedAt = now
	}
	threads := s.taskThreads[response.TaskID]
	for i := range threads {
		if threads[i].ThreadID != response.ThreadID {
			continue
		}
		threads[i].RunID = response.RunID
		threads[i].WorkStatus = response.WorkStatus
		threads[i].AssignmentState = response.AssignmentState
		threads[i].CommentKind = response.CommentKind
		threads[i].Message = response.Message
		if sourceMessageID != "" {
			threads[i].SourceMessageID = sourceMessageID
		}
		if threads[i].StartedAt.IsZero() {
			threads[i].StartedAt = now
		}
		if taskThreadHasActiveStream(threads[i]) {
			threads[i].CompletedAt = time.Time{}
		} else {
			threads[i].CompletedAt = now
		}
		s.taskThreads[response.TaskID] = threads
		return
	}
	s.taskThreads[response.TaskID] = append(threads, thread)
}

func (s *MockAIAgentClientStore) appendThreadProgressLocked(event AgentThreadProgressEvent) {
	now := time.Now().UTC()
	threads := s.taskThreads[event.TaskID]
	for i := range threads {
		if threads[i].ThreadID != event.ThreadID {
			continue
		}
		threads[i].RunID = event.RunID
		threads[i].WorkStatus = event.WorkStatus
		threads[i].AssignmentState = event.AssignmentState
		threads[i].CommentKind = event.CommentKind
		if len(event.Lines) > 0 {
			threads[i].Message = event.Lines[len(event.Lines)-1].Message
		}
		threads[i].Lines = append(threads[i].Lines, copyProgressLines(event.Lines)...)
		s.taskThreads[event.TaskID] = threads
		return
	}
	message := "agent progress updated"
	if len(event.Lines) > 0 {
		message = event.Lines[len(event.Lines)-1].Message
	}
	s.taskThreads[event.TaskID] = append(threads, AIAgentTaskThreadRecord{
		ThreadID:        event.ThreadID,
		TaskID:          event.TaskID,
		AgentID:         event.AgentID,
		RunID:           event.RunID,
		WorkStatus:      event.WorkStatus,
		AssignmentState: event.AssignmentState,
		CommentKind:     event.CommentKind,
		Message:         message,
		StartedAt:       now,
		Lines:           copyProgressLines(event.Lines),
	})
}

func (s *MockAIAgentClientStore) markTaskActiveThreadsStoppedLocked(taskID string, kind AgentTaskCommentKind, message string) {
	now := time.Now().UTC()
	threads := s.taskThreads[taskID]
	for i := range threads {
		if !taskThreadHasActiveStream(threads[i]) {
			continue
		}
		threads[i].WorkStatus = AgentWorkStatusIdle
		threads[i].AssignmentState = AgentAssignmentStateStopped
		threads[i].CommentKind = kind
		threads[i].Message = message
		threads[i].CompletedAt = now
		if agent := s.agents[threads[i].AgentID]; agent.AgentID != "" && agent.AssignedTaskCount > 0 {
			agent.AssignedTaskCount--
			agent.Editability = editabilityForAssignedTasks(agent.AssignedTaskCount)
			if agent.AssignedTaskCount == 0 {
				agent.WorkStatus = AgentWorkStatusIdle
			}
			s.agents[agent.AgentID] = agent
		}
	}
	s.taskThreads[taskID] = threads
}

func (s *MockAIAgentClientStore) markAgentTaskThreadsStoppedLocked(agentID string, kind AgentTaskCommentKind, message string) {
	now := time.Now().UTC()
	for taskID, threads := range s.taskThreads {
		for i := range threads {
			if threads[i].AgentID != agentID || !taskThreadHasActiveStream(threads[i]) {
				continue
			}
			threads[i].WorkStatus = AgentWorkStatusOffline
			threads[i].AssignmentState = AgentAssignmentStateStopped
			threads[i].CommentKind = kind
			threads[i].Message = message
			threads[i].CompletedAt = now
		}
		s.taskThreads[taskID] = threads
	}
}

func (s *MockAIAgentClientStore) markTaskAgentThreadsStoppedLocked(taskID, agentID string, kind AgentTaskCommentKind, message string) {
	now := time.Now().UTC()
	threads := s.taskThreads[taskID]
	for i := range threads {
		if threads[i].AgentID != agentID || !taskThreadHasActiveStream(threads[i]) {
			continue
		}
		threads[i].WorkStatus = AgentWorkStatusIdle
		threads[i].AssignmentState = AgentAssignmentStateStopped
		threads[i].CommentKind = kind
		threads[i].Message = message
		threads[i].CompletedAt = now
	}
	s.taskThreads[taskID] = threads
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
		ThreadID:        response.ThreadID,
		RunID:           response.RunID,
		WorkStatus:      response.WorkStatus,
		AssignmentState: response.AssignmentState,
		CommentKind:     response.CommentKind,
	})
}

func (s *MockAIAgentClientStore) clientEventsForPrincipalLocked(principal AuthorizationResult) []ClientStreamEvent {
	events := make([]ClientStreamEvent, 0, len(s.events))
	for _, event := range s.events {
		visible, ok := clientEventForPrincipalLocked(s, principal, event)
		if !ok {
			continue
		}
		events = append(events, visible)
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
		visible, ok := clientEventForPrincipalLocked(s, subscriber.principal, event)
		if !ok {
			continue
		}
		select {
		case subscriber.events <- visible:
		default:
		}
	}
	return event
}

func clientEventForPrincipalLocked(s *MockAIAgentClientStore, principal AuthorizationResult, event ClientStreamEvent) (ClientStreamEvent, bool) {
	if deviceEvent, ok := event.Payload.(DeviceRuntimeSnapshotEvent); ok {
		device, ok := s.visibleDeviceRecordLocked(principal, deviceEvent.Device)
		if !ok {
			return ClientStreamEvent{}, false
		}
		deviceEvent.Device = device
		event.Payload = deviceEvent
		return event, true
	}
	if !clientEventVisibleToLocked(s, principal, event) {
		return ClientStreamEvent{}, false
	}
	return event, true
}

func clientEventVisibleToLocked(s *MockAIAgentClientStore, principal AuthorizationResult, event ClientStreamEvent) bool {
	if daemon, ok := eventDeviceDaemon(event.Payload); ok {
		return s.daemonVisibleToPrincipalLocked(principal, daemon)
	}
	if device, ok := eventDeviceRecord(event.Payload); ok {
		return s.deviceVisibleToPrincipalLocked(principal, device)
	}
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

func eventDeviceDaemon(payload any) (DeviceDaemonRecord, bool) {
	switch event := payload.(type) {
	case DeviceDaemonStatusEvent:
		return event.Daemon, true
	default:
		return DeviceDaemonRecord{}, false
	}
}

func eventDeviceRecord(payload any) (DeviceRecord, bool) {
	switch event := payload.(type) {
	case DeviceRuntimeSnapshotEvent:
		return event.Device, true
	default:
		return DeviceRecord{}, false
	}
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

func (s *MockAIAgentClientStore) visibleDevicesLocked(principal AuthorizationResult) []DeviceRecord {
	if aiAgentIsAdmin(principal) {
		return copyDevices(s.devices)
	}
	visibleRuntimeIDs := map[string]struct{}{}
	for _, agent := range s.agents {
		if !aiAgentVisibleTo(principal, agent) || agent.RuntimeID == "" {
			continue
		}
		visibleRuntimeIDs[agent.RuntimeID] = struct{}{}
	}
	out := make([]DeviceRecord, 0, len(s.devices))
	for _, device := range s.devices {
		if device.OwnerPrincipalID == principal.PrincipalID {
			out = append(out, copyDevice(device))
			continue
		}
		filtered := device
		filtered.Runtimes = nil
		for _, runtime := range device.Runtimes {
			if _, ok := visibleRuntimeIDs[runtime.RuntimeID]; ok {
				filtered.Runtimes = append(filtered.Runtimes, copyRuntime(runtime))
			}
		}
		if len(filtered.Runtimes) > 0 {
			out = append(out, filtered)
		}
	}
	return out
}

func (s *MockAIAgentClientStore) visibleDeviceRecordLocked(principal AuthorizationResult, device DeviceRecord) (DeviceRecord, bool) {
	if aiAgentIsAdmin(principal) || device.OwnerPrincipalID == principal.PrincipalID {
		return copyDevice(device), true
	}
	filtered := device
	filtered.Runtimes = nil
	for _, runtime := range device.Runtimes {
		if s.runtimeVisibleThroughAgentLocked(principal, runtime.RuntimeID) {
			filtered.Runtimes = append(filtered.Runtimes, copyRuntime(runtime))
		}
	}
	if len(filtered.Runtimes) == 0 {
		return DeviceRecord{}, false
	}
	return filtered, true
}

func (s *MockAIAgentClientStore) deviceVisibleToPrincipalLocked(principal AuthorizationResult, device DeviceRecord) bool {
	if aiAgentIsAdmin(principal) || device.OwnerPrincipalID == principal.PrincipalID {
		return true
	}
	for _, runtime := range device.Runtimes {
		if s.runtimeVisibleThroughAgentLocked(principal, runtime.RuntimeID) {
			return true
		}
	}
	return false
}

func (s *MockAIAgentClientStore) daemonVisibleToPrincipalLocked(principal AuthorizationResult, daemon DeviceDaemonRecord) bool {
	if aiAgentIsAdmin(principal) || daemon.OwnerPrincipalID == principal.PrincipalID {
		return true
	}
	for _, device := range s.devices {
		if device.DeviceID != daemon.DeviceID {
			continue
		}
		return s.deviceVisibleToPrincipalLocked(principal, device)
	}
	return false
}

func (s *MockAIAgentClientStore) runtimeVisibleThroughAgentLocked(principal AuthorizationResult, runtimeID string) bool {
	for _, agent := range s.agents {
		if agent.RuntimeID == runtimeID && aiAgentVisibleTo(principal, agent) {
			return true
		}
	}
	return false
}

func (s *MockAIAgentClientStore) deviceDaemonForAgentAccessLocked(principal AuthorizationResult, agentID string) (AgentClientRecord, DeviceDaemonRecord, bool) {
	agent, ok := s.visibleAgent(principal, agentID)
	if !ok {
		return AgentClientRecord{}, DeviceDaemonRecord{}, false
	}
	device, ok := s.deviceByRuntimeIDLocked(agent.RuntimeID)
	if !ok {
		return AgentClientRecord{}, DeviceDaemonRecord{}, false
	}
	daemon, ok := s.daemons[device.DeviceID]
	if !ok {
		return AgentClientRecord{}, DeviceDaemonRecord{}, false
	}
	return agent, copyDeviceDaemon(daemon), true
}

func (s *MockAIAgentClientStore) deviceDaemonForOwnerLocked(principal AuthorizationResult, deviceID string) (DeviceDaemonRecord, bool) {
	for _, device := range s.devices {
		if device.DeviceID != deviceID || device.OwnerPrincipalID != principal.PrincipalID {
			continue
		}
		daemon, ok := s.daemons[deviceID]
		if !ok {
			return DeviceDaemonRecord{}, false
		}
		return copyDeviceDaemon(daemon), true
	}
	return DeviceDaemonRecord{}, false
}

func (s *MockAIAgentClientStore) deviceByRuntimeIDLocked(runtimeID string) (DeviceRecord, bool) {
	for _, device := range s.devices {
		for _, runtime := range device.Runtimes {
			if runtime.RuntimeID == runtimeID {
				return copyDevice(device), true
			}
		}
	}
	return DeviceRecord{}, false
}

func (s *MockAIAgentClientStore) deviceByIDLocked(deviceID string) (DeviceRecord, bool) {
	for _, device := range s.devices {
		if device.DeviceID == deviceID {
			return copyDevice(device), true
		}
	}
	return DeviceRecord{}, false
}

func (s *MockAIAgentClientStore) markDeviceRuntimesOfflineLocked(deviceID string, observedAt time.Time) {
	for deviceIndex := range s.devices {
		if s.devices[deviceIndex].DeviceID != deviceID {
			continue
		}
		s.devices[deviceIndex].DaemonLastSeenAt = observedAt
		for runtimeIndex := range s.devices[deviceIndex].Runtimes {
			s.devices[deviceIndex].Runtimes[runtimeIndex].Availability = RuntimeAvailabilityOffline
			s.devices[deviceIndex].Runtimes[runtimeIndex].DetectionState = RuntimeDetectionStateMissing
			s.devices[deviceIndex].Runtimes[runtimeIndex].LastDetectedAt = observedAt
		}
		return
	}
}

func daemonSupportsAction(daemon DeviceDaemonRecord, action DaemonControlAction) bool {
	for _, supported := range daemon.SupportedActions {
		if supported == action {
			return true
		}
	}
	return false
}

func copyDevices(devices []DeviceRecord) []DeviceRecord {
	out := make([]DeviceRecord, 0, len(devices))
	for _, device := range devices {
		out = append(out, copyDevice(device))
	}
	return out
}

func copyDevice(device DeviceRecord) DeviceRecord {
	runtimes := make([]RuntimeRecord, len(device.Runtimes))
	for i, runtime := range device.Runtimes {
		runtimes[i] = copyRuntime(runtime)
	}
	device.Runtimes = runtimes
	return device
}

func copyRuntime(runtime RuntimeRecord) RuntimeRecord {
	runtime.Models = append([]RuntimeModelRecord(nil), runtime.Models...)
	return runtime
}

func copyDeviceDaemon(daemon DeviceDaemonRecord) DeviceDaemonRecord {
	daemon.SupportedActions = append([]DaemonControlAction(nil), daemon.SupportedActions...)
	return daemon
}

func copyTaskThread(thread AIAgentTaskThreadRecord) AIAgentTaskThreadRecord {
	thread.Lines = copyProgressLines(thread.Lines)
	return thread
}

func copyProgressLines(lines []AgentThreadProgressLine) []AgentThreadProgressLine {
	return append([]AgentThreadProgressLine(nil), lines...)
}

func copyAgentTemplates(templates []AgentOnboardingTemplate) []AgentOnboardingTemplate {
	return append([]AgentOnboardingTemplate(nil), templates...)
}

func taskThreadHasActiveStream(thread AIAgentTaskThreadRecord) bool {
	switch thread.AssignmentState {
	case AgentAssignmentStateQueued, AgentAssignmentStateRunning, AgentAssignmentStateStopping:
		return true
	default:
		return false
	}
}

func activeStreamLinkForThread(thread AIAgentTaskThreadRecord) AIAgentTaskThreadStreamLink {
	return AIAgentTaskThreadStreamLink{
		Rel:       "agent_thread_progress_stream",
		Href:      "/v1/client/ai-agent/events",
		EventType: AgentClientEventThreadProgress,
		TaskID:    thread.TaskID,
		ThreadID:  thread.ThreadID,
		RunID:     thread.RunID,
	}
}

func threadIDForRun(taskID, agentID, runID string) string {
	return "thread-" + slugAIAgentIDComponent(taskID+"-"+agentID+"-"+runID)
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

func runtimeSelectionForPrincipal(devices []DeviceRecord, principal AuthorizationResult, runtimeID string, requestedModelID *string) (RuntimeKind, RuntimeModelRecord, bool) {
	for _, device := range filterDevicesForPrincipal(devices, principal) {
		for _, runtime := range device.Runtimes {
			if runtime.RuntimeID != runtimeID {
				continue
			}
			model, ok := runtimeModelSelection(runtime, requestedModelID)
			return runtime.Kind, model, ok
		}
	}
	return "", RuntimeModelRecord{}, false
}

func runtimeModelSelection(runtime RuntimeRecord, requestedModelID *string) (RuntimeModelRecord, bool) {
	modelID := ""
	if requestedModelID != nil {
		modelID = strings.TrimSpace(*requestedModelID)
	}
	if modelID != "" {
		for _, model := range runtime.Models {
			if model.ModelID == modelID {
				return model, true
			}
		}
		return RuntimeModelRecord{}, false
	}
	for _, model := range runtime.Models {
		if model.IsDefault {
			return model, true
		}
	}
	return RuntimeModelRecord{}, false
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
