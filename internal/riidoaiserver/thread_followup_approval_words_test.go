package riidoaiserver

import (
	"context"
	"net/http"
	"testing"
)

func TestHTTPThreadFollowupApprovalReferencesStartNewExecution(t *testing.T) {
	const instruction = "RIID-6145를 이어서 처리해줘. 근거에 맞춘 URL 단언 정합화는 허용된 범위야. 검증과 PR 리뷰 후 승인·저장도 확인해줘."
	for _, terminal := range []AssignmentState{AssignmentCompleted, AssignmentFailed} {
		for _, body := range []struct{ name, html string }{
			{"plain", "<p>" + instruction + "</p>"},
			{"mentioned", "<p>" + mentionHTML("agent-owned-codex") + instruction + "</p>"},
		} {
			t.Run(string(terminal)+"/"+body.name, func(t *testing.T) {
				f := newNativeResumeHTTPFixture(t)
				first := f.finishFirst(t, terminal)
				var next AIAgentTaskActionResponse
				f.post(t, f.base+"/tasks/task-new/threads/"+first.ThreadID+"/messages", CreateAIAgentTaskThreadMessageRequest{
					Body:      body.html,
					RequestID: "followup-referencing-approval",
				}, http.StatusAccepted, &next)
				if next.AssignmentID == first.AssignmentID || next.ThreadID == first.ThreadID {
					t.Fatalf("new work instruction was consumed as an approval reply: first=%+v next=%+v", first, next)
				}
				poll, err := f.store.PollAgent(context.Background(), "agent-owned-codex", PollRequest{DaemonID: "daemon-dev-macbook", DeviceID: "device-dev-macbook", RuntimeID: "runtime-codex-dev"})
				if err != nil || poll.Assignment == nil || poll.Assignment.ID != next.AssignmentID || poll.Assignment.ResumeSessionID != "native-http-session" {
					t.Fatal("follow-up was not dispatched with its authoritative native session", err)
				}
			})
		}
	}
}
