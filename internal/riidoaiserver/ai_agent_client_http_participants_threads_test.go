package riidoaiserver

import (
	"net/http"
	"testing"
)

func participantThreads(t *testing.T, server http.Handler, token string) AIAgentTaskThreadCollectionResponse {
	t.Helper()
	bytes := aiAgentSmokeRequest(t, server, http.MethodGet, "/v1/client/ai-agent/tasks/task-new/threads", token, "", http.StatusOK)
	var threads AIAgentTaskThreadCollectionResponse
	aiAgentSmokeDecode(t, bytes, &threads)
	return threads
}

func assertParticipantActiveThread(t *testing.T, server http.Handler, token, threadID string) {
	t.Helper()
	threads := participantThreads(t, server, token)
	if len(threads.Threads) != 1 ||
		threads.ActiveStream == nil ||
		threads.ActiveStream.ThreadID != threadID {
		t.Fatalf("threads after assign = %+v", threads)
	}
}

func assertParticipantSourceMessage(t *testing.T, server http.Handler, token string) {
	t.Helper()
	threads := participantThreads(t, server, token)
	if len(threads.Threads) != 1 ||
		threads.Threads[0].SourceMessageID != "message-next-1" {
		t.Fatalf("threads after message = %+v", threads)
	}
}

func assertParticipantStoppedThread(t *testing.T, server http.Handler, token string) {
	t.Helper()
	threads := participantThreads(t, server, token)
	if threads.ActiveStream != nil ||
		len(threads.Threads) != 1 ||
		threads.Threads[0].CommentKind != AgentTaskCommentStoppedByUserRequest {
		t.Fatalf("threads after unassign = %+v", threads)
	}
}
