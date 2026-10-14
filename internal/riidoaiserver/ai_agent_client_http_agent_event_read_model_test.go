package riidoaiserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/teamswyg/riido-contracts/metadatakeys"
)

func TestHTTPAgentEventsUpdateAIAgentTaskThreadReadModel(t *testing.T) {
	ctx := context.Background()
	aiAgentStore := NewDevelopmentAIAgentClientStore()
	assignmentStore := NewStoreWithConfig(StoreConfig{AgentRegistry: aiAgentStore})
	defer assignmentStore.Close()
	authorizer, err := NewStaticTokenAuthorizer([]StaticTokenCredential{{
		PrincipalID: "user-1",
		Token:       "user-token",
		Scopes:      []string{"ai-agent:*", "task:task-new:read", "task:task-new:assign"},
	}, {
		PrincipalID: "daemon-shared-studio",
		Token:       "daemon-token",
		Scopes:      []string{"agent:*:events:write"},
	}})
	if err != nil {
		t.Fatalf("NewStaticTokenAuthorizer: %v", err)
	}
	handler := NewServer(ServerConfig{
		Assignment:    assignmentStore,
		AIAgentClient: aiAgentStore,
		TaskContext:   &assignmentHTTPTaskContextReader{contextSnapshot: aiAgentTaskContextHTTPFixture()},
		Authorizer:    authorizer,
	}).Handler()

	assignReq := httptest.NewRequest(http.MethodPost, "/v1/client/ai-agent/tasks/task-new/assignment", strings.NewReader(`{"agent_id":"agent-public-openclaw"}`))
	assignReq.Header.Set("Authorization", "Bearer user-token")
	assignResp := httptest.NewRecorder()
	handler.ServeHTTP(assignResp, assignReq)
	if assignResp.Code != http.StatusAccepted {
		t.Fatalf("assign status=%d body=%s", assignResp.Code, assignResp.Body.String())
	}
	var assigned AIAgentTaskActionResponse
	if err := json.Unmarshal(assignResp.Body.Bytes(), &assigned); err != nil {
		t.Fatalf("assign json: %v", err)
	}

	poll, err := assignmentStore.PollAgent(ctx, "agent-public-openclaw", PollRequest{
		DaemonID:  "daemon-shared-studio",
		DeviceID:  "device-shared-studio",
		RuntimeID: "runtime-openclaw-shared",
	})
	if err != nil {
		t.Fatalf("PollAgent: %v", err)
	}
	if poll.Assignment == nil || poll.Assignment.ID == "" {
		t.Fatalf("poll response = %+v", poll)
	}

	readyBody := `{"assignment_id":"` + poll.Assignment.ID + `","task_id":"task-new","daemon_id":"daemon-shared-studio","device_id":"device-shared-studio","runtime_id":"runtime-openclaw-shared","state":"ready","event_type":"assignment_ready","message":"runtime accepted assignment"}`
	readyReq := httptest.NewRequest(http.MethodPost, "/v1/agents/agent-public-openclaw/events", strings.NewReader(readyBody))
	readyReq.Header.Set("Authorization", "Bearer daemon-token")
	readyResp := httptest.NewRecorder()
	handler.ServeHTTP(readyResp, readyReq)
	if readyResp.Code != http.StatusOK {
		t.Fatalf("ready event status=%d body=%s", readyResp.Code, readyResp.Body.String())
	}

	readyThreadsReq := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/tasks/task-new/threads", nil)
	readyThreadsReq.Header.Set("Authorization", "Bearer user-token")
	readyThreadsResp := httptest.NewRecorder()
	handler.ServeHTTP(readyThreadsResp, readyThreadsReq)
	if readyThreadsResp.Code != http.StatusOK {
		t.Fatalf("ready threads status=%d body=%s", readyThreadsResp.Code, readyThreadsResp.Body.String())
	}
	var readyThreads AIAgentTaskThreadCollectionResponse
	if err := json.Unmarshal(readyThreadsResp.Body.Bytes(), &readyThreads); err != nil {
		t.Fatalf("ready threads json: %v", err)
	}
	if len(readyThreads.Threads) != 1 ||
		readyThreads.ActiveStream == nil ||
		readyThreads.ActiveStream.ThreadID != assigned.ThreadID ||
		readyThreads.Threads[0].WorkStatus != AgentWorkStatusRunning ||
		readyThreads.Threads[0].AssignmentState != AgentAssignmentStateRunning ||
		readyThreads.Threads[0].CommentKind != AgentTaskCommentAssignmentStarted {
		t.Fatalf("threads after ready event = %+v", readyThreads)
	}

	logBody := `{"assignment_id":"` + poll.Assignment.ID + `","task_id":"task-new","daemon_id":"daemon-shared-studio","device_id":"device-shared-studio","runtime_id":"runtime-openclaw-shared","state":"running","event_type":"riido_log","message":"팀 프로젝트 수집 중 - 진행 상태를 조회 중.","metadata":{"` + metadatakeys.ThreadProgressSeq.String() + `":"1"}}`
	logReq := httptest.NewRequest(http.MethodPost, "/v1/agents/agent-public-openclaw/events", strings.NewReader(logBody))
	logReq.Header.Set("Authorization", "Bearer daemon-token")
	logResp := httptest.NewRecorder()
	handler.ServeHTTP(logResp, logReq)
	if logResp.Code != http.StatusOK {
		t.Fatalf("log event status=%d body=%s", logResp.Code, logResp.Body.String())
	}

	threadsReq := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/tasks/task-new/threads", nil)
	threadsReq.Header.Set("Authorization", "Bearer user-token")
	threadsResp := httptest.NewRecorder()
	handler.ServeHTTP(threadsResp, threadsReq)
	if threadsResp.Code != http.StatusOK {
		t.Fatalf("threads status=%d body=%s", threadsResp.Code, threadsResp.Body.String())
	}
	var threads AIAgentTaskThreadCollectionResponse
	if err := json.Unmarshal(threadsResp.Body.Bytes(), &threads); err != nil {
		t.Fatalf("threads json: %v", err)
	}
	if len(threads.Threads) != 1 ||
		threads.ActiveStream == nil ||
		threads.ActiveStream.ThreadID != assigned.ThreadID ||
		len(threads.Threads[0].Lines) != 1 ||
		threads.Threads[0].Lines[0].Message != "팀 프로젝트 수집 중 - 진행 상태를 조회 중." {
		t.Fatalf("threads after log event = %+v", threads)
	}

	progressBody := `{"assignment_id":"` + poll.Assignment.ID + `","task_id":"task-new","daemon_id":"daemon-shared-studio","device_id":"device-shared-studio","runtime_id":"runtime-openclaw-shared","run_id":"` + poll.Assignment.ID + `","lines":[{"seq":2,"message":"파일 생성 중 - 산출물을 작성 중."}]}`
	progressReq := httptest.NewRequest(http.MethodPost, "/v1/agents/agent-public-openclaw/thread-progress", strings.NewReader(progressBody))
	progressReq.Header.Set("Authorization", "Bearer daemon-token")
	progressResp := httptest.NewRecorder()
	handler.ServeHTTP(progressResp, progressReq)
	if progressResp.Code != http.StatusAccepted {
		t.Fatalf("progress event status=%d body=%s", progressResp.Code, progressResp.Body.String())
	}
	var progress AgentThreadProgressBatchResponse
	if err := json.Unmarshal(progressResp.Body.Bytes(), &progress); err != nil {
		t.Fatalf("progress json: %v", err)
	}
	if progress.Event.ThreadID != assigned.ThreadID || progress.Event.RunID != poll.Assignment.ID {
		t.Fatalf("thread-progress should reconcile to assignment thread: assigned=%+v progress=%+v", assigned, progress.Event)
	}

	progressThreadsReq := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/tasks/task-new/threads", nil)
	progressThreadsReq.Header.Set("Authorization", "Bearer user-token")
	progressThreadsResp := httptest.NewRecorder()
	handler.ServeHTTP(progressThreadsResp, progressThreadsReq)
	if progressThreadsResp.Code != http.StatusOK {
		t.Fatalf("progress threads status=%d body=%s", progressThreadsResp.Code, progressThreadsResp.Body.String())
	}
	var progressThreads AIAgentTaskThreadCollectionResponse
	if err := json.Unmarshal(progressThreadsResp.Body.Bytes(), &progressThreads); err != nil {
		t.Fatalf("progress threads json: %v", err)
	}
	if len(progressThreads.Threads) != 1 ||
		progressThreads.ActiveStream == nil ||
		progressThreads.ActiveStream.ThreadID != assigned.ThreadID ||
		progressThreads.Threads[0].ThreadID != assigned.ThreadID ||
		progressThreads.Threads[0].RunID != poll.Assignment.ID ||
		len(progressThreads.Threads[0].Lines) != 2 {
		t.Fatalf("threads after thread-progress = %+v", progressThreads)
	}

	completedBody := `{"assignment_id":"` + poll.Assignment.ID + `","task_id":"task-new","daemon_id":"daemon-shared-studio","device_id":"device-shared-studio","runtime_id":"runtime-openclaw-shared","state":"completed","event_type":"assignment_completed","message":"작업 완료"}`
	completedReq := httptest.NewRequest(http.MethodPost, "/v1/agents/agent-public-openclaw/events", strings.NewReader(completedBody))
	completedReq.Header.Set("Authorization", "Bearer daemon-token")
	completedResp := httptest.NewRecorder()
	handler.ServeHTTP(completedResp, completedReq)
	if completedResp.Code != http.StatusOK {
		t.Fatalf("completed event status=%d body=%s", completedResp.Code, completedResp.Body.String())
	}

	completedThreadsReq := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/tasks/task-new/threads", nil)
	completedThreadsReq.Header.Set("Authorization", "Bearer user-token")
	completedThreadsResp := httptest.NewRecorder()
	handler.ServeHTTP(completedThreadsResp, completedThreadsReq)
	if completedThreadsResp.Code != http.StatusOK {
		t.Fatalf("completed threads status=%d body=%s", completedThreadsResp.Code, completedThreadsResp.Body.String())
	}
	var completedThreads AIAgentTaskThreadCollectionResponse
	if err := json.Unmarshal(completedThreadsResp.Body.Bytes(), &completedThreads); err != nil {
		t.Fatalf("completed threads json: %v", err)
	}
	if completedThreads.ActiveStream != nil ||
		len(completedThreads.Threads) != 1 ||
		completedThreads.Threads[0].ThreadID != assigned.ThreadID ||
		completedThreads.Threads[0].AssignmentState != AgentAssignmentStateCompleted ||
		completedThreads.Threads[0].CommentKind != AgentTaskCommentTaskCompleted ||
		completedThreads.Threads[0].Message != "작업 완료" ||
		completedThreads.Threads[0].CompletedAt.IsZero() ||
		len(completedThreads.Threads[0].Lines) != 2 {
		t.Fatalf("completed threads = %+v", completedThreads)
	}

	followupBody := `{"body":"README.md를 추가하고 실행 방법을 정리해 주세요.","source_message_id":"message-followup-1"}`
	followupReq := httptest.NewRequest(http.MethodPost, "/v1/client/ai-agent/tasks/task-new/threads/"+assigned.ThreadID+"/messages", strings.NewReader(followupBody))
	followupReq.Header.Set("Authorization", "Bearer user-token")
	followupResp := httptest.NewRecorder()
	handler.ServeHTTP(followupResp, followupReq)
	if followupResp.Code != http.StatusAccepted {
		t.Fatalf("followup status=%d body=%s", followupResp.Code, followupResp.Body.String())
	}
	var followup AIAgentTaskActionResponse
	if err := json.Unmarshal(followupResp.Body.Bytes(), &followup); err != nil {
		t.Fatalf("followup json: %v", err)
	}
	if followup.ThreadID == assigned.ThreadID ||
		followup.RunID == assigned.RunID ||
		followup.AssignmentID == assigned.AssignmentID ||
		followup.AgentID != assigned.AgentID ||
		followup.WorkStatus != AgentWorkStatusIdle ||
		followup.AssignmentState != "" ||
		followup.CommentKind != "" ||
		followup.Message != "" {
		t.Fatalf("followup response = %+v", followup)
	}
	followupPoll, err := assignmentStore.PollAgent(ctx, "agent-public-openclaw", PollRequest{
		DaemonID:  "daemon-shared-studio",
		DeviceID:  "device-shared-studio",
		RuntimeID: "runtime-openclaw-shared",
	})
	if err != nil {
		t.Fatalf("PollAgent followup: %v", err)
	}
	if followupPoll.Assignment == nil ||
		followupPoll.Assignment.AgentID != assigned.AgentID ||
		!strings.Contains(followupPoll.Assignment.Prompt, "README.md를 추가하고 실행 방법을 정리해 주세요.") ||
		!strings.Contains(followupPoll.Assignment.Prompt, "## Follow-up Thread Message") ||
		!strings.Contains(followupPoll.Assignment.Prompt, assigned.ThreadID) {
		t.Fatalf("followup poll assignment = %+v", followupPoll.Assignment)
	}
	followupThreadsReq := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/tasks/task-new/threads", nil)
	followupThreadsReq.Header.Set("Authorization", "Bearer user-token")
	followupThreadsResp := httptest.NewRecorder()
	handler.ServeHTTP(followupThreadsResp, followupThreadsReq)
	if followupThreadsResp.Code != http.StatusOK {
		t.Fatalf("followup threads status=%d body=%s", followupThreadsResp.Code, followupThreadsResp.Body.String())
	}
	var followupThreads AIAgentTaskThreadCollectionResponse
	if err := json.Unmarshal(followupThreadsResp.Body.Bytes(), &followupThreads); err != nil {
		t.Fatalf("followup threads json: %v", err)
	}
	if followupThreads.ActiveStream == nil ||
		followupThreads.ActiveStream.ThreadID != followup.ThreadID ||
		len(followupThreads.Threads) != 2 {
		t.Fatalf("followup threads = %+v", followupThreads)
	}
	assertThreadPreservedAfterFollowup(t, followupThreads.Threads, assigned, followup)

	eventsReq := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/events?replay=1", nil)
	eventsReq.Header.Set("Authorization", "Bearer user-token")
	eventsResp := httptest.NewRecorder()
	handler.ServeHTTP(eventsResp, eventsReq)
	if eventsResp.Code != http.StatusOK {
		t.Fatalf("events status=%d body=%s", eventsResp.Code, eventsResp.Body.String())
	}
	eventsBody := eventsResp.Body.String()
	if !strings.Contains(eventsBody, "event: agent_thread_progress\n") ||
		!strings.Contains(eventsBody, "event: agent_work_status_changed\n") ||
		strings.Contains(eventsBody, string(AgentTaskCommentQueuedByBusyAgent)) {
		t.Fatalf("events body = %q", eventsBody)
	}
}

func assertThreadPreservedAfterFollowup(t *testing.T, threads []AIAgentTaskThreadRecord, assigned, followup AIAgentTaskActionResponse) {
	t.Helper()
	var oldFound, followupFound bool
	for _, thread := range threads {
		switch thread.ThreadID {
		case assigned.ThreadID:
			oldFound = thread.AssignmentID == assigned.AssignmentID &&
				thread.AssignmentState == AgentAssignmentStateCompleted
		case followup.ThreadID:
			followupFound = thread.AssignmentID == followup.AssignmentID &&
				thread.RunID == followup.RunID &&
				thread.SourceMessageID == "message-followup-1"
		}
	}
	if !oldFound || !followupFound {
		t.Fatalf("oldFound=%v followupFound=%v threads=%+v", oldFound, followupFound, threads)
	}
}
