package riidoaiserver

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTaskNotificationDisplacesBufferedProgress(t *testing.T) {
	subscriber := aiAgentClientSubscriber{events: make(chan ClientStreamEvent, 2)}
	subscriber.events <- ClientStreamEvent{EventType: AgentClientEventThreadProgress, Payload: bufferedProgressEvent(1)}
	subscriber.events <- ClientStreamEvent{EventType: AgentClientEventThreadProgress, Payload: bufferedProgressEvent(2)}
	subscriber = deliverAIAgentClientEvent(subscriber, ClientStreamEvent{EventType: AgentClientEventTaskNotification, Payload: AgentTaskNotificationEvent{EventID: "result-1"}})
	found := false
	for range 2 {
		e := <-subscriber.events
		if n, ok := e.Payload.(AgentTaskNotificationEvent); ok && n.EventID == "result-1" {
			found = true
		}
	}
	if !found || subscriber.terminalCompensations != 1 {
		t.Fatal("actionable notification dropped behind buffered progress")
	}
}

type retryProjectionRecorder struct {
	*handlerRecorderStore
	fail bool
}

func (s *retryProjectionRecorder) RecordAIAgentAssignmentEvent(ctx context.Context, agentID string, req AgentEventRequest, event TaskEvent) error {
	if s.fail {
		return errors.New("private snapshot diagnostic")
	}
	return s.handlerRecorderStore.RecordAIAgentAssignmentEvent(ctx, agentID, req, event)
}

func TestAgentEventProjectionFailureRequestsRetry(t *testing.T) {
	event := TaskEvent{EventID: "asn:7", Seq: 7, AssignmentID: "asn", TaskID: "task", AgentID: "agent", State: AssignmentCompleted}
	store := &handlerAssignmentStore{record: func(context.Context, string, AgentEventRequest) (AgentEventResponse, error) {
		return AgentEventResponse{SchemaVersion: SchemaVersion, Event: event}, nil
	}}
	recorder := &retryProjectionRecorder{handlerRecorderStore: &handlerRecorderStore{}, fail: true}
	server := Server{assignment: store, aiAgent: recorder, config: ServerConfig{Authorizer: assignmentHTTPAuthorizer(t, []string{"agent:agent:events:write"})}}
	request := func() *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"assignment_id":"asn","task_id":"task","state":"completed"}`))
		req.Header.Set("Authorization", "Bearer assignment-token")
		response := httptest.NewRecorder()
		server.handleAgentEvent(response, req, "agent")
		return response
	}
	response := request()
	if response.Code != http.StatusServiceUnavailable || strings.Contains(response.Body.String(), "private snapshot diagnostic") {
		t.Fatalf("projection failure acknowledged: %d %s", response.Code, response.Body.String())
	}
	recorder.fail = false
	response = request()
	if response.Code != http.StatusOK || recorder.event.EventID != event.EventID || recorder.eventCalls != 1 {
		t.Fatalf("retry failed: %d %+v", response.Code, recorder.handlerRecorderStore)
	}
}

func TestLegacyNeedsInputNotificationUsesDeclaredAssignmentState(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	store.appendTaskNotificationLocked(TaskEvent{EventID: "event", AssignmentID: "assignment", WorkspaceID: "workspace", CreatedBy: "assigner", State: AssignmentCompleted, Metadata: map[string]string{"assignment_result_status": "needs_input"}}, AIAgentTaskActionResponse{WorkStatus: AgentWorkStatusWaitingForUser, AssignmentState: AgentAssignmentState("waiting_for_user")})
	notification := store.events[len(store.events)-1].Payload.(AgentTaskNotificationEvent)
	if notification.AssignmentState != AgentAssignmentStateCompleted || notification.WorkStatus != AgentWorkStatusWaitingForUser || notification.NotificationType != "answer_required" {
		t.Fatalf("legacy schema mismatch: %+v", notification)
	}
}
