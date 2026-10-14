package riidoaiserver

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestDevelopmentAIAgentClientStoreHidesLocalRuntimePathsFromAssignmentActionResponse(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	principal := AuthorizationResult{PrincipalID: "user-1"}
	assigned, err := store.AssignAIAgentTask(context.Background(), principal, "task-local-path-action", AssignAIAgentTaskRequest{
		AgentID:      "agent-owned-codex",
		AssignmentID: "asn-local-path-action",
	})
	if err != nil {
		t.Fatalf("AssignAIAgentTask: %v", err)
	}
	completedEvent := TaskEvent{
		TaskID:       "task-local-path-action",
		AssignmentID: assigned.AssignmentID,
		AgentID:      assigned.AgentID,
		Type:         EventAssignmentCompleted,
		State:        AssignmentCompleted,
		Message:      "완료했습니다. 결과: [go/hello.go](</Users/teddy/riido/go/hello.go>), 로그: /tmp/riido-action/log.txt",
		At:           time.Now().UTC(),
	}
	if err := store.RecordAIAgentAssignmentEvent(context.Background(), assigned.AgentID, AgentEventRequest{}, completedEvent); err != nil {
		t.Fatalf("RecordAIAgentAssignmentEvent: %v", err)
	}
	threads, err := store.ListAIAgentTaskThreads(context.Background(), principal, "task-local-path-action")
	if err != nil {
		t.Fatalf("ListAIAgentTaskThreads: %v", err)
	}
	response := actionResponseFromThread(threads.Threads[0], "")
	for _, leaked := range []string{"/Users/", "/tmp/", "file://"} {
		if strings.Contains(response.Message, leaked) || strings.Contains(response.ResultMessage, leaked) {
			t.Fatalf("action response leaked local runtime path marker %q: %+v", leaked, response)
		}
	}
	for _, want := range []string{"go/hello.go", "로컬 파일"} {
		if !strings.Contains(response.Message, want) || !strings.Contains(response.ResultMessage, want) {
			t.Fatalf("action response = %+v, want to contain %q", response, want)
		}
	}
}
