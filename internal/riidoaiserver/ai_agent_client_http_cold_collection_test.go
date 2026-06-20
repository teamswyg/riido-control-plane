package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPAIAgentClientDevelopmentTaskThreadColdCollectionAfterViewerAwayAssignment(t *testing.T) {
	server := newAIAgentClientHTTPTestServer(t, []StaticTokenCredential{{
		PrincipalID: "user-1",
		Token:       "user-token",
		Scopes:      []string{"ai-agent:*", "task:task-viewer-away:read", "task:task-viewer-away:comment"},
	}})

	commentReq := httptest.NewRequest(http.MethodPost, "/v1/client/ai-agent/tasks/task-viewer-away/comments", strings.NewReader(`{"agent_id":"agent-public-openclaw","body":"Start while the viewer is looking elsewhere","source_comment_id":"comment-viewer-away"}`))
	commentReq.Header.Set("Authorization", "Bearer user-token")
	commentResp := httptest.NewRecorder()
	server.ServeHTTP(commentResp, commentReq)
	if commentResp.Code != http.StatusAccepted {
		t.Fatalf("comment status=%d body=%s", commentResp.Code, commentResp.Body.String())
	}
	var comment AIAgentTaskActionResponse
	if err := json.Unmarshal(commentResp.Body.Bytes(), &comment); err != nil {
		t.Fatalf("comment json: %v", err)
	}
	if comment.ThreadID == "" || comment.AssignmentState != AgentAssignmentStateRunning {
		t.Fatalf("comment response = %+v", comment)
	}

	threadsReq := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/tasks/task-viewer-away/threads", nil)
	threadsReq.Header.Set("Authorization", "Bearer user-token")
	threadsResp := httptest.NewRecorder()
	server.ServeHTTP(threadsResp, threadsReq)
	if threadsResp.Code != http.StatusOK {
		t.Fatalf("threads status=%d body=%s", threadsResp.Code, threadsResp.Body.String())
	}
	var threads AIAgentTaskThreadCollectionResponse
	if err := json.Unmarshal(threadsResp.Body.Bytes(), &threads); err != nil {
		t.Fatalf("threads json: %v", err)
	}
	if len(threads.Threads) != 1 || threads.Threads[0].ThreadID != comment.ThreadID {
		t.Fatalf("threads did not return persisted viewer-away thread: %+v", threads)
	}
	if threads.Threads[0].SourceCommentID != "comment-viewer-away" {
		t.Fatalf("source comment id = %q", threads.Threads[0].SourceCommentID)
	}
	if threads.ActiveStream == nil || threads.ActiveStream.ThreadID != comment.ThreadID || threads.ActiveStream.TaskID != "task-viewer-away" {
		t.Fatalf("active stream = %+v", threads.ActiveStream)
	}
}
