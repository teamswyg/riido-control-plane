package riidoaiserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const onboardingCompletionWorkspacePath = "/v2/client/workspaces/workspace-dev-riid/ai-agent/tasks/"

func postOnboardingCompletion(t *testing.T, handler http.Handler, taskID, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, onboardingCompletionWorkspacePath+taskID+"/onboarding-completion", strings.NewReader(body))
	req.Header.Set(aiAgentTokenHeader, "user-token")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func decodeOnboardingCompletionResponse(t *testing.T, rec *httptest.ResponseRecorder) AIAgentTaskActionResponse {
	t.Helper()
	if rec.Code != http.StatusAccepted {
		t.Fatalf("onboarding completion status=%d body=%s", rec.Code, rec.Body.String())
	}
	var response AIAgentTaskActionResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("onboarding completion json: %v", err)
	}
	return response
}

func assertOnboardingCompletionNotLeasable(t *testing.T, assignmentStore *Store) {
	t.Helper()
	poll, err := assignmentStore.PollAgent(t.Context(), "agent-owned-codex", PollRequest{
		DaemonID:  "daemon-dev-macbook",
		DeviceID:  "device-dev-macbook",
		RuntimeID: "runtime-codex-dev",
	})
	if err != nil {
		t.Fatalf("PollAgent after onboarding completion: %v", err)
	}
	if poll.Action != PollNone || poll.Assignment != nil {
		t.Fatalf("onboarding completion must not create a leasable execution: %+v", poll)
	}
}

func TestHTTPOnboardingCompletionRecordsAgentCommentWithoutExecution(t *testing.T) {
	handler, assignmentStore := newIntentGateHTTPTestServer(t)

	response := decodeOnboardingCompletionResponse(t, postOnboardingCompletion(t, handler, "task-onboarding-1",
		`{"agent_id":"agent-owned-codex","template_key":"schedule"}`))

	if response.WorkStatus != AgentWorkStatusCompleted || response.AssignmentState != AgentAssignmentStateCompleted {
		t.Fatalf("completion status = %s/%s, want completed", response.WorkStatus, response.AssignmentState)
	}
	if response.CommentKind != AgentTaskCommentTaskCompleted {
		t.Fatalf("comment_kind = %s, want task_completed", response.CommentKind)
	}
	if want := onboardingCompletionCopy[AIAgentOnboardingTemplateSchedule][AIAgentOnboardingLocaleKo]; response.Message != want {
		t.Fatalf("message = %q, want approved ko schedule copy", response.Message)
	}
	if !isOnboardingSeedAssignmentID(response.AssignmentID) {
		t.Fatalf("assignment_id = %q, want onboarding-seed prefix", response.AssignmentID)
	}
	assertOnboardingCompletionNotLeasable(t, assignmentStore)

	threadsReq := httptest.NewRequest(http.MethodGet, onboardingCompletionWorkspacePath+"task-onboarding-1/threads", nil)
	threadsReq.Header.Set(aiAgentTokenHeader, "user-token")
	threadsRec := httptest.NewRecorder()
	handler.ServeHTTP(threadsRec, threadsReq)
	if threadsRec.Code != http.StatusOK {
		t.Fatalf("threads status=%d body=%s", threadsRec.Code, threadsRec.Body.String())
	}
	var threads AIAgentTaskThreadCollectionResponse
	if err := json.Unmarshal(threadsRec.Body.Bytes(), &threads); err != nil {
		t.Fatalf("threads json: %v", err)
	}
	if len(threads.Threads) != 1 {
		t.Fatalf("threads = %d, want exactly one completed onboarding thread", len(threads.Threads))
	}
	thread := threads.Threads[0]
	if thread.WorkStatus != AgentWorkStatusCompleted || thread.AssignmentState != AgentAssignmentStateCompleted {
		t.Fatalf("thread state = %s/%s, want completed; reconcile must not change it", thread.WorkStatus, thread.AssignmentState)
	}
	if thread.Message != response.Message {
		t.Fatalf("thread message = %q, want the completion copy", thread.Message)
	}
}

func TestHTTPOnboardingCompletionUsesApprovedCopyPerTemplateAndLocale(t *testing.T) {
	cases := []struct {
		template string
		locale   string
		want     string
	}{
		{"schedule", "ko", "단계 4개를 이 작업의 하위 작업으로 추가하고 마감일을 넣어 뒀어요. 참여자는 비워 뒀으니 팀원을 초대하면 바로 넣을 수 있어요."},
		{"schedule", "en", "Added the 4 phases as subtasks of this task with due dates. Assignees are empty, so you can add them as soon as you invite teammates."},
		{"interview", "ko", "질문지를 이 작업의 문서로 남겼어요. \"~라면 쓰시겠어요?\"처럼 답을 유도하는 질문은 뺐어요."},
		{"interview", "en", "Saved the questions as this task's document. Leading questions like \"Would you use it if…?\" were left out."},
		{"launch", "ko", "항목 9개를 이 작업의 하위 작업으로 추가하고 D-3, 당일, D+7 마감일을 넣어 뒀어요."},
		{"launch", "en", "Added the 9 items as subtasks of this task with D-3, launch-day, and D+7 due dates."},
	}
	for i, tc := range cases {
		handler, _ := newIntentGateHTTPTestServer(t)
		body := `{"agent_id":"agent-owned-codex","template_key":"` + tc.template + `","locale":"` + tc.locale + `"}`
		response := decodeOnboardingCompletionResponse(t, postOnboardingCompletion(t, handler, "task-copy-"+string(rune('a'+i)), body))
		if response.Message != tc.want {
			t.Fatalf("%s/%s message = %q, want %q", tc.template, tc.locale, response.Message, tc.want)
		}
	}
}

func TestHTTPOnboardingCompletionDefaultsToKoreanLocale(t *testing.T) {
	handler, _ := newIntentGateHTTPTestServer(t)
	response := decodeOnboardingCompletionResponse(t, postOnboardingCompletion(t, handler, "task-launch-default",
		`{"agent_id":"agent-owned-codex","template_key":"launch"}`))
	if want := "항목 9개를 이 작업의 하위 작업으로 추가하고 D-3, 당일, D+7 마감일을 넣어 뒀어요."; response.Message != want {
		t.Fatalf("default locale message = %q, want ko launch copy", response.Message)
	}
}

func TestHTTPOnboardingCompletionIsIdempotent(t *testing.T) {
	handler, _ := newIntentGateHTTPTestServer(t)
	first := decodeOnboardingCompletionResponse(t, postOnboardingCompletion(t, handler, "task-idem-1",
		`{"agent_id":"agent-owned-codex","template_key":"schedule","request_id":"req-onboarding-1"}`))

	replay := decodeOnboardingCompletionResponse(t, postOnboardingCompletion(t, handler, "task-idem-1",
		`{"agent_id":"agent-owned-codex","template_key":"schedule","request_id":"req-onboarding-1"}`))
	if replay.AssignmentID != first.AssignmentID || replay.ThreadID != first.ThreadID || replay.Message != first.Message {
		t.Fatalf("request_id replay changed the record: first=%+v replay=%+v", first, replay)
	}

	// Same (task, agent) without request_id also returns the existing record, even with another template.
	repeat := decodeOnboardingCompletionResponse(t, postOnboardingCompletion(t, handler, "task-idem-1",
		`{"agent_id":"agent-owned-codex","template_key":"launch"}`))
	if repeat.ThreadID != first.ThreadID || repeat.Message != first.Message {
		t.Fatalf("repeat for same task/agent duplicated the completion: %+v", repeat)
	}

	threadsReq := httptest.NewRequest(http.MethodGet, onboardingCompletionWorkspacePath+"task-idem-1/threads", nil)
	threadsReq.Header.Set(aiAgentTokenHeader, "user-token")
	threadsRec := httptest.NewRecorder()
	handler.ServeHTTP(threadsRec, threadsReq)
	var threads AIAgentTaskThreadCollectionResponse
	if err := json.Unmarshal(threadsRec.Body.Bytes(), &threads); err != nil {
		t.Fatalf("threads json: %v", err)
	}
	if len(threads.Threads) != 1 {
		t.Fatalf("idempotent completion created %d threads, want 1", len(threads.Threads))
	}
}

func TestHTTPOnboardingCompletionRejectsInvalidRequestsAndAccess(t *testing.T) {
	cases := []struct {
		name   string
		path   string
		token  string
		body   string
		status int
	}{
		{"unknown agent", onboardingCompletionWorkspacePath + "task-bad-1/onboarding-completion", "user-token", `{"agent_id":"agent-missing","template_key":"schedule"}`, http.StatusNotFound},
		{"missing agent", onboardingCompletionWorkspacePath + "task-bad-2/onboarding-completion", "user-token", `{"template_key":"schedule"}`, http.StatusBadRequest},
		{"unknown template", onboardingCompletionWorkspacePath + "task-bad-3/onboarding-completion", "user-token", `{"agent_id":"agent-owned-codex","template_key":"bogus"}`, http.StatusBadRequest},
		{"unknown locale", onboardingCompletionWorkspacePath + "task-bad-4/onboarding-completion", "user-token", `{"agent_id":"agent-owned-codex","template_key":"schedule","locale":"fr"}`, http.StatusBadRequest},
		{"unknown field", onboardingCompletionWorkspacePath + "task-bad-5/onboarding-completion", "user-token", `{"agent_id":"agent-owned-codex","template_key":"schedule","status":"done"}`, http.StatusBadRequest},
		{"wrong workspace", "/v2/client/workspaces/workspace-other/ai-agent/tasks/task-bad-6/onboarding-completion", "user-token", `{"agent_id":"agent-owned-codex","template_key":"schedule"}`, http.StatusNotFound},
		{"no token", onboardingCompletionWorkspacePath + "task-bad-7/onboarding-completion", "", `{"agent_id":"agent-owned-codex","template_key":"schedule"}`, http.StatusUnauthorized},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			handler, assignmentStore := newIntentGateHTTPTestServer(t)
			req := httptest.NewRequest(http.MethodPost, tc.path, strings.NewReader(tc.body))
			if tc.token != "" {
				req.Header.Set(aiAgentTokenHeader, tc.token)
			}
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			if rec.Code != tc.status {
				t.Fatalf("status=%d, want %d; body=%s", rec.Code, tc.status, rec.Body.String())
			}
			assertOnboardingCompletionNotLeasable(t, assignmentStore)
		})
	}
}

func TestHTTPOnboardingCompletionRemovalDoesNotStartOrCancelExecution(t *testing.T) {
	handler, assignmentStore := newIntentGateHTTPTestServer(t)
	decodeOnboardingCompletionResponse(t, postOnboardingCompletion(t, handler, "task-remove-1",
		`{"agent_id":"agent-owned-codex","template_key":"interview"}`))

	if got := onboardingParticipantIDs(t, handler, "task-remove-1"); len(got) != 1 || got[0] != "agent-owned-codex" {
		t.Fatalf("participants before removal = %v, want agent-owned-codex", got)
	}

	req := httptest.NewRequest(http.MethodDelete, onboardingCompletionWorkspacePath+"task-remove-1/agent-assignments/agent-owned-codex", nil)
	req.Header.Set(aiAgentTokenHeader, "user-token")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("remove participant status=%d body=%s; a non-durable completion must be removable", rec.Code, rec.Body.String())
	}
	if got := onboardingParticipantIDs(t, handler, "task-remove-1"); len(got) != 0 {
		t.Fatalf("participants after removal = %v, want none", got)
	}
	assertOnboardingCompletionNotLeasable(t, assignmentStore)
}

func onboardingParticipantIDs(t *testing.T, handler http.Handler, taskID string) []string {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/v2/client/workspaces/workspace-dev-riid/ai-agent/tasks/assigned-agent-profiles", nil)
	req.Header.Set(aiAgentTokenHeader, "user-token")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("assigned-agent-profiles status=%d body=%s", rec.Code, rec.Body.String())
	}
	var body AssignedAgentProfileMapResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("assigned-agent-profiles json: %v", err)
	}
	var ids []string
	for _, participant := range body.TaskAgentParticipants[taskID].Agents {
		ids = append(ids, participant.AgentID)
	}
	return ids
}

func TestOnboardingCompletionHistoryHasAgentAuthoredTaskCompletedMessage(t *testing.T) {
	ctx := context.Background()
	principal := profileColorTestPrincipal()
	store := NewDevelopmentAIAgentClientStore()
	if _, err := store.CompleteAIAgentOnboardingTask(ctx, principal, "task-history-1", CompleteAIAgentOnboardingTaskRequest{
		AgentID:     "agent-owned-codex",
		TemplateKey: AIAgentOnboardingTemplateLaunch,
		Locale:      AIAgentOnboardingLocaleEn,
	}); err != nil {
		t.Fatalf("CompleteAIAgentOnboardingTask: %v", err)
	}
	history, err := store.ListAIAgentTaskThreadHistory(ctx, principal, "task-history-1")
	if err != nil {
		t.Fatalf("ListAIAgentTaskThreadHistory: %v", err)
	}
	if len(history.Threads) != 1 {
		t.Fatalf("history threads = %d, want 1", len(history.Threads))
	}
	want := onboardingCompletionCopy[AIAgentOnboardingTemplateLaunch][AIAgentOnboardingLocaleEn]
	for _, message := range history.Threads[0].Messages {
		if message.Role == AIAgentTaskThreadMessageRoleAgent && message.CommentKind == AgentTaskCommentTaskCompleted && message.Body == want {
			return
		}
	}
	t.Fatalf("history has no agent-authored task_completed message with the en launch copy: %+v", history.Threads[0].Messages)
}

func TestNonDurableAssignmentIDPrefixes(t *testing.T) {
	for _, id := range []string{"onboarding-seed-task-1-agent-owned-codex", "intent-gate-task-1-2"} {
		if !isNonDurableAssignmentID(id) {
			t.Fatalf("%q must be non-durable", id)
		}
	}
	for _, id := range []string{"asn-dev-assignment-task-1-1", "", "onboarding-seedling-x"} {
		if isNonDurableAssignmentID(id) {
			t.Fatalf("%q must stay durable", id)
		}
	}
}
