package riidoaiserver

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestRequestReceiptsSurviveDisplayRetentionAndUnassignment(t *testing.T) {
	ctx := context.Background()
	s := NewDevelopmentAIAgentClientStore()
	p := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: "workspace-dev-riid"}
	first := AssignAIAgentTaskRequest{AgentID: "agent-public-openclaw", RequestID: "first", intentGateRequired: true}
	response, err := s.CreateAIAgentTaskAgentAssignment(ctx, p, "task", first)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 60; i++ {
		_, err := s.CreateAIAgentTaskAgentAssignment(ctx, p, "task", AssignAIAgentTaskRequest{AgentID: first.AgentID, AssignmentID: fmt.Sprintf("asn-request-%d", i), RequestID: fmt.Sprint(i)})
		if err != nil {
			t.Fatal(err)
		}
	}
	if _, err := s.UnassignAIAgentTask(ctx, p, "task", UnassignAIAgentTaskRequest{AgentID: first.AgentID}); err != nil {
		t.Fatal(err)
	}
	snap, err := s.snapshot(time.Now())
	if err != nil {
		t.Fatal(err)
	}
	restored := NewDevelopmentAIAgentClientStore()
	if err = restored.restoreSnapshot(snap); err != nil {
		t.Fatal(err)
	}
	replay, found, err := restored.FindTaskRequestReceipt(ctx, p, "task", first)
	if err != nil || !found || replay.AssignmentID != response.AssignmentID {
		t.Fatalf("receipt missing: %+v %v %v", replay, found, err)
	}
	if _, err = restored.CreateAIAgentTaskAgentAssignment(ctx, p, "task", first); err != nil {
		t.Fatal(err)
	}
	threads, err := restored.ListAIAgentTaskThreads(ctx, p, "task")
	if err != nil || len(threads.AssignedAgentIDs) != 0 {
		t.Fatalf("replay restored membership: %+v %v", threads, err)
	}
}
func TestSnapshotBackfillsLegacyMembershipBeforeFirstRead(t *testing.T) {
	s := NewDevelopmentAIAgentClientStore()
	s.taskAssignedAgents = nil
	s.taskThreads["legacy"] = []AIAgentTaskThreadRecord{{TaskID: "legacy", AgentID: "agent-public-openclaw", AssignmentState: AgentAssignmentStateRunning}}
	snap, err := s.snapshot(time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if len(snap.TaskAssignedAgents["legacy"]) != 1 {
		t.Fatalf("lost membership: %+v", snap.TaskAssignedAgents)
	}
}
func TestSnapshotKeepsEveryActiveRequest(t *testing.T) {
	s := NewDevelopmentAIAgentClientStore()
	for i := 0; i < 60; i++ {
		s.taskThreads["queued"] = append(s.taskThreads["queued"], AIAgentTaskThreadRecord{TaskID: "queued", ThreadID: fmt.Sprint(i), AgentID: "agent-public-openclaw", AssignmentState: AgentAssignmentStateQueued})
	}
	snap, err := s.snapshot(time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if len(snap.TaskThreads["queued"]) != 60 {
		t.Fatalf("queued requests lost: %d", len(snap.TaskThreads["queued"]))
	}
}
