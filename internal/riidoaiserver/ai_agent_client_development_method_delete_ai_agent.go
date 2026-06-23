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
	stopped := s.markAgentTaskThreadsStoppedLocked(agent.AgentID, AgentTaskCommentStoppedByAgentDeleted, "에이전트가 삭제되어 진행 중이던 작업이 중지됐어요.")
	for _, thread := range stopped {
		s.appendAgentTaskActionEvent(actionResponseFromThread(thread, principal.WorkspaceID))
	}
	return DeleteAgentResponse{
		SchemaVersion:            SchemaVersion,
		AgentID:                  agent.AgentID,
		QueuedTasksUnassigned:    queued,
		RunningTasksForceStopped: running,
	}, nil
}
