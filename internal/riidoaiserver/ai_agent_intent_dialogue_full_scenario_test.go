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
	recordNeedsInput(t, store, root, "@민준용 님, 어떤 작업부터 진행할까요?")
	followup, err := store.CreateAIAgentTaskThreadMessage(ctx, principal, root.TaskID, root.ThreadID, CreateAIAgentTaskThreadMessageRequest{
		AssignmentID: "asn-copy-research",
		Body:         "2번째 안이 좋은데 팩트 기반으로 작성하자. 자세한 리서치 후에 팩트인지 아닌지 보고해.",
	})
	if err != nil {
		t.Fatalf("CreateAIAgentTaskThreadMessage: %v", err)
	}
	recordThreadProgress(t, store, followup, "리서치 진행 중...", "Surfing Google...")
	recordAssignmentFailed(t, store, followup, "토큰 이용 한도 초과")
	history, err := store.ListAIAgentTaskThreadHistory(ctx, principal, root.TaskID)
	if err != nil {
		t.Fatalf("ListAIAgentTaskThreadHistory: %v", err)
	}
	assertConversationFollowup(t, history, root.ThreadID, followup.ThreadID)
	assertRootIntentQuestion(t, historyThreadByID(t, history, root.ThreadID))
	assertFollowupResearchLimit(t, historyThreadByID(t, history, followup.ThreadID))
}
