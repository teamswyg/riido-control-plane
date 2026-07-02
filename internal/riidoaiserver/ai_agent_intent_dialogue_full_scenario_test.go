package riidoaiserver

import (
	"context"
	"testing"
)

func TestAIAgentIntentDialogueScenarioKeepsProgressAndLimitResult(t *testing.T) {
	t.Parallel()
	assertMarketingPromptAsksForIntent(t)
	ctx := context.Background()
	store := NewDevelopmentAIAgentClientStore()
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
	root := mustAssignCopywritingTask(t, store, ctx, principal)
	recordThreadProgress(t, store, root, "생각중…", "본문을 먼저 읽겠습니다.", "파악 완료.")
	recordNeedsInput(t, store, root, "@민준용 님, 작업 내용을 확인했어요. 원하는 결과물이나 방향을 댓글로 알려주세요.")
	draft, err := store.CreateAIAgentTaskThreadMessage(ctx, principal, root.TaskID, root.ThreadID, CreateAIAgentTaskThreadMessageRequest{
		AssignmentID: "asn-copy-draft",
		Body:         "우리 신기능 셀링 포인트 세 가지 반영해서 훅이 강한 카피라이팅 초안 3개만 짜줘.",
	})
	if err != nil {
		t.Fatalf("CreateAIAgentTaskThreadMessage draft: %v", err)
	}
	recordThreadProgress(t, store, draft, "생각중…", "본문을 먼저 읽겠습니다.", "파악 완료.")
	recordAssignmentCompleted(t, store, draft, "본문을 모두 읽어보았습니다.\n\nRiido 카피라이팅 3개 대안 비교\n\n### 미디어 믹스 핵심 실행 전략")
	followup, err := store.CreateAIAgentTaskThreadMessage(ctx, principal, draft.TaskID, draft.ThreadID, CreateAIAgentTaskThreadMessageRequest{
		AssignmentID: "asn-copy-research",
		Body:         "2번째 안이 좋은데 팩트 기반으로 작성하자. 자세한 리서치 후에 팩트인지 아닌지 보고해.",
	})
	if err != nil {
		t.Fatalf("CreateAIAgentTaskThreadMessage research: %v", err)
	}
	recordThreadProgress(t, store, followup, "리서치 진행 중...", "Surfing Google...")
	recordAssignmentFailed(t, store, followup, "토큰 이용 한도 초과")
	history, err := store.ListAIAgentTaskThreadHistory(ctx, principal, root.TaskID)
	if err != nil {
		t.Fatalf("ListAIAgentTaskThreadHistory: %v", err)
	}
	assertConversationFollowup(t, history, root.ThreadID, draft.ThreadID)
	assertNestedConversationFollowup(t, history, root.ThreadID, draft.ThreadID, followup.ThreadID)
	assertRootIntentQuestion(t, historyThreadByID(t, history, root.ThreadID))
	assertDraftCopywritingComplete(t, historyThreadByID(t, history, draft.ThreadID))
	assertFollowupResearchLimit(t, historyThreadByID(t, history, followup.ThreadID))
}
