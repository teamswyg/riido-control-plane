package riidoaiserver

import (
	"context"
)

type AIAgentClientStore interface {
	BootstrapAIAgentClient(ctx context.Context, principal AuthorizationResult, clientKind ClientKind) (ClientBootstrapResponse, error)
	ListAIAgentOnboardingFixtures(ctx context.Context, principal AuthorizationResult) (AgentOnboardingFixtureListResponse, error)
	CreateAIAgentFromOnboardingFixture(ctx context.Context, principal AuthorizationResult, fixtureID string, req CreateAgentConfigurationRequest) (AgentClientRecordResponse, error)
	ListAIAgentDevices(ctx context.Context, principal AuthorizationResult) (DeviceRuntimeListResponse, error)
	GetAIAgentDaemon(ctx context.Context, principal AuthorizationResult, agentID string) (DeviceDaemonDetailResponse, error)
	ControlAIAgentDaemon(ctx context.Context, principal AuthorizationResult, agentID string, action DaemonControlAction, req ControlDeviceDaemonRequest) (DeviceDaemonCommandResponse, error)
	GetAIAgentDeviceDaemon(ctx context.Context, principal AuthorizationResult, deviceID string) (DeviceDaemonDetailResponse, error)
	ControlAIAgentDeviceDaemon(ctx context.Context, principal AuthorizationResult, deviceID string, action DaemonControlAction, req ControlDeviceDaemonRequest) (DeviceDaemonCommandResponse, error)
	ListAIAgentTaskAssignableAgents(ctx context.Context, principal AuthorizationResult, taskID string) (AgentClientListResponse, error)
	ListWorkspaceAssignedAgentProfiles(ctx context.Context, principal AuthorizationResult) (AssignedAgentProfileMapResponse, error)
	ListAIAgentTaskThreads(ctx context.Context, principal AuthorizationResult, taskID string) (AIAgentTaskThreadCollectionResponse, error)
	ListAIAgentTaskThreadHistory(ctx context.Context, principal AuthorizationResult, taskID string) (AIAgentTaskThreadHistoryCollectionResponse, error)
	FindAIAgentTaskThreadByID(ctx context.Context, workspaceID, threadID string) (AIAgentTaskThreadRecord, error)
	AssignAIAgentTask(ctx context.Context, principal AuthorizationResult, taskID string, req AssignAIAgentTaskRequest) (AIAgentTaskActionResponse, error)
	CreateAIAgentTaskAgentAssignment(ctx context.Context, principal AuthorizationResult, taskID string, req AssignAIAgentTaskRequest) (AIAgentTaskActionResponse, error)
	UnassignAIAgentTask(ctx context.Context, principal AuthorizationResult, taskID string, req UnassignAIAgentTaskRequest) (AIAgentTaskActionResponse, error)
	DeleteAIAgentTaskAgentAssignment(ctx context.Context, principal AuthorizationResult, taskID, agentID string, req AgentAssignmentActionRequest) (AIAgentTaskActionResponse, error)
	SubmitAIAgentTaskComment(ctx context.Context, principal AuthorizationResult, taskID string, req SubmitAIAgentTaskCommentRequest) (AIAgentTaskActionResponse, error)
	CreateAIAgentTaskThreadMessage(ctx context.Context, principal AuthorizationResult, taskID, threadID string, req CreateAIAgentTaskThreadMessageRequest) (AIAgentTaskActionResponse, error)
	StopAIAgentTask(ctx context.Context, principal AuthorizationResult, taskID string, req StopAIAgentTaskRequest) (AIAgentTaskActionResponse, error)
	StopAIAgentTaskAgentAssignment(ctx context.Context, principal AuthorizationResult, taskID, agentID string, req AgentAssignmentActionRequest) (AIAgentTaskActionResponse, error)
	GetAIAgentTaskThreadStreamSubscription(ctx context.Context, principal AuthorizationResult, taskID string) (AIAgentTaskThreadStreamSubscriptionResponse, error)
	CreateAIAgent(ctx context.Context, principal AuthorizationResult, req CreateAgentConfigurationRequest) (AgentClientRecordResponse, error)
	GetAIAgentEditability(ctx context.Context, principal AuthorizationResult, agentID string) (AgentEditabilityResponse, error)
	UpdateAIAgentConfiguration(ctx context.Context, principal AuthorizationResult, agentID string, req UpdateAgentConfigurationRequest) (AgentClientRecordResponse, error)
	DeleteAIAgent(ctx context.Context, principal AuthorizationResult, agentID string) (DeleteAgentResponse, error)
	AIAgentClientEvents(ctx context.Context, principal AuthorizationResult) ([]ClientStreamEvent, error)
}
