package riidoaiserver

const (
	clientMessageAgentBusyQueued = "지금은 다른 작업을 처리 중이에요. 현재 작업이 끝나는 대로 바로 시작할게요."
	clientMessageAgentDeleted    = "에이전트가 삭제되어 진행 중이던 작업이 중지됐어요."
	clientMessageRecoveryBlocked = "이전 실행 세션을 확인할 수 없어 작업을 다시 시작하지 않았어요."
	clientMessageTaskCompleted   = "작업이 완료됐어요."
	clientMessageTaskFailed      = "작업을 수행하지 못했어요."
	clientMessageTaskRunning     = "작업을 진행 중이에요."
	clientMessageTaskStopped     = "작업이 중지됐어요."
	clientMessageTaskTimeout     = "에이전트 응답이 지연되어 작업을 중지했어요."
)

func clientVisibleTaskThreadFallback(kind AgentTaskCommentKind) string {
	switch kind {
	case AgentTaskCommentTaskCompleted:
		return clientMessageTaskCompleted
	case AgentTaskCommentTaskFailed:
		return clientMessageTaskFailed
	case AgentTaskCommentStoppedByAgentDeleted:
		return clientMessageAgentDeleted
	case AgentTaskCommentStoppedByUserRequest:
		return clientMessageTaskStopped
	case AgentTaskCommentQueuedByBusyAgent:
		return clientMessageAgentBusyQueued
	case AgentTaskCommentAssignmentStarted, AgentTaskCommentRuntimeProgress:
		return clientMessageTaskRunning
	default:
		return ""
	}
}
