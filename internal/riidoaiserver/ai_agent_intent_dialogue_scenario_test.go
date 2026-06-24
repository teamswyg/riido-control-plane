package riidoaiserver

import (
	"context"
	"testing"
)

func TestAIAgentIntentDialogueScenarioStaysInConversation(t *testing.T) {
	t.Parallel()
	assertMarketingPromptAsksForIntent(t)
	ctx := context.Background()
	store := NewDevelopmentAIAgentClientStore()
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
	root, err := store.AssignAIAgentTask(ctx, principal, "task-copywriting", AssignAIAgentTaskRequest{
		AgentID: "agent-owned-codex", AssignmentID: "asn-copy-root",
	})
	if err != nil {
		t.Fatalf("AssignAIAgentTask: %v", err)
	}
	recordNeedsInput(t, store, root, "@민준용 님, 어떤 작업부터 진행할까요?")
	followup, err := store.CreateAIAgentTaskThreadMessage(ctx, principal, root.TaskID, root.ThreadID, CreateAIAgentTaskThreadMessageRequest{
		AssignmentID: "asn-copy-research",
		Body:         "2번째 안이 좋은데 팩트 기반으로 작성하자. 자세한 리서치 후에 팩트인지 아닌지 보고해.",
	})
	if err != nil {
		t.Fatalf("CreateAIAgentTaskThreadMessage: %v", err)
	}
	recordProviderLimit(t, store, followup)
	history, err := store.ListAIAgentTaskThreadHistory(ctx, principal, root.TaskID)
	if err != nil {
		t.Fatalf("ListAIAgentTaskThreadHistory: %v", err)
	}
	assertConversationFollowup(t, history, root.ThreadID, followup.ThreadID)
	rootThread := historyThreadByID(t, history, root.ThreadID)
	if rootThread.WorkStatus != AgentWorkStatusWaitingForUser || !historyAgentMessageContains(rootThread, "어떤 작업부터") {
		t.Fatalf("root thread did not wait for user intent: %+v", rootThread)
	}
	followupThread := historyThreadByID(t, history, followup.ThreadID)
	if !historyMessagesContainUserBody(followupThread.Messages, "2번째 안이 좋은데 팩트 기반으로 작성하자. 자세한 리서치 후에 팩트인지 아닌지 보고해.") {
		t.Fatalf("followup user message missing: %+v", followupThread.Messages)
	}
	if !historyAgentResultContains(followupThread, clientMessageCloudCreditInsufficient) {
		t.Fatalf("provider limit result missing: %+v", followupThread.Messages)
	}
}
