package riidoaiserver

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestAIAgentTaskThreadHistoryRecordsRequesterPerRun(t *testing.T) {
	ctx := context.Background()
	store := NewDevelopmentAIAgentClientStore()
	assigner := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
	// 개발 저장소에서 같은 에이전트에 접근 가능한 사용자는 하나라서 후속 실행도 같은 사용자로 만든다.
	// 다른 사용자가 이어서 요청해도 첫 요청자를 덮어쓰지 않는지는 아래 단위 테스트가 검증한다.
	replier := assigner
	assigned, err := store.AssignAIAgentTask(ctx, assigner, "task-requester", AssignAIAgentTaskRequest{AgentID: "agent-owned-codex"})
	if err != nil {
		t.Fatalf("AssignAIAgentTask: %v", err)
	}
	followup, err := store.CreateAIAgentTaskThreadMessage(ctx, replier, "task-requester", assigned.ThreadID, CreateAIAgentTaskThreadMessageRequest{
		Body:            `<p>이어서 해줘</p>`,
		SourceMessageID: "source-message-requester",
	})
	if err != nil {
		t.Fatalf("CreateAIAgentTaskThreadMessage: %v", err)
	}
	history, err := store.ListAIAgentTaskThreadHistory(ctx, assigner, "task-requester")
	if err != nil {
		t.Fatalf("ListAIAgentTaskThreadHistory: %v", err)
	}
	if got := historyThreadByID(t, history, assigned.ThreadID).RequestedBy; got != assigner.PrincipalID {
		t.Fatalf("assigned run requested_by = %q, want %q", got, assigner.PrincipalID)
	}
	if got := historyThreadByID(t, history, followup.ThreadID).RequestedBy; got != replier.PrincipalID {
		t.Fatalf("follow-up run requested_by = %q, want %q", got, replier.PrincipalID)
	}
}

func TestAIAgentTaskThreadRequesterIsNotOverwritten(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	store.taskThreads["task-keep"] = []AIAgentTaskThreadRecord{{ThreadID: "thread-keep", TaskID: "task-keep", RequestedBy: "user-first"}}
	store.recordTaskThreadRequesterLocked("task-keep", "thread-keep", "user-second")
	if got := store.taskThreads["task-keep"][0].RequestedBy; got != "user-first" {
		t.Fatalf("requested_by = %q, want user-first", got)
	}
	store.recordTaskThreadRequesterLocked("task-keep", "thread-missing", "user-second")
	store.recordTaskThreadRequesterLocked("task-keep", "thread-keep", " ")
}

func TestAIAgentTaskThreadHistoryOmitsEmptyRequester(t *testing.T) {
	body, err := json.Marshal(AIAgentTaskThreadHistoryRecord{ThreadID: "thread-legacy"})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if strings.Contains(string(body), "requested_by") {
		t.Fatalf("legacy thread should omit requested_by: %s", body)
	}
}

func TestAIAgentTaskThreadHistoryProjectsConfirmationReason(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	record := store.taskThreadHistoryRecordLocked(AIAgentTaskThreadRecord{
		ThreadID:           "thread-approval",
		TaskID:             "task-approval",
		AssignmentState:    AgentAssignmentStateWaiting,
		ConfirmationReason: "approval_required",
	}, "")
	if record.ConfirmationReason != "approval_required" {
		t.Fatalf("confirmation_reason = %q, want approval_required", record.ConfirmationReason)
	}
}
