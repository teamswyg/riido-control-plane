package riidoaiserver

import (
	"context"
	"testing"
)

func TestAIAgentIntentDialogueCanHandoffAfterProviderLimit(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := NewDevelopmentAIAgentClientStore()
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
	root := mustAssignCopywritingTask(t, store, ctx, principal)
	recordNeedsInput(t, store, root, "@민준용 님, 작업 내용을 확인했어요. 원하는 결과물이나 방향을 댓글로 알려주세요.")
	draft := mustCreateCopywritingDraft(t, store, ctx, principal, root)
	recordAssignmentCompleted(t, store, draft, "Riido 카피라이팅 3개 대안 비교")
	research := mustCreateResearchFollowup(t, store, ctx, principal, draft)
	recordThreadProgress(t, store, research, "리서치 진행 중...", "Surfing Google...")
	recordAssignmentFailed(t, store, research, "토큰 이용 한도 초과")
	handoff, err := store.CreateAIAgentTaskAgentAssignment(ctx, principal, root.TaskID, AssignAIAgentTaskRequest{
		AgentID:      "agent-owned-claude",
		AssignmentID: "asn-copy-research-bot",
	})
	if err != nil {
		t.Fatalf("CreateAIAgentTaskAgentAssignment handoff: %v", err)
	}
	history, err := store.ListAIAgentTaskThreadHistory(ctx, principal, root.TaskID)
	if err != nil {
		t.Fatalf("ListAIAgentTaskThreadHistory: %v", err)
	}
	assertConversationFollowup(t, history, root.ThreadID, draft.ThreadID)
	assertNestedConversationFollowup(t, history, root.ThreadID, draft.ThreadID, research.ThreadID)
	assertConversationReassignment(t, history, root.ThreadID, handoff.ThreadID)
	assertFollowupResearchLimit(t, historyThreadByID(t, history, research.ThreadID))
}

func mustCreateCopywritingDraft(
	t *testing.T,
	store *DevelopmentAIAgentClientStore,
	ctx context.Context,
	principal AuthorizationResult,
	root AIAgentTaskActionResponse,
) AIAgentTaskActionResponse {
	t.Helper()
	out, err := store.CreateAIAgentTaskThreadMessage(ctx, principal, root.TaskID, root.ThreadID, CreateAIAgentTaskThreadMessageRequest{
		AssignmentID: "asn-copy-draft",
		Body:         "우리 신기능 셀링 포인트 세 가지 반영해서 훅이 강한 카피라이팅 초안 3개만 짜줘.",
	})
	if err != nil {
		t.Fatalf("CreateAIAgentTaskThreadMessage draft: %v", err)
	}
	return out
}

func mustCreateResearchFollowup(
	t *testing.T,
	store *DevelopmentAIAgentClientStore,
	ctx context.Context,
	principal AuthorizationResult,
	draft AIAgentTaskActionResponse,
) AIAgentTaskActionResponse {
	t.Helper()
	out, err := store.CreateAIAgentTaskThreadMessage(ctx, principal, draft.TaskID, draft.ThreadID, CreateAIAgentTaskThreadMessageRequest{
		AssignmentID: "asn-copy-research",
		Body:         "2번째 안이 좋은데 팩트 기반으로 작성하자. 자세한 리서치 후에 팩트인지 아닌지 보고해.",
	})
	if err != nil {
		t.Fatalf("CreateAIAgentTaskThreadMessage research: %v", err)
	}
	return out
}
