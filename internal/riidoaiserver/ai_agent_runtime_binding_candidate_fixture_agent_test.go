package riidoaiserver

import "time"

func cursorAgentFixture(agentID string, updatedAt time.Time, status AgentWorkStatus, assigned int) AgentClientRecord {
	return AgentClientRecord{
		AgentID:           agentID,
		OwnerPrincipalID:  "user-1",
		WorkspaceID:       defaultAIAgentClientWorkspaceID,
		Name:              agentID,
		Visibility:        AgentVisibilityPrivate,
		RuntimeID:         "runtime-cursor-dev",
		RuntimeKind:       RuntimeKindCursor,
		ModelID:           "cursor-auto",
		WorkStatus:        status,
		AssignedTaskCount: assigned,
		CreatedAt:         updatedAt.Add(-time.Hour),
		UpdatedAt:         updatedAt,
	}
}

func cursorActiveThreadFixture(now time.Time) AIAgentTaskThreadRecord {
	return AIAgentTaskThreadRecord{
		ThreadID:        "thread-cursor-active",
		TaskID:          "task-cursor-active",
		AgentID:         "agent-cursor-active",
		WorkStatus:      AgentWorkStatusRunning,
		AssignmentState: AgentAssignmentStateRunning,
		CommentKind:     AgentTaskCommentRuntimeProgress,
		StartedAt:       now,
	}
}
