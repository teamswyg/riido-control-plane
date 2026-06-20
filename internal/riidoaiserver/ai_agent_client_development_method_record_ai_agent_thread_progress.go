package riidoaiserver

import "context"

func (s *DevelopmentAIAgentClientStore) RecordAIAgentThreadProgress(ctx context.Context, agentID string, req AgentThreadProgressBatchRequest) (AgentThreadProgressBatchResponse, error) {
	if err := ctx.Err(); err != nil {
		return AgentThreadProgressBatchResponse{}, err
	}
	input, err := newThreadProgressInput(agentID, req)
	if err != nil {
		return AgentThreadProgressBatchResponse{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	agent, ok := s.agents[input.AgentID]
	if !ok {
		return AgentThreadProgressBatchResponse{}, ErrAIAgentNotFound
	}
	if existing, found := s.taskThreadForAssignmentLocked(input.Request.TaskID, input.AgentID, input.Request.AssignmentID); found {
		if !agentAssignmentStateAcceptsRuntimeProgress(existing.AssignmentState) {
			return AgentThreadProgressBatchResponse{SchemaVersion: SchemaVersion, AcceptedLines: 0}, nil
		}
		input.Request.ThreadID = existing.ThreadID
		if input.Request.RunID == "" {
			input.Request.RunID = existing.RunID
		}
		input.GeneratedThreadID = false
	}
	if active, ok := s.activeTaskThreadForAgentLocked(input.Request.TaskID, input.AgentID); ok &&
		active.ThreadID != input.Request.ThreadID &&
		input.GeneratedThreadID {
		input.Request.ThreadID = active.ThreadID
	}
	if existing, ok := s.taskThreadByIDLocked(input.Request.TaskID, input.Request.ThreadID); ok &&
		existing.AssignmentID == input.Request.AssignmentID {
		input.Lines = filterUnseenProgressLines(existing.Lines, input.Lines)
		if len(input.Lines) == 0 {
			return input.noopResponse(), nil
		}
	}
	agent.WorkStatus = AgentWorkStatusRunning
	agent.Editability = AgentEditabilityBlockedAssignedTasks
	if agent.AssignedTaskCount == 0 {
		agent.AssignedTaskCount = 1
	}
	s.agents[input.AgentID] = agent
	event := input.event()
	s.appendThreadProgressLocked(event)
	s.appendClientEventLocked(event.EventType, event)
	return AgentThreadProgressBatchResponse{
		SchemaVersion: SchemaVersion,
		AcceptedLines: len(input.Lines),
		Event:         event,
	}, nil
}
