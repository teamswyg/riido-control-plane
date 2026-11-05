package riidoaiserver

import (
	"net/http"
	"testing"
)

func TestHTTPThreadReplyMentionPolicy(t *testing.T) {
	for _, tc := range []struct {
		name, body, agent string
		status            int
	}{
		{"plain reply", "이어서 작업해봐", "agent-owned-codex", 202},
		{"same Agent", mentionHTML("agent-owned-codex") + " continue", "agent-owned-codex", 202},
		{"human mention", `<p><span data-type="mention" data-mention-kind="member" data-mention-id="agent-public-openclaw">Member</span> continue</p>`, "agent-owned-codex", 202},
		{"quoted mention", "<blockquote>" + mentionHTML("agent-public-openclaw") + "</blockquote> continue", "agent-owned-codex", 202},
		{"approval phrase for other Agent", mentionHTML("agent-public-openclaw") + " 실행해줘", "agent-public-openclaw", 202},
		{"unknown Agent", mentionHTML("missing-agent"), "", 404},
		{"private Agent", mentionHTML("agent-private-cursor"), "", 404},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newNativeResumeHTTPFixture(t)
			first := f.finishFirst(t, AssignmentCompleted)
			var next AIAgentTaskActionResponse
			f.post(t, f.base+"/tasks/task-new/threads/"+first.ThreadID+"/messages", CreateAIAgentTaskThreadMessageRequest{Body: tc.body, RequestID: "reply"}, tc.status, &next)
			if tc.status != http.StatusAccepted {
				if len(f.client.taskThreads[first.TaskID]) != 1 {
					t.Fatal("rejected mention silently continued the original Agent")
				}
				return
			}
			projection, found, err := f.store.LoadAssignmentProjection(t.Context(), next.AssignmentID)
			if err != nil || !found || projection.Assignment.AgentID != tc.agent || next.AgentID != tc.agent {
				t.Fatal("reply routed to the wrong Agent")
			}
			if tc.agent == "agent-owned-codex" && projection.Assignment.ResumeSessionID != "native-http-session" {
				t.Fatal("ordinary continuation lost its native session")
			}
		})
	}
}

func TestHTTPThreadReplyMultipleMentionsStayOneMessageOnRetry(t *testing.T) {
	f := newNativeResumeHTTPFixture(t)
	first := f.finishFirst(t, AssignmentCompleted)
	request := CreateAIAgentTaskThreadMessageRequest{RequestID: "two-agents", Body: mentionHTML("agent-public-openclaw") + mentionHTML("agent-owned-codex") + mentionHTML("agent-public-openclaw") + " review together"}
	path := f.base + "/tasks/task-new/threads/" + first.ThreadID + "/messages"
	for range 2 {
		f.post(t, path, request, http.StatusAccepted, nil)
	}
	threads := f.client.taskThreads[first.TaskID]
	if len(threads) != 3 {
		t.Fatalf("expected root and two executions, got %d", len(threads))
	}
	ids := map[string]bool{}
	agents := map[string]bool{}
	for _, thread := range threads {
		if thread.ThreadID == first.ThreadID {
			continue
		}
		agents[thread.AgentID] = true
		if thread.ConversationID != first.ThreadID || thread.ParentThreadID != first.ThreadID {
			t.Fatal("fanout split the original conversation")
		}
		for _, message := range f.client.taskThreadMessages[thread.ThreadID] {
			if message.Role == AIAgentTaskThreadMessageRoleUser {
				ids[message.MessageID] = true
			}
		}
	}
	if len(ids) != 1 || len(agents) != 2 {
		t.Fatal("fanout duplicated the human reply or selected only one Agent")
	}
}

func TestHTTPThreadReplyMentionDoesNotFoldIntoActiveOriginalAgent(t *testing.T) {
	f := newNativeResumeHTTPFixture(t)
	var first, next AIAgentTaskActionResponse
	f.post(t, f.base+"/tasks/task-new/agent-assignments", AssignAIAgentTaskRequest{AgentID: "agent-owned-codex", RequestID: "first-active"}, http.StatusAccepted, &first)
	f.post(t, f.base+"/tasks/task-new/threads/"+first.ThreadID+"/messages", CreateAIAgentTaskThreadMessageRequest{Body: mentionHTML("agent-public-openclaw") + " review", RequestID: "second-agent"}, http.StatusAccepted, &next)
	if next.AssignmentID == first.AssignmentID || next.AgentID != "agent-public-openclaw" {
		t.Fatal("explicit handoff was folded into the active original Agent")
	}
}

func TestHTTPThreadReplyValidatesEveryMentionBeforeExecution(t *testing.T) {
	f := newNativeResumeHTTPFixture(t)
	first := f.finishFirst(t, AssignmentCompleted)
	f.post(t, f.base+"/tasks/task-new/threads/"+first.ThreadID+"/messages", CreateAIAgentTaskThreadMessageRequest{Body: mentionHTML("agent-public-openclaw") + mentionHTML("missing-agent"), RequestID: "invalid-fanout"}, http.StatusNotFound, nil)
	if len(f.client.taskThreads[first.TaskID]) != 1 {
		t.Fatal("partially dispatched an invalid mention request")
	}
}
