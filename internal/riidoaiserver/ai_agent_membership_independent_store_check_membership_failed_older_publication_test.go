package riidoaiserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func checkMembershipFailedOlderPublication(t *testing.T, state string) {
	t.Helper()
	const taskID, agentID, token = "task-failed-publication-receipt", "agent-public-openclaw", "ai-agent-token"
	ctx := context.Background()
	f := newMembershipSharedDynamoFixture(t)
	clientA, storeA := newMembershipDynamoReplica(t, f)
	config := func(client AIAgentClientStore, store AssignmentStore) ServerConfig {
		return ServerConfig{AIAgentClient: client, Assignment: store, TaskContext: &assignmentHTTPTaskContextReader{contextSnapshot: aiAgentTaskContextHTTPFixture()}, Authorizer: aiAgentClientHTTPAuthorizer(t, []string{"ai-agent:*"}, "user-1")}
	}
	remover := NewServer(config(clientA, storeA)).Handler()
	base := "/v2/client/workspaces/" + defaultAIAgentClientWorkspaceID + "/ai-agent"
	first := createProviderMultiAssignment(t, remover, base, token, taskID, agentID)
	stop := httptest.NewRequest(http.MethodPost, base+"/tasks/"+taskID+"/agent-assignments/"+agentID+"/stop", strings.NewReader(`{"assignment_id":"`+first.AssignmentID+`"}`))
	stop.Header.Set(aiAgentTokenHeader, token)
	stopped := httptest.NewRecorder()
	remover.ServeHTTP(stopped, stop)
	if stopped.Code != http.StatusAccepted {
		t.Fatalf("stop=%d", stopped.Code)
	}
	clientB, storeB := newMembershipDynamoReplica(t, f)
	clientC, storeC := newMembershipDynamoReplica(t, f)
	paused := &durableBeforePublishGate{Store: storeB, entered: make(chan Assignment, 1), resume: make(chan struct{})}
	paused.armed.Store(true)
	publisherB := NewServer(config(clientB, paused)).Handler()
	publisherC := NewServer(config(clientC, storeC)).Handler()
	request := httptest.NewRequest(http.MethodPost, base+"/tasks/"+taskID+"/assignment", strings.NewReader(`{"agent_id":"`+agentID+`"}`))
	request.Header.Set(aiAgentTokenHeader, token)
	posted := httptest.NewRecorder()
	finished := make(chan struct{})
	go func() { publisherB.ServeHTTP(posted, request); close(finished) }()
	var older Assignment
	select {
	case older = <-paused.entered:
	case <-time.After(3 * time.Second):
		close(paused.resume)
		<-finished
		t.Fatal("older publication did not pause")
	}
	if state == "running" {
		poll, err := storeB.PollAgent(ctx, agentID, PollRequest{DaemonID: "daemon-shared-studio", DeviceID: "device-shared-studio", RuntimeID: "runtime-openclaw-shared"})
		if err != nil || poll.Assignment == nil || poll.Assignment.ID != older.ID {
			close(paused.resume)
			<-finished
			t.Fatalf("older poll=%+v err=%v", poll, err)
		}
		mustRecordActorEvent(t, storeB, ctx, agentID, AgentEventRequest{AssignmentID: older.ID, DaemonID: "daemon-shared-studio", DeviceID: "device-shared-studio", RuntimeID: "runtime-openclaw-shared", State: AssignmentRunning, EventType: EventAssignmentRunning})
	}
	// C started before B's durable creation, so both executions really exist.
	newerRequest := httptest.NewRequest(http.MethodPost, base+"/tasks/"+taskID+"/assignment", strings.NewReader(`{"agent_id":"`+agentID+`"}`))
	newerRequest.Header.Set(aiAgentTokenHeader, token)
	newerResult := httptest.NewRecorder()
	publisherC.ServeHTTP(newerResult, newerRequest)
	close(paused.resume)
	<-finished
	if newerResult.Code != http.StatusAccepted || posted.Code != http.StatusConflict {
		t.Fatalf("newer=%d older=%d; want accepted/conflict", newerResult.Code, posted.Code)
	}
	profiles, err := clientA.ListWorkspaceAssignedAgentProfiles(ctx, AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID})
	if err != nil {
		t.Fatal(err)
	}
	latest := profiles.TaskAgentParticipants[taskID].Agents[0].LatestAssignmentID
	if latest == older.ID || latest == "" {
		t.Fatalf("failed older publication replaced latest identity: latest=%q older=%q", latest, older.ID)
	}
	removed := serveUnassignTaskBoundary(remover, base+"/tasks/"+taskID+"/agent-assignments/"+agentID, token, `{"assignment_id":"`+latest+`"}`)
	if removed.Code != http.StatusAccepted {
		t.Fatalf("remove=%d %s", removed.Code, removed.Body.String())
	}
	reader := storeA.operationStore.(AssignmentProjectionReader)
	for _, id := range []string{older.ID, latest} {
		projection, found, err := reader.LoadAssignmentProjection(ctx, id)
		want := AssignmentCancelled
		if state == "running" && id == older.ID {
			want = AssignmentCancelling
		}
		if err != nil || !found || projection.Assignment.State != want {
			t.Fatalf("hidden or current execution survived removal: id=%s projection=%+v err=%v", id, projection, err)
		}
	}
	if state == "running" {
		poll, err := storeB.PollAgent(ctx, agentID, PollRequest{DaemonID: "daemon-shared-studio", DeviceID: "device-shared-studio", RuntimeID: "runtime-openclaw-shared"})
		if err != nil || poll.Action != PollCancel || poll.Assignment == nil || poll.Assignment.ID != older.ID {
			t.Fatalf("running unpublished execution did not observe durable cancellation: %+v err=%v", poll, err)
		}
		mustRecordActorEvent(t, storeB, ctx, agentID, AgentEventRequest{AssignmentID: older.ID, DaemonID: "daemon-shared-studio", DeviceID: "device-shared-studio", RuntimeID: "runtime-openclaw-shared", State: AssignmentCancelled, EventType: EventAssignmentCancelled})
	}
	_, active, err := storeA.loadTaskAgentDurableExecutions(ctx, taskID, agentID)
	if err != nil || len(active) != 0 {
		t.Fatalf("terminal receipts remained active: %+v err=%v", active, err)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, item := range f.items {
		if item["pk"]["S"] == dynamoDBTaskAgentPK(taskID, agentID) && strings.HasPrefix(item["sk"]["S"], "EXEC#") {
			t.Fatalf("terminal creation receipt was not pruned: %+v", item)
		}
	}
}
