package riidoaiserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHTTPFailedPublicationRecoveryPublishesCancellationAfterOtherActorEvent(t *testing.T) {
	ctx := context.Background()
	s := newMembershipNativeScenario(t, "task-recovery-event-sequence")
	first := s.stoppedAssignment(t)
	client, store := newMembershipDynamoReplica(t, s.f)
	paused := &durableBeforePublishGate{Store: store, entered: make(chan Assignment, 1), resume: make(chan struct{})}
	paused.armed.Store(true)
	s.f.mu.Lock()
	s.f.rejectScopedPublication = true
	s.f.mu.Unlock()
	result := make(chan *httptest.ResponseRecorder, 1)
	go func() {
		result <- s.post(s.replicaHandler(client, paused), "/assignment", `{"agent_id":"`+s.agentID+`"}`)
	}()
	var pending Assignment
	select {
	case pending = <-paused.entered:
	case <-time.After(3 * time.Second):
		close(paused.resume)
		t.Fatal("durable creation did not pause")
	}
	_, live, cancel, err := store.SubscribeTask(ctx, s.taskID)
	if err != nil {
		close(paused.resume)
		t.Fatal(err)
	}
	defer cancel()
	competing, err := store.RecordAgentEvent(ctx, s.agentID, AgentEventRequest{AssignmentID: first.AssignmentID, DaemonID: "daemon-shared-studio", DeviceID: "device-shared-studio", RuntimeID: "runtime-openclaw-shared", EventType: EventRiidoLog, Message: "older execution final runtime line"})
	if err != nil {
		close(paused.resume)
		t.Fatal(err)
	}
	close(paused.resume)
	if response := <-result; response.Code == http.StatusAccepted {
		t.Fatal("publication failure was accepted")
	}
	projection, found, err := store.LoadAssignmentProjection(ctx, pending.ID)
	if err != nil || !found || projection.Assignment.State != AssignmentCancelled {
		t.Fatalf("native cancellation did not commit: %+v err=%v", projection, err)
	}
	if projection.LastEventSeq <= competing.Event.Seq {
		t.Fatalf("recovery sequence collided: cancellation=%d competing=%d", projection.LastEventSeq, competing.Event.Seq)
	}
	replay, _, unsubscribe, err := store.SubscribeTask(ctx, s.taskID)
	if err != nil {
		t.Fatal(err)
	}
	defer unsubscribe()
	sawReplay, sawLive := false, false
	for _, event := range replay {
		if event.AssignmentID == pending.ID && event.Type == EventAssignmentCancelled {
			sawReplay = true
		}
	}
drain:
	for {
		select {
		case event := <-live:
			if event.AssignmentID == pending.ID && event.Type == EventAssignmentCancelled {
				sawLive = true
			}
		default:
			break drain
		}
	}
	if !sawReplay || !sawLive {
		t.Fatalf("committed cancellation dropped from event delivery: replay=%v live=%v", sawReplay, sawLive)
	}
}
