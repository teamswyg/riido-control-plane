package riidoaiserver

import (
	"context"
	"time"
)

func (s *DevelopmentAIAgentClientStore) DeleteAIAgent(ctx context.Context, principal AuthorizationResult, agentID string) (DeleteAgentResponse, error) {
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
	default:
	}
	s.snapshotAgentTaskThreadsLocked(agent, time.Now().UTC())
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
