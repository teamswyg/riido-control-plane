package riidoaiserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
)

func TestTaskNotificationTransitionsAndRecipient(t *testing.T) {
	ctx := context.Background()
	s := NewDevelopmentAIAgentClientStore()
	s.agents["agent"] = AgentClientRecord{AgentID: "agent", OwnerPrincipalID: "owner", WorkspaceID: "ws", Visibility: AgentVisibilityPublic}
	s.taskThreads["task"] = []AIAgentTaskThreadRecord{{ThreadID: "thread", TaskID: "task", AgentID: "agent", AssignmentID: "assignment", RunID: "run", WorkStatus: AgentWorkStatusRunning, AssignmentState: AgentAssignmentStateRunning}}
	event := TaskEvent{EventID: "event-1", Seq: 1, TaskID: "task", AssignmentID: "assignment", AgentID: "agent", WorkspaceID: "ws", CreatedBy: "assigner", State: AssignmentRunning, At: time.Now(), Type: assignmentcontract.EventAssignmentWaitingForUser, ConfirmationReason: assignmentcontract.ConfirmationApprovalRequired}
	record := func(e TaskEvent) {
		t.Helper()
		if err := s.RecordAIAgentAssignmentEvent(ctx, "agent", AgentEventRequest{}, e); err != nil {
			t.Fatal(err)
		}
	}
	notifications := func(user, workspace string) []AgentTaskNotificationEvent {
		t.Helper()
		events, err := s.AIAgentClientEvents(ctx, AuthorizationResult{PrincipalID: user, WorkspaceID: workspace})
		if err != nil {
			t.Fatal(err)
		}
		var out []AgentTaskNotificationEvent
		for _, e := range events {
			if n, ok := e.Payload.(AgentTaskNotificationEvent); ok {
				out = append(out, n)
			}
		}
		return out
	}
	record(event)
	record(event)
	if got := notifications("assigner", "ws"); len(got) != 1 || got[0].NotificationType != "approval_required" {
		t.Fatalf("approval: %+v", got)
	}
	for _, scope := range [][2]string{{"owner", "ws"}, {"other", "ws"}, {"assigner", "other"}} {
		if len(notifications(scope[0], scope[1])) != 0 {
			t.Fatal("notification leaked")
		}
	}
	event.Type = EventRiidoLog
	event.Message = "still waiting"
	record(event)
	if s.taskThreads["task"][0].WorkStatus != AgentWorkStatusWaitingForUser {
		t.Fatal("log cleared waiting state")
	}
	event.Type = EventProviderError
	record(event)
	if len(notifications("assigner", "ws")) != 1 {
		t.Fatal("recoverable error notified")
	}
	event.Type = EventAssignmentRunning
	event.ConfirmationReason = ""
	record(event)
	if s.taskThreads["task"][0].ConfirmationReason != "" {
		t.Fatal("resume retained reason")
	}
	event.Type = assignmentcontract.EventAssignmentWaitingForUser
	event.EventID = "event-2"
	event.ConfirmationReason = assignmentcontract.ConfirmationAnswerRequired
	record(event)
	if got := notifications("assigner", "ws"); len(got) != 2 || got[1].NotificationType != "answer_required" {
		t.Fatalf("answer: %+v", got)
	}
	event.Type = EventAssignmentCompleted
	event.EventID = "event-3"
	event.State = AssignmentCompleted
	event.ConfirmationReason = ""
	record(event)
	record(event)
	if got := notifications("assigner", "ws"); len(got) != 3 || got[2].NotificationType != "result_ready" {
		t.Fatalf("completion: %+v", got)
	}
	event.State = AssignmentRunning
	event.Type = EventAssignmentRunning
	record(event)
	if s.taskThreads["task"][0].WorkStatus != AgentWorkStatusCompleted {
		t.Fatal("late event revived completed execution")
	}
}

func TestTaskNotificationLegacyAndSnapshot(t *testing.T) {
	s := NewDevelopmentAIAgentClientStore()
	response := AIAgentTaskActionResponse{WorkStatus: AgentWorkStatusFailed}
	s.appendTaskNotificationLocked(TaskEvent{EventID: "old", WorkspaceID: "ws"}, response)
	for _, event := range s.events {
		if event.EventType == AgentClientEventTaskNotification {
			t.Fatal("legacy assignment inferred recipient")
		}
	}
	event := AgentTaskNotificationEvent{EventType: AgentClientEventTaskNotification, RecipientPrincipalID: "assigner", WorkspaceID: "ws", NotificationType: "failed"}
	raw, _ := json.Marshal(event)
	restored, err := restoreSnapshotEventPayload(AgentClientEventTaskNotification, raw)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := restored.(AgentTaskNotificationEvent); !ok {
		t.Fatalf("snapshot lost recipient type: %T", restored)
	}
	if clientEventVisibleToLocked(s, AuthorizationResult{PrincipalID: "owner", WorkspaceID: "ws"}, ClientStreamEvent{EventType: AgentClientEventTaskNotification, Payload: restored}) {
		t.Fatal("snapshot leaked")
	}
}

type notificationAuthorizer struct{ allowed bool }

func (a notificationAuthorizer) Authorize(_ context.Context, _ string, req AuthorizationRequest) (AuthorizationResult, error) {
	if !a.allowed || req.TaskID != "task" || req.WorkspaceID != "ws" {
		return AuthorizationResult{}, ErrAuthorizationForbidden
	}
	return AuthorizationResult{PrincipalID: "assigner", WorkspaceID: "ws"}, nil
}

func TestTaskNotificationRechecksTaskAccess(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Authorization", "Bearer token")
	principal := AuthorizationResult{PrincipalID: "assigner", WorkspaceID: "ws"}
	event := ClientStreamEvent{EventType: AgentClientEventTaskNotification, Payload: AgentTaskNotificationEvent{RecipientPrincipalID: "assigner", WorkspaceID: "ws", TaskID: "task"}}
	for _, allowed := range []bool{true, false} {
		server := Server{}
		server.config.Authorizer = notificationAuthorizer{allowed}
		if got := server.clientNotificationAuthorized(request, principal, event); got != allowed {
			t.Fatalf("access=%v, got=%v", allowed, got)
		}
	}
}

type notificationAssignAuthorizer struct{}

func (notificationAssignAuthorizer) Authorize(context.Context, string, AuthorizationRequest) (AuthorizationResult, error) {
	return AuthorizationResult{PrincipalID: "assigner", WorkspaceID: "workspace"}, nil
}

func TestNotificationRecipientComesFromAuthenticatedAssignment(t *testing.T) {
	store := NewStore()
	defer store.Close()
	server := Server{assignment: store}
	server.config.Authorizer = notificationAssignAuthorizer{}
	request := httptest.NewRequest(http.MethodPost, "/v1/component-tasks/task/assignment", strings.NewReader(`{"component_id":"task","agent_id":"agent","runtime_provider":"codex","prompt":"work","created_by":"victim","workspace_id":"other"}`))
	request.Header.Set("Authorization", "Bearer token")
	response := httptest.NewRecorder()
	server.handleComponentTaskAssign(response, request, "task")
	if response.Code != 200 {
		t.Fatalf("status %d: %s", response.Code, response.Body.String())
	}
	var result struct {
		Assignment Assignment `json:"assignment"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Assignment.CreatedBy != "assigner" || result.Assignment.WorkspaceID != "workspace" {
		t.Fatalf("untrusted recipient accepted: %+v", result.Assignment)
	}
}

func TestWaitingReasonAndRecipientSurviveAssignmentEventPersistence(t *testing.T) {
	store := NewStore()
	defer store.Close()
	state := newStoreState()
	assignment, err := store.handleAssign(&state, "task", AssignRequest{ComponentID: "task", AgentID: "agent", RuntimeProvider: "codex", Prompt: "work", CreatedBy: "assigner", WorkspaceID: "workspace", TeamID: "team"}, true, false)
	if err != nil {
		t.Fatal(err)
	}
	assignment.State = AssignmentRunning
	state.assignments[assignment.ID] = assignment
	request := AgentEventRequest{AssignmentID: assignment.ID, TaskID: "task", State: AssignmentRunning, EventType: assignmentcontract.EventAssignmentWaitingForUser, ConfirmationReason: assignmentcontract.ConfirmationAnswerRequired}
	response, err := store.handleEvent(&state, "agent", request)
	if err != nil {
		t.Fatal(err)
	}
	event := response.Event
	retry, err := store.handleEvent(&state, "agent", request)
	if err != nil || retry.Event.EventID != event.EventID {
		t.Fatalf("waiting retry changed event identity: %+v %v", retry, err)
	}
	if event.EventID == "" || event.CreatedBy != "assigner" || event.WorkspaceID != "workspace" || event.TeamID != "team" || event.ConfirmationReason != assignmentcontract.ConfirmationAnswerRequired {
		t.Fatalf("event lost routing/recipient: %+v", event)
	}
	raw, err := json.Marshal(response)
	if err != nil {
		t.Fatal(err)
	}
	var restored AgentEventResponse
	if err := json.Unmarshal(raw, &restored); err != nil {
		t.Fatal(err)
	}
	if restored.Assignment.ConfirmationReason != assignmentcontract.ConfirmationAnswerRequired || restored.Event.EventID != event.EventID {
		t.Fatal("snapshot lost waiting/identity")
	}
	request.EventType = EventAssignmentRunning
	request.ConfirmationReason = ""
	response, err = store.handleEvent(&state, "agent", request)
	if err != nil || response.Assignment.ConfirmationReason != "" {
		t.Fatalf("resume: %+v %v", response, err)
	}
	request.State = AssignmentFailed
	request.EventType = EventAssignmentFailed
	response, err = store.handleEvent(&state, "agent", request)
	if err != nil {
		t.Fatal(err)
	}
	retry, err = store.handleEvent(&state, "agent", request)
	if err != nil || retry.Event.EventID != response.Event.EventID {
		t.Fatalf("terminal retry changed event identity: %+v %v", retry, err)
	}
	client := NewDevelopmentAIAgentClientStore()
	client.appendTaskNotificationLocked(response.Event, AIAgentTaskActionResponse{WorkStatus: AgentWorkStatusFailed, RunID: "run", AssignmentState: AgentAssignmentStateFailed})
	last := client.events[len(client.events)-1].Payload.(AgentTaskNotificationEvent)
	if last.NotificationType != "failed" || last.RecipientPrincipalID != "assigner" {
		t.Fatalf("terminal failure: %+v", last)
	}
}

func TestTaskNotificationIsNotReplayedOnNewSubscription(t *testing.T) {
	client := NewDevelopmentAIAgentClientStore()
	client.events = []ClientStreamEvent{{Seq: 1, EventType: AgentClientEventTaskNotification, Payload: AgentTaskNotificationEvent{EventType: AgentClientEventTaskNotification, WorkspaceID: "workspace", RecipientPrincipalID: "assigner", TaskID: "task", EventID: "old"}}}
	server := Server{aiAgent: client}
	server.config.Authorizer = notificationAssignAuthorizer{}
	request := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/events?replay=1", nil)
	request.Header.Set("Authorization", "Bearer token")
	response := httptest.NewRecorder()
	server.handleAIAgentClientEvents(response, request)
	if response.Code != 200 || strings.Contains(response.Body.String(), AgentClientEventTaskNotification) {
		t.Fatalf("historical notification replayed: %d %s", response.Code, response.Body.String())
	}
}

func TestTaskNotificationAfterTerminalSnapshotRepair(t *testing.T) {
	s := NewDevelopmentAIAgentClientStore()
	s.agents["agent"] = AgentClientRecord{AgentID: "agent", WorkspaceID: "ws", OwnerPrincipalID: "assigner"}
	s.taskThreads["task"] = []AIAgentTaskThreadRecord{{ThreadID: "thread", TaskID: "task", AgentID: "agent", AssignmentID: "assignment", RunID: "run", WorkStatus: AgentWorkStatusCompleted, AssignmentState: AgentAssignmentStateCompleted}}
	event := TaskEvent{EventID: "assignment:5", Seq: 5, TaskID: "task", AssignmentID: "assignment", AgentID: "agent", WorkspaceID: "ws", CreatedBy: "assigner", State: AssignmentCompleted, Type: EventAssignmentCompleted, At: time.Now()}
	for i := 0; i < 2; i++ {
		if err := s.RecordAIAgentAssignmentEvent(context.Background(), "agent", AgentEventRequest{}, event); err != nil {
			t.Fatal(err)
		}
	}
	count := 0
	for _, e := range s.events {
		if n, ok := e.Payload.(AgentTaskNotificationEvent); ok {
			count++
			if n.NotificationType != "result_ready" {
				t.Fatal(n)
			}
		}
	}
	if count != 1 {
		t.Fatalf("notifications after snapshot repair: %d", count)
	}
}

func TestNeedsInputResultNotifiesAnswerInsteadOfCompletion(t *testing.T) {
	s := NewDevelopmentAIAgentClientStore()
	s.events = nil
	s.appendTaskNotificationLocked(TaskEvent{EventID: "event", AssignmentID: "assignment", WorkspaceID: "ws", CreatedBy: "user", State: AssignmentCompleted, Metadata: map[string]string{"assignment_result_status": "needs_input"}}, AIAgentTaskActionResponse{WorkStatus: AgentWorkStatusWaitingForUser})
	if len(s.events) != 1 {
		t.Fatalf("events: %+v", s.events)
	}
	if got := s.events[0].Payload.(AgentTaskNotificationEvent).NotificationType; got != "answer_required" {
		t.Fatal(got)
	}
}

func TestTaskNotificationReactionLifecycleExclusions(t *testing.T) {
	for _, tc := range []struct {
		name       string
		target     string
		invocation *CommentInvocation
		state      AssignmentState
		work       AgentWorkStatus
		want       int
	}{
		{name: "started once", target: "message", state: AssignmentRunning, work: AgentWorkStatusRunning, want: 1},
		{name: "without reaction target", state: AssignmentRunning, work: AgentWorkStatusRunning, want: 0},
		{name: "agent to agent", target: "message", invocation: &CommentInvocation{OriginMessageID: "parent-message"}, state: AssignmentCompleted, work: AgentWorkStatusCompleted, want: 0},
		{name: "cancelled", target: "message", state: AssignmentCancelled, work: AgentWorkStatusCompleted, want: 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store := NewDevelopmentAIAgentClientStore()
			store.taskThreads["task"] = []AIAgentTaskThreadRecord{{TaskID: "task", AssignmentID: "run", SourceMessageID: tc.target, ReactionSourceMessageID: &tc.target, CommentInvocation: tc.invocation}}
			for _, id := range []string{"accepted", "running"} {
				store.appendTaskNotificationLocked(TaskEvent{EventID: id, TaskID: "task", AssignmentID: "run", WorkspaceID: "ws", CreatedBy: "user", State: tc.state}, AIAgentTaskActionResponse{WorkStatus: tc.work})
			}
			count := 0
			for _, e := range store.events {
				if n, ok := e.Payload.(AgentTaskNotificationEvent); ok {
					count++
					if n.NotificationType != "started" {
						t.Fatalf("unexpected kind %s", n.NotificationType)
					}
				}
			}
			if count != tc.want {
				t.Fatalf("notifications=%d want %d", count, tc.want)
			}
		})
	}
}
