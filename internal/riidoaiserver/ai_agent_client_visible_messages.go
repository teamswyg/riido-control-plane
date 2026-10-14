package riidoaiserver

const (
	clientMessageAgentBusyQueued         = "지금은 다른 작업을 처리 중이에요. 현재 작업이 끝나는 대로 바로 시작할게요."
	clientMessageAgentDeleted            = "에이전트가 삭제되어 진행 중이던 작업이 중지됐어요."
	clientMessageRecoveryBlocked         = "이전 실행 세션을 확인할 수 없어 작업을 다시 시작하지 않았어요."
	clientMessageTaskCompleted           = "작업이 완료됐어요."
	clientMessageTaskFailed              = "작업을 수행하지 못했어요."
	clientMessageTaskRunning             = "작업을 진행 중이에요."
	clientMessageTaskStopped             = "작업이 중지됐어요."
	clientMessageTaskTimeout             = "에이전트 응답이 지연되어 작업을 중지했어요."
	clientMessageCloudCreditInsufficient = "보유하신 크레딧이 부족합니다. 더 깊은 리서치를 진행하려면 추가 Cloud AI 자원이 필요합니다."
	clientMessageProviderAuthFailed      = "연결된 AI 계정 인증이 만료되었거나 유효하지 않아요. 데스크탑에서 해당 AI를 다시 로그인한 뒤 다시 시도해 주세요."
	clientMessageThreadConfirmation      = "파일 생성이나 명령 실행이 필요한 작업이에요. 진행해도 괜찮다면 댓글로 알려주세요."
	clientMessageToolApprovalUnavailable = "현재 승인 대기 중인 실행이 없어 이어서 진행할 수 없어요. 다시 필요한 작업을 요청해 주세요."
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
