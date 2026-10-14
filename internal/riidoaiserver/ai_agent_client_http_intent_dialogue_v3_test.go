package riidoaiserver

import "testing"

func TestHTTPAIAgentIntentDialogueV3KeepsScenarioTimeline(t *testing.T) {
	handler, aiStore, assignmentStore := newIntentDialogueHTTPTestServer(t)
	root := postIntentDialogueAssignment(t, handler, "agent-owned-codex")
	assertIntentGateAction(t, root)
	assertIntentDialoguePollNone(t, assignmentStore, root)

	assertHTTPRootIntentQuestion(t, getIntentDialogueV3Thread(t, handler, root, root.ThreadID))

	draft := postIntentDialogueThreadMessage(t, handler, root, "우리 신기능 셀링 포인트 세 가지 반영해서 훅이 강한 카피라이팅 초안 3개만 짜줘.")
	assertIntentDialoguePollWork(t, assignmentStore, draft)
	recordIntentDialogueRunning(t, assignmentStore, aiStore, draft)
	recordThreadProgress(t, aiStore, draft, "생각중…", "본문을 먼저 읽겠습니다.", "파악 완료.")
	recordIntentDialogueCompleted(t, assignmentStore, aiStore, draft, "본문을 모두 읽어보았습니다.\n\nRiido 카피라이팅 3개 대안 비교")

	research := postIntentDialogueThreadMessage(t, handler, draft, "2번째 안이 좋은데 팩트 기반으로 작성하자. 자세한 리서치 후에 팩트인지 아닌지 보고해.")
	assertIntentDialoguePollWork(t, assignmentStore, research)
	recordIntentDialogueRunning(t, assignmentStore, aiStore, research)
	recordThreadProgress(t, aiStore, research, "리서치 진행 중...", "Surfing Google...")
	recordIntentDialogueFailed(t, assignmentStore, aiStore, research, "토큰 이용 한도 초과")

	handoff := postIntentDialogueAssignment(t, handler, "agent-public-openclaw")
	history := getIntentDialogueV3History(t, handler, root.TaskID)
	assertConversationFollowup(t, history, root.ThreadID, draft.ThreadID)
	assertNestedConversationFollowup(t, history, root.ThreadID, draft.ThreadID, research.ThreadID)
	assertConversationReassignment(t, history, root.ThreadID, handoff.ThreadID)
	assertPostLimitHandoffKeepsPriorConversation(t, history, research.ThreadID, handoff.ThreadID)
	assertHTTPDraftReadsBodyBeforeCompletion(t, historyThreadByID(t, history, draft.ThreadID))
	assertDraftCopywritingComplete(t, historyThreadByID(t, history, draft.ThreadID))
	assertFollowupResearchLimit(t, historyThreadByID(t, history, research.ThreadID))
}
