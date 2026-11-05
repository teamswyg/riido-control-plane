package riidoaiserver

import (
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestHTTPResumedFollowupWithoutNewContextIsOnlyUserInput(t *testing.T) {
	f := newNativeResumeHTTPFixture(t)
	first := f.finishFirst(t, AssignmentCompleted)
	const instruction = "codex 세션 id 알려줘"
	var next AIAgentTaskActionResponse
	f.post(t, f.base+"/tasks/task-new/threads/"+first.ThreadID+"/messages", CreateAIAgentTaskThreadMessageRequest{Body: instruction}, http.StatusAccepted, &next)
	p, found, err := f.store.LoadAssignmentProjection(t.Context(), next.AssignmentID)
	if err != nil || !found || p.Assignment.Prompt != instruction {
		t.Fatalf("followup repeated boilerplate: %q err=%v", p.Assignment.Prompt, err)
	}
}

func TestResumedCommentPromptDoesNotRepeatDirectory(t *testing.T) {
	f := newNativeResumeHTTPFixture(t)
	s := NewServer(ServerConfig{AIAgentClient: f.client})
	req := AssignRequest{AgentID: "agent-owned-codex", ResumeSessionID: "native-http-session", Prompt: "codex 세션 id 알려줘"}
	err := s.appendCommentInvocationPrompt(t.Context(), AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}, "task-new", &req)
	if err != nil || req.Prompt != "codex 세션 id 알려줘" {
		t.Fatalf("resumed request repeated directory and delegation instructions: %q err=%v", req.Prompt, err)
	}
}

func TestHTTPResumedFollowupDoesNotRepeatPreviouslyForwardedPeerResult(t *testing.T) {
	f := newNativeResumeHTTPFixture(t)
	first := f.finishFirst(t, AssignmentCompleted)
	f.client.mu.Lock()
	f.client.taskThreads["task-new"] = append(f.client.taskThreads["task-new"], AIAgentTaskThreadRecord{ThreadID: "peer", ParentThreadID: first.ThreadID, ConversationID: first.ThreadID, TaskID: "task-new", AgentID: "other-agent", WorkStatus: AgentWorkStatusCompleted, AssignmentState: AgentAssignmentStateCompleted, StartedAt: time.Now().UTC()})
	f.client.taskThreadMessages["peer"] = []AIAgentTaskThreadHistoryMessage{{MessageID: "peer-result", Role: AIAgentTaskThreadMessageRoleAgent, Body: "Previously forwarded peer result", ObservedAt: time.Now().UTC()}}
	f.client.mu.Unlock()
	var previous AIAgentTaskActionResponse
	f.post(t, f.base+"/tasks/task-new/threads/"+first.ThreadID+"/messages", CreateAIAgentTaskThreadMessageRequest{Body: "Review the peer result", RequestID: "first-followup"}, http.StatusAccepted, &previous)
	p, _, err := f.store.LoadAssignmentProjection(t.Context(), previous.AssignmentID)
	if err != nil || !strings.Contains(p.Assignment.Prompt, "Previously forwarded peer result") {
		t.Fatal("new peer result was not delivered", err)
	}
	if _, err = f.store.PollAgent(t.Context(), previous.AgentID, PollRequest{DaemonID: "daemon-dev-macbook", DeviceID: "device-dev-macbook", RuntimeID: "runtime-codex-dev"}); err != nil {
		t.Fatal(err)
	}
	for _, phase := range []struct {
		state AssignmentState
		kind  string
	}{{AssignmentReady, EventAssignmentReady}, {AssignmentRunning, EventProviderSessionPinned}, {AssignmentCompleted, EventAssignmentCompleted}} {
		f.post(t, "/v1/agents/"+previous.AgentID+"/events", AgentEventRequest{AssignmentID: previous.AssignmentID, TaskID: "task-new", DaemonID: "daemon-dev-macbook", DeviceID: "device-dev-macbook", RuntimeID: "runtime-codex-dev", State: phase.state, EventType: phase.kind, ProviderSessionID: "native-http-session"}, http.StatusOK, nil)
	}
	var next AIAgentTaskActionResponse
	f.post(t, f.base+"/tasks/task-new/threads/"+previous.ThreadID+"/messages", CreateAIAgentTaskThreadMessageRequest{Body: "codex 세션 id 알려줘", RequestID: "second-followup"}, http.StatusAccepted, &next)
	p, _, err = f.store.LoadAssignmentProjection(t.Context(), next.AssignmentID)
	if err != nil || strings.Contains(p.Assignment.Prompt, "Previously forwarded peer result") {
		t.Fatalf("old peer result was forwarded again: %q err=%v", p.Assignment.Prompt, err)
	}
}
