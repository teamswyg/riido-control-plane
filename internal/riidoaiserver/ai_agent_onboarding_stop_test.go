package riidoaiserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// The onboarding completion is a completed record, not an execution. A stop that
// names it, or resolves to it, must be refused and must leave the record as it was.
func TestHTTPOnboardingCompletionIsNotOverwrittenByStop(t *testing.T) {
	const taskID = "task-stop-1"
	stopCases := []struct {
		name string
		path string
		body func(agentID, assignmentID string) string
	}{
		{
			name: "task stop naming the onboarding assignment",
			path: onboardingCompletionWorkspacePath + taskID + "/stop",
			body: func(agentID, assignmentID string) string {
				return fmt.Sprintf(`{"agent_id":%q,"assignment_id":%q,"reason":"user stop"}`, agentID, assignmentID)
			},
		},
		{
			name: "task stop naming only the agent",
			path: onboardingCompletionWorkspacePath + taskID + "/stop",
			body: func(agentID, _ string) string {
				return fmt.Sprintf(`{"agent_id":%q,"reason":"user stop"}`, agentID)
			},
		},
		{
			name: "agent assignment stop",
			path: "",
			body: func(_, assignmentID string) string {
				return fmt.Sprintf(`{"assignment_id":%q,"reason":"user stop"}`, assignmentID)
			},
		},
	}
	for _, tc := range stopCases {
		t.Run(tc.name, func(t *testing.T) {
			handler, _ := newOnboardingCompletionHTTPTestServer(t)
			agentID := createOnboardingAgentHTTP(t, handler, "stop agent")
			completion := decodeOnboardingCompletionResponse(t, postOnboardingCompletion(t, handler, taskID,
				onboardingCompletionBody(agentID, "schedule")))

			path := tc.path
			if path == "" {
				path = onboardingCompletionWorkspacePath + taskID + "/agent-assignments/" + agentID + "/stop"
			}
			req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(tc.body(agentID, completion.AssignmentID)))
			req.Header.Set(aiAgentTokenHeader, "user-token")
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			if rec.Code != http.StatusConflict {
				t.Fatalf("stop status=%d, want 409; body=%s", rec.Code, rec.Body.String())
			}

			threads := onboardingThreadsHTTP(t, handler, taskID)
			if len(threads) != 1 {
				t.Fatalf("threads = %d, want the single completion thread", len(threads))
			}
			thread := threads[0]
			if thread.WorkStatus != AgentWorkStatusCompleted || thread.AssignmentState != AgentAssignmentStateCompleted {
				t.Fatalf("thread state = %s/%s after refused stop, want completed", thread.WorkStatus, thread.AssignmentState)
			}
			if thread.CommentKind != AgentTaskCommentTaskCompleted || thread.Message != completion.Message {
				t.Fatalf("completion message changed after refused stop: kind=%s message=%q", thread.CommentKind, thread.Message)
			}
		})
	}
}

func onboardingThreadsHTTP(t *testing.T, handler http.Handler, taskID string) []AIAgentTaskThreadRecord {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, onboardingCompletionWorkspacePath+taskID+"/threads", nil)
	req.Header.Set(aiAgentTokenHeader, "user-token")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("threads status=%d body=%s", rec.Code, rec.Body.String())
	}
	var collection AIAgentTaskThreadCollectionResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &collection); err != nil {
		t.Fatalf("threads json: %v", err)
	}
	return collection.Threads
}
