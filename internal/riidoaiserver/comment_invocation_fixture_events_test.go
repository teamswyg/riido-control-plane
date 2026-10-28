package riidoaiserver

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func (f *commentCallFixture) complete(t *testing.T, body string) *httptest.ResponseRecorder {
	t.Helper()
	f.assignments.completion = TaskEvent{EventID: "completed", Type: EventAssignmentCompleted, State: AssignmentCompleted, TaskID: f.source.TaskID, AssignmentID: f.source.AssignmentID, AgentID: f.source.AgentID, WorkspaceID: f.evidence.WorkspaceID, CreatedBy: f.evidence.PrincipalID, Message: body, At: time.Now()}
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"assignment_id":"`+f.source.AssignmentID+`","state":"completed"}`))
	r.Header.Set("Authorization", "Bearer assignment-token")
	w := httptest.NewRecorder()
	f.s.handleAgentEvent(w, r, f.source.AgentID)
	return w
}

func (f *commentCallFixture) threads(t *testing.T) []AIAgentTaskThreadRecord {
	t.Helper()
	r, err := f.client.ListAIAgentTaskThreads(t.Context(), AuthorizationResult{PrincipalID: f.evidence.PrincipalID, WorkspaceID: f.evidence.WorkspaceID}, f.evidence.TaskID)
	if err != nil {
		t.Fatal(err)
	}
	return r.Threads
}
