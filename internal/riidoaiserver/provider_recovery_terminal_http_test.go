package riidoaiserver

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
	"time"
)

func TestRecoveryJournalThenTerminalHTTPPreservesTaskCompletion(t *testing.T) {
	for _, terminal := range []AssignmentState{AssignmentCompleted, AssignmentFailed} {
		t.Run(string(terminal), func(t *testing.T) {
			recorder := &recordingTraceRecorder{}
			ctx := WithTraceRecorder(context.Background(), recorder)
			client := NewDevelopmentAIAgentClientStore()
			binding, ok := client.LookupDaemonBinding("agent-owned-codex")
			if !ok {
				t.Fatal("fixture binding missing")
			}
			current := Assignment{ID: "terminal-asn", TaskID: "task-recovery-terminal", AgentID: binding.AgentID, RuntimeProvider: "codex", State: AssignmentRunning, LeaseToken: "private-lease", UpdatedAt: time.Now()}
			ops := &conditionalTestOperations{projection: AssignmentProjection{Assignment: current, LastEventSeq: 10}}
			store := NewStoreWithConfig(StoreConfig{OperationStore: ops, AgentRegistry: client})
			defer store.Close()
			principal := AuthorizationResult{PrincipalID: "user-1"}
			if _, err := client.AssignAIAgentTask(ctx, principal, current.TaskID, AssignAIAgentTaskRequest{AgentID: current.AgentID, AssignmentID: current.ID, durableState: AssignmentRunning}); err != nil {
				t.Fatal(err)
			}
			server := Server{assignment: store, aiAgent: client, config: ServerConfig{Authorizer: assignmentHTTPAuthorizer(t, []string{"agent:" + current.AgentID + ":events:write"})}}
			recovery := recoveryEventRequest(current)
			recovery.DaemonID = binding.DaemonID
			recovery.DeviceID = binding.DeviceID
			recovery.RuntimeID = binding.RuntimeID
			terminalEvent := AgentEventRequest{AssignmentID: current.ID, TaskID: current.TaskID, DaemonID: binding.DaemonID, DeviceID: binding.DeviceID, RuntimeID: binding.RuntimeID, State: terminal, EventType: EventAssignmentCompleted, Message: "result"}
			if terminal == AssignmentFailed {
				terminalEvent.EventType = EventAssignmentFailed
				terminalEvent.Metadata = map[string]string{"assignment_result_status": "failed", "assignment_failure_category": "provider_result_failed"}
			}
			for _, event := range []AgentEventRequest{recovery, terminalEvent} {
				body, err := json.Marshal(event)
				if err != nil {
					t.Fatal(err)
				}
				req := httptest.NewRequest(http.MethodPost, "/v1/agents/"+current.AgentID+"/events", bytes.NewReader(body))
				req.Header.Set("Authorization", "Bearer assignment-token")
				response := httptest.NewRecorder()
				server.handleAgentEvent(response, req.WithContext(ctx), current.AgentID)
				if response.Code != http.StatusOK {
					t.Fatalf("event %s rejected: status=%d body=%s", event.EventType, response.Code, response.Body.String())
				}
			}
			threads, err := client.ListAIAgentTaskThreads(ctx, principal, current.TaskID)
			if err != nil || len(threads.Threads) != 1 {
				t.Fatalf("cold history: %+v %v", threads, err)
			}
			if got := string(threads.Threads[0].AssignmentState); got != string(terminal) {
				t.Fatalf("terminal history=%s want=%s", got, terminal)
			}
			if ops.projection.Assignment.State != terminal || len(ops.records) != 2 {
				t.Fatal("terminal execution did not commit after recovery journal")
			}
			if got := slices.Contains(traceSpanNames(recorder.snapshot()), "assignment failure committed"); got != (terminal == AssignmentFailed) {
				t.Fatal("HTTP 200 terminal acknowledgment lost committed failure trace evidence")
			}
		})
	}
}
