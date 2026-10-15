package riidoaiserver

import (
	"context"
	"testing"
)

func TestMembershipFollowupAfterRemovalRestoresParticipant(t *testing.T) {
	ctx := context.Background()
	s := NewDevelopmentAIAgentClientStore()
	p := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
	root := assignConversationThread(t, s, ctx, p, "membership-followup", "asn-first")
	markConversationThreadCompleted(s, "membership-followup", root.ThreadID, root.AgentID)
	if _, err := s.DeleteAIAgentTaskAgentAssignment(ctx, p, "membership-followup", root.AgentID, AgentAssignmentActionRequest{}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateAIAgentTaskThreadMessage(ctx, p, "membership-followup", root.ThreadID, CreateAIAgentTaskThreadMessageRequest{Body: "continue fixture", AssignmentID: "asn-next", durableState: AssignmentRunning}); err != nil {
		t.Fatal(err)
	}
	profiles, err := s.ListWorkspaceAssignedAgentProfiles(ctx, p)
	if err != nil {
		t.Fatal(err)
	}
	members := profiles.TaskAgentParticipants["membership-followup"].Agents
	if len(members) != 1 || members[0].AgentID != root.AgentID || !members[0].IsStopRequired {
		t.Fatalf("followup members = %+v", members)
	}
	snapshot, err := s.snapshot(fixedSnapshotTestNow())
	if err != nil {
		t.Fatal(err)
	}
	reopened := NewDevelopmentAIAgentClientStore()
	if err := reopened.restoreSnapshot(snapshot); err != nil {
		t.Fatal(err)
	}
	if ids := membershipAgentIDs(t, reopened, p, "membership-followup"); len(ids) != 1 {
		t.Fatalf("restart lost followup membership: %v", ids)
	}
}
