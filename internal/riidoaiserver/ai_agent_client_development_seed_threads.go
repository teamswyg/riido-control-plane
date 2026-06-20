package riidoaiserver

import "time"

func developmentSeedTaskThreads(now time.Time) map[string][]AIAgentTaskThreadRecord {
	return map[string][]AIAgentTaskThreadRecord{
		"task-1": {
			developmentCompletedTaskThread(now),
			developmentRunningTaskThread(now),
		},
	}
}

func developmentCompletedTaskThread(now time.Time) AIAgentTaskThreadRecord {
	return AIAgentTaskThreadRecord{
		ThreadID:        "thread-task-1-claude-1",
		TaskID:          "task-1",
		AgentID:         "agent-owned-claude",
		RunID:           "run-dev-completed-1",
		SourceCommentID: "comment-dev-1",
		WorkStatus:      AgentWorkStatusCompleted,
		AssignmentState: AgentAssignmentStateCompleted,
		CommentKind:     AgentTaskCommentTaskCompleted,
		Message:         "이전 AI Agent 작업이 완료됐어요.",
		StartedAt:       now.Add(-20 * time.Minute),
		CompletedAt:     now.Add(-15 * time.Minute),
		Lines: []AgentThreadProgressLine{
			{Seq: 1, Message: "팀 프로젝트 조회 완료 - 프로젝트 3건의 요약을 가져왔습니다.", ObservedAt: now.Add(-18 * time.Minute)},
		},
	}
}

func developmentRunningTaskThread(now time.Time) AIAgentTaskThreadRecord {
	return AIAgentTaskThreadRecord{
		ThreadID:        "thread-task-1-codex-2",
		TaskID:          "task-1",
		AgentID:         "agent-owned-codex",
		RunID:           "run-dev-1",
		SourceCommentID: "comment-dev-2",
		WorkStatus:      AgentWorkStatusRunning,
		AssignmentState: AgentAssignmentStateRunning,
		CommentKind:     AgentTaskCommentRuntimeProgress,
		Message:         "팀 프로젝트 수집 중 - 팀의 프로젝트 목록을 조회 중.",
		StartedAt:       now.Add(-3 * time.Minute),
		Lines: []AgentThreadProgressLine{
			{Seq: 1, Message: "생각 중...", ObservedAt: now.Add(-3 * time.Minute)},
			{Seq: 2, Message: "팀 프로젝트 수집 중 - 팀의 프로젝트 목록, 진행 상태, 우선순위와 담당자 정보를 조회해 요약을 준비 중.", ObservedAt: now.Add(-2 * time.Minute)},
		},
	}
}
