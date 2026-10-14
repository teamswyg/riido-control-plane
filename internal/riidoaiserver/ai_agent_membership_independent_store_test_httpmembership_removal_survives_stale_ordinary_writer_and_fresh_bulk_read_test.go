package riidoaiserver

import (
	"context"
	"net/http"
	"testing"
)

func TestHTTPMembershipRemovalSurvivesStaleOrdinaryWriterAndFreshBulkRead(t *testing.T) {
	const taskID, agentID, token = "task-stale-ordinary-writer", "agent-public-openclaw", "ai-agent-token"
	ctx := context.Background()
	p := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
	f := newMembershipSharedDynamoFixture(t)
	clientA, storeA := newMembershipDynamoReplica(t, f)
	config := ServerConfig{AIAgentClient: clientA, Assignment: storeA, TaskContext: &assignmentHTTPTaskContextReader{contextSnapshot: aiAgentTaskContextHTTPFixture()}, Authorizer: aiAgentClientHTTPAuthorizer(t, []string{"ai-agent:*"}, "user-1")}
	handler := NewServer(config).Handler()
	base := "/v2/client/workspaces/" + defaultAIAgentClientWorkspaceID + "/ai-agent"
	first := createProviderMultiAssignment(t, handler, base, token, taskID, agentID)
	peer := createProviderMultiAssignment(t, handler, base, token, taskID, "agent-owned-codex")
	other := createProviderMultiAssignment(t, handler, base, token, "task-other-relation", agentID)
	clientB, _ := newMembershipDynamoReplica(t, f)
	if err := clientB.reloadSnapshot(ctx); err != nil {
		t.Fatal(err)
	}
	removed := serveUnassignTaskBoundary(handler, base+"/tasks/"+taskID+"/agent-assignments/"+agentID, token, `{"assignment_id":"`+first.AssignmentID+`"}`)
	if removed.Code != http.StatusAccepted {
		t.Fatalf("remove=%d %s", removed.Code, removed.Body.String())
	}
	// This unrelated profile mutation writes the other replica's cached snapshot.
	if _, err := clientB.UpdateAIAgentConfiguration(ctx, p, "agent-owned-claude", UpdateAgentConfigurationRequest{Name: "Other Agent renamed"}); err != nil {
		t.Fatal(err)
	}
	profiles, err := clientB.ListWorkspaceAssignedAgentProfiles(ctx, p)
	if err != nil {
		t.Fatal(err)
	}
	members := profiles.TaskAgentParticipants[taskID].Agents
	if len(members) != 1 || members[0].AgentID != "agent-owned-codex" {
		t.Fatalf("stale writer revived member or lost peer: %+v", members)
	}
	if len(profiles.TaskAgentParticipants["task-other-relation"].Agents) != 1 {
		t.Fatalf("other task membership lost: %+v", profiles.TaskAgentParticipants)
	}
	reader := storeA.operationStore.(AssignmentProjectionReader)
	for _, id := range []string{peer.AssignmentID, other.AssignmentID} {
		projection, found, err := reader.LoadAssignmentProjection(ctx, id)
		if err != nil || !found || isTerminal(projection.Assignment.State) {
			t.Fatalf("peer execution stopped: id=%s projection=%+v err=%v", id, projection, err)
		}
	}
}

func TestHTTPMembershipRemovalCancelsExecutionLeftByFailedOlderPublication(t *testing.T) {
	for _, state := range []string{"queued", "running"} {
		t.Run(state, func(t *testing.T) { checkMembershipFailedOlderPublication(t, state) })
	}
}
