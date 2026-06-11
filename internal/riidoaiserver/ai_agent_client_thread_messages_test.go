package riidoaiserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func mustJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func hasAssistantBody(messages []AIAgentTaskThreadMessageRecord, want string) bool {
	for _, m := range messages {
		if m.Role == "assistant" && strings.Contains(m.Body, want) {
			return true
		}
	}
	return false
}

// TestAIAgentTaskThreadProgressLineExposesAssistantPartialKeyOnly asserts the
// wire keeps message_key (so the client can spot the live assistant body) while
// still hiding the structured-progress internals message_code/message_args.
func TestAIAgentTaskThreadProgressLineExposesAssistantPartialKeyOnly(t *testing.T) {
	line := AgentThreadProgressLine{
		Seq:         1,
		Message:     "the answer",
		MessageCode: 1101,
		MessageKey:  aiAgentClientAssistantPartialKey,
		MessageArgs: map[string]string{"label": "x"},
	}
	body, err := json.Marshal(line)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	got := string(body)
	if !strings.Contains(got, `"message_key":"assistant.partial"`) {
		t.Fatalf("message_key must reach the client: %s", got)
	}
	if strings.Contains(got, "message_code") || strings.Contains(got, "message_args") {
		t.Fatalf("structured-progress internals leaked: %s", got)
	}
}

// findThreadByID returns the thread with the given id from a collection response.
func findThreadByID(threads AIAgentTaskThreadCollectionResponse, threadID string) (AIAgentTaskThreadRecord, bool) {
	for _, thread := range threads.Threads {
		if thread.ThreadID == threadID {
			return thread, true
		}
	}
	return AIAgentTaskThreadRecord{}, false
}

// streamAssistantAnswer drives a streamed assistant answer (assistant.partial
// progress deltas) onto a fresh thread and returns its thread id.
func streamAssistantAnswer(t *testing.T, store *DevelopmentAIAgentClientStore, taskID string, lines ...AgentThreadProgressLine) string {
	t.Helper()
	progress, err := store.RecordAIAgentThreadProgress(context.Background(), "agent-owned-codex", AgentThreadProgressBatchRequest{
		AssignmentID: "asn-" + taskID,
		TaskID:       taskID,
		RunID:        "run-" + taskID,
		Lines:        lines,
	})
	if err != nil {
		t.Fatalf("RecordAIAgentThreadProgress: %v", err)
	}
	if progress.Event.ThreadID == "" {
		t.Fatal("progress produced no thread id")
	}
	return progress.Event.ThreadID
}

// TestAIAgentTaskThreadHistorySurvivesFollowUp is the regression for the reported
// bug: a completed agent answer must remain visible after a follow-up reply on
// the same thread, and the follow-up must be recorded as a user turn.
func TestAIAgentTaskThreadHistorySurvivesFollowUp(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	ctx := context.Background()
	principal := AuthorizationResult{PrincipalID: "user-1"}
	const taskID = "task-history-followup"

	threadID := streamAssistantAnswer(t, store, taskID,
		AgentThreadProgressLine{Seq: 1, Message: "Here is the complete answer.", MessageKey: aiAgentClientAssistantPartialKey},
	)

	if _, err := store.CreateAIAgentTaskThreadMessage(ctx, principal, taskID, threadID, CreateAIAgentTaskThreadMessageRequest{
		Body:            "please also add tests",
		SourceMessageID: "msg-2",
	}); err != nil {
		t.Fatalf("CreateAIAgentTaskThreadMessage: %v", err)
	}

	threads, err := store.ListAIAgentTaskThreads(ctx, principal, taskID)
	if err != nil {
		t.Fatalf("ListAIAgentTaskThreads: %v", err)
	}
	thread, ok := findThreadByID(threads, threadID)
	if !ok {
		t.Fatalf("thread %q missing from %+v", threadID, threads.Threads)
	}
	var gotAssistant, gotUser bool
	for _, m := range thread.Messages {
		if m.Role == "assistant" && strings.Contains(m.Body, "Here is the complete answer.") {
			gotAssistant = true
		}
		if m.Role == "user" && strings.Contains(m.Body, "please also add tests") {
			gotUser = true
		}
	}
	if !gotAssistant {
		t.Fatalf("prior assistant answer was lost after follow-up; messages=%+v", thread.Messages)
	}
	if !gotUser {
		t.Fatalf("user follow-up turn was not recorded; messages=%+v", thread.Messages)
	}
}

// TestHTTPThreadMessageHistoryOverWire drives the real HTTP routes
// (assign -> ready -> assistant.partial progress -> completed -> follow-up) and
// asserts the threads response JSON carries message_key and a messages[] history
// that survives the follow-up. This is the wire-level view of the fix.
func TestHTTPThreadMessageHistoryOverWire(t *testing.T) {
	ctx := context.Background()
	aiAgentStore := NewDevelopmentAIAgentClientStore()
	assignmentStore := NewStoreWithConfig(StoreConfig{AgentRegistry: aiAgentStore})
	defer assignmentStore.Close()
	authorizer, err := NewStaticTokenAuthorizer([]StaticTokenCredential{
		{PrincipalID: "user-1", Token: "user-token", Scopes: []string{"ai-agent:*", "task:task-new:read", "task:task-new:assign"}},
		{PrincipalID: "daemon-shared-studio", Token: "daemon-token", Scopes: []string{"agent:*:events:write"}},
	})
	if err != nil {
		t.Fatalf("authorizer: %v", err)
	}
	handler := NewServer(ServerConfig{
		Assignment:    assignmentStore,
		AIAgentClient: aiAgentStore,
		TaskContext:   &assignmentHTTPTaskContextReader{contextSnapshot: aiAgentTaskContextHTTPFixture()},
		Authorizer:    authorizer,
	}).Handler()

	do := func(method, path, token, body string, wantStatus int) string {
		t.Helper()
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		handler.ServeHTTP(resp, req)
		if resp.Code != wantStatus {
			t.Fatalf("%s %s -> status=%d body=%s", method, path, resp.Code, resp.Body.String())
		}
		return resp.Body.String()
	}
	getThreads := func() (AIAgentTaskThreadCollectionResponse, string) {
		body := do(http.MethodGet, "/v1/client/ai-agent/tasks/task-new/threads", "user-token", "", http.StatusOK)
		var out AIAgentTaskThreadCollectionResponse
		if err := json.Unmarshal([]byte(body), &out); err != nil {
			t.Fatalf("threads json: %v", err)
		}
		return out, body
	}

	assignBody := do(http.MethodPost, "/v1/client/ai-agent/tasks/task-new/assignment", "user-token", `{"agent_id":"agent-public-openclaw"}`, http.StatusAccepted)
	var assigned AIAgentTaskActionResponse
	if err := json.Unmarshal([]byte(assignBody), &assigned); err != nil {
		t.Fatalf("assign json: %v", err)
	}
	poll, err := assignmentStore.PollAgent(ctx, "agent-public-openclaw", PollRequest{DaemonID: "daemon-shared-studio", DeviceID: "device-shared-studio", RuntimeID: "runtime-openclaw-shared"})
	if err != nil || poll.Assignment == nil {
		t.Fatalf("PollAgent: %v %+v", err, poll)
	}
	aid := poll.Assignment.ID
	base := `"assignment_id":"` + aid + `","task_id":"task-new","daemon_id":"daemon-shared-studio","device_id":"device-shared-studio","runtime_id":"runtime-openclaw-shared"`

	do(http.MethodPost, "/v1/agents/agent-public-openclaw/events", "daemon-token", `{`+base+`,"state":"ready","event_type":"assignment_ready","message":"runtime accepted assignment"}`, http.StatusOK)
	do(http.MethodPost, "/v1/agents/agent-public-openclaw/thread-progress", "daemon-token", `{`+base+`,"run_id":"`+aid+`","lines":[{"seq":1,"message":"Here is the full answer.","message_key":"assistant.partial"}]}`, http.StatusAccepted)
	do(http.MethodPost, "/v1/agents/agent-public-openclaw/events", "daemon-token", `{`+base+`,"state":"completed","event_type":"assignment_completed","message":"작업 완료"}`, http.StatusOK)

	completed, completedBody := getThreads()
	if !strings.Contains(completedBody, `"message_key":"assistant.partial"`) {
		t.Fatalf("message_key must reach the client over the wire: %s", completedBody)
	}
	thread, ok := findThreadByID(completed, assigned.ThreadID)
	if !ok {
		t.Fatalf("thread %q missing", assigned.ThreadID)
	}
	t.Logf("after completed: thread.message(status)=%q messages=%s", thread.Message, mustJSON(thread.Messages))
	if !hasAssistantBody(thread.Messages, "Here is the full answer.") {
		t.Fatalf("assistant answer was not archived on completion: %+v", thread.Messages)
	}

	do(http.MethodPost, "/v1/client/ai-agent/tasks/task-new/threads/"+assigned.ThreadID+"/messages", "user-token", `{"body":"please also add tests","source_message_id":"message-followup-1"}`, http.StatusAccepted)

	after, afterBody := getThreads()
	thread2, _ := findThreadByID(after, assigned.ThreadID)
	t.Logf("after follow-up: thread.message(status)=%q messages=%s", thread2.Message, mustJSON(thread2.Messages))
	if !hasAssistantBody(thread2.Messages, "Here is the full answer.") {
		t.Fatalf("prior assistant answer LOST after follow-up: %s", afterBody)
	}
	var gotUser bool
	for _, m := range thread2.Messages {
		if m.Role == "user" && strings.Contains(m.Body, "please also add tests") {
			gotUser = true
		}
	}
	if !gotUser {
		t.Fatalf("user follow-up turn missing: %+v", thread2.Messages)
	}

	// Cross-run bleed regression: a SECOND run on the same thread must archive only
	// its own answer, not concatenated with the first run's answer (the progress
	// Lines accumulate per thread; the follow-up clears them so each run's answer
	// is isolated).
	poll2, err := assignmentStore.PollAgent(ctx, "agent-public-openclaw", PollRequest{DaemonID: "daemon-shared-studio", DeviceID: "device-shared-studio", RuntimeID: "runtime-openclaw-shared"})
	if err != nil || poll2.Assignment == nil {
		t.Fatalf("PollAgent run2: %v %+v", err, poll2)
	}
	aid2 := poll2.Assignment.ID
	base2 := `"assignment_id":"` + aid2 + `","task_id":"task-new","daemon_id":"daemon-shared-studio","device_id":"device-shared-studio","runtime_id":"runtime-openclaw-shared"`
	do(http.MethodPost, "/v1/agents/agent-public-openclaw/thread-progress", "daemon-token", `{`+base2+`,"run_id":"`+aid2+`","lines":[{"seq":1,"message":"Second run distinct answer.","message_key":"assistant.partial"}]}`, http.StatusAccepted)
	do(http.MethodPost, "/v1/agents/agent-public-openclaw/events", "daemon-token", `{`+base2+`,"state":"completed","event_type":"assignment_completed","message":"작업 완료"}`, http.StatusOK)

	final, finalBody := getThreads()
	finalThread, _ := findThreadByID(final, assigned.ThreadID)
	t.Logf("after run2: messages=%s", mustJSON(finalThread.Messages))
	if !hasAssistantBody(finalThread.Messages, "Here is the full answer.") {
		t.Fatalf("run1 answer lost after run2: %s", finalBody)
	}
	var run2OK bool
	for _, m := range finalThread.Messages {
		if m.Role == "assistant" && strings.Contains(m.Body, "Second run distinct answer.") {
			run2OK = true
			if strings.Contains(m.Body, "Here is the full answer.") {
				t.Fatalf("cross-run bleed: run2 message embeds run1 answer: %q", m.Body)
			}
		}
	}
	if !run2OK {
		t.Fatalf("run2 answer not archived as its own message: %s", finalBody)
	}
}

// TestAIAgentTaskThreadHistoryIdempotentAndDurable asserts a duplicate follow-up
// does not double-append, and that history round-trips through snapshot/restore.
func TestAIAgentTaskThreadHistoryIdempotentAndDurable(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	ctx := context.Background()
	principal := AuthorizationResult{PrincipalID: "user-1"}
	const taskID = "task-history-idem"

	threadID := streamAssistantAnswer(t, store, taskID,
		AgentThreadProgressLine{Seq: 1, Message: "answer body", MessageKey: aiAgentClientAssistantPartialKey},
	)

	// Same follow-up source id twice → user turn appended once; prior assistant
	// answer archived once (idempotent per run).
	for i := 0; i < 2; i++ {
		if _, err := store.CreateAIAgentTaskThreadMessage(ctx, principal, taskID, threadID, CreateAIAgentTaskThreadMessageRequest{
			Body:            "same follow-up",
			SourceMessageID: "msg-2",
		}); err != nil {
			t.Fatalf("CreateAIAgentTaskThreadMessage[%d]: %v", i, err)
		}
	}
	assertHistory := func(label string, s *DevelopmentAIAgentClientStore) {
		threads, err := s.ListAIAgentTaskThreads(ctx, principal, taskID)
		if err != nil {
			t.Fatalf("%s ListAIAgentTaskThreads: %v", label, err)
		}
		thread, ok := findThreadByID(threads, threadID)
		if !ok {
			t.Fatalf("%s thread %q missing", label, threadID)
		}
		var assistants, users int
		for _, m := range thread.Messages {
			switch m.Role {
			case "assistant":
				assistants++
			case "user":
				users++
			}
		}
		if assistants != 1 || users != 1 {
			t.Fatalf("%s want exactly 1 assistant + 1 user message, got assistants=%d users=%d (%+v)", label, assistants, users, thread.Messages)
		}
	}
	assertHistory("live", store)

	// Durability: snapshot -> restore into a fresh store preserves history.
	snap, err := store.snapshot(time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	restored := NewDevelopmentAIAgentClientStore()
	if err := restored.restoreSnapshot(snap); err != nil {
		t.Fatalf("restoreSnapshot: %v", err)
	}
	assertHistory("restored", restored)
}
