package riidoaiserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func checkMembershipIndependentStorePublication(t *testing.T, mode, state, history string) {
	t.Helper()
	const taskID, agentID, token = "task-independent-replica-removal", "agent-public-openclaw", "ai-agent-token"
	f := newMembershipSharedDynamoFixture(t)
	clientA, storeA := newMembershipDynamoReplica(t, f)
	config := func(client AIAgentClientStore, assignments AssignmentStore) ServerConfig {
		return ServerConfig{AIAgentClient: client, Assignment: assignments, TaskContext: &assignmentHTTPTaskContextReader{contextSnapshot: aiAgentTaskContextHTTPFixture()}, Authorizer: aiAgentClientHTTPAuthorizer(t, []string{"ai-agent:*"}, "user-1")}
	}
	remover := NewServer(config(clientA, storeA)).Handler()
	base := "/v2/client/workspaces/" + defaultAIAgentClientWorkspaceID + "/ai-agent"
	var first AIAgentTaskActionResponse
	if history != "absent" {
		first = createProviderMultiAssignment(t, remover, base, token, taskID, agentID)
		stop := httptest.NewRequest(http.MethodPost, base+"/tasks/"+taskID+"/agent-assignments/"+agentID+"/stop", strings.NewReader(`{"assignment_id":"`+first.AssignmentID+`"}`))
		stop.Header.Set(aiAgentTokenHeader, token)
		stopped := httptest.NewRecorder()
		remover.ServeHTTP(stopped, stop)
		if stopped.Code != http.StatusAccepted {
			t.Fatalf("initial stop=%d %s", stopped.Code, stopped.Body.String())
		}
	}
	if history == "compacted" && mode != "reply" {
		snapshot, _, err := clientA.snapshotStore.LoadAIAgentClientSnapshot(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		delete(snapshot.TaskThreads, taskID)
		snapshot.ExpectedVersion = snapshot.Version
		snapshot.Version++
		if err := clientA.snapshotStore.SaveAIAgentClientSnapshot(context.Background(), snapshot); err != nil {
			t.Fatal(err)
		}
	}
	clientB, storeB := newMembershipDynamoReplica(t, f)
	paused := &durableBeforePublishGate{Store: storeB, entered: make(chan Assignment, 1), resume: make(chan struct{})}
	paused.armed.Store(true)
	publisher := NewServer(config(clientB, paused)).Handler()
	postPath, postBody := base+"/tasks/"+taskID+"/assignment", `{"agent_id":"`+agentID+`"}`
	if mode == "additive" {
		postPath = base + "/tasks/" + taskID + "/agent-assignments"
	}
	if mode == "reply" {
		postPath = base + "/tasks/" + taskID + "/threads/" + first.ThreadID + "/messages"
		postBody = `{"body":"continue this task"}`
	}
	post := httptest.NewRequest(http.MethodPost, postPath, strings.NewReader(postBody))
	post.Header.Set(aiAgentTokenHeader, token)
	posted := httptest.NewRecorder()
	finished := make(chan struct{})
	go func() { publisher.ServeHTTP(posted, post); close(finished) }()
	var next Assignment
	select {
	case next = <-paused.entered:
	case <-time.After(3 * time.Second):
		close(paused.resume)
		<-finished
		t.Fatalf("did not reach durable-before-publication boundary: POST=%d %s", posted.Code, posted.Body.String())
	}
	if state == "completed" {
		poll, err := storeB.PollAgent(context.Background(), agentID, PollRequest{DaemonID: "daemon-shared-studio", DeviceID: "device-shared-studio", RuntimeID: "runtime-openclaw-shared"})
		if err != nil || poll.Assignment == nil || poll.Assignment.ID != next.ID {
			close(paused.resume)
			<-finished
			t.Fatalf("unpublished poll=%+v err=%v", poll, err)
		}
		mustRecordActorEvent(t, storeB, context.Background(), agentID, AgentEventRequest{AssignmentID: next.ID, DaemonID: "daemon-shared-studio", DeviceID: "device-shared-studio", RuntimeID: "runtime-openclaw-shared", State: AssignmentRunning, EventType: EventAssignmentRunning})
		mustRecordActorEvent(t, storeB, context.Background(), agentID, AgentEventRequest{AssignmentID: next.ID, DaemonID: "daemon-shared-studio", DeviceID: "device-shared-studio", RuntimeID: "runtime-openclaw-shared", State: AssignmentCompleted, EventType: EventAssignmentCompleted})
	}
	removeBody := `{"reason":"participant_removed"}`
	if history == "observed" {
		removeBody = `{"assignment_id":"` + first.AssignmentID + `","reason":"participant_removed"}`
	}
	removed := serveUnassignTaskBoundary(remover, base+"/tasks/"+taskID+"/agent-assignments/"+agentID, token, removeBody)
	close(paused.resume)
	<-finished
	if posted.Code != http.StatusAccepted {
		t.Fatalf("new execution publication=%d %s", posted.Code, posted.Body.String())
	}
	if removed.Code != http.StatusConflict {
		t.Fatalf("independent replica removed membership across unpublished durable execution: DELETE=%d, expected 409", removed.Code)
	}
	profiles, err := clientA.ListWorkspaceAssignedAgentProfiles(context.Background(), AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID})
	if err != nil || len(profiles.TaskAgentParticipants[taskID].Agents) != 1 {
		t.Fatalf("publication lost membership: %+v err=%v", profiles.TaskAgentParticipants, err)
	}
}
