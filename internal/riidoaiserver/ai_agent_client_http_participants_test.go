package riidoaiserver

import "testing"

func TestHTTPAIAgentClientDevelopmentTaskAssignmentAndParticipantRemoval(t *testing.T) {
	const token = "user-token"
	server := newAIAgentClientHTTPTestServer(t, []StaticTokenCredential{{
		PrincipalID: "user-1",
		Token:       token,
		Scopes:      []string{"ai-agent:*", "task:task-new:read", "task:task-new:assign", "task:task-new:comment", "task:task-new:stop"},
	}})

	assigned := participantAssignOpenClaw(t, server, token)
	assertParticipantActiveThread(t, server, token, assigned.ThreadID)

	message := participantPostFollowup(t, server, token, assigned.ThreadID)
	if message.ThreadID != assigned.ThreadID ||
		message.AssignmentState != AgentAssignmentStateQueued ||
		message.CommentKind != AgentTaskCommentQueuedByBusyAgent {
		t.Fatalf("thread message response = %+v", message)
	}
	assertParticipantSourceMessage(t, server, token)

	unassigned := participantUnassignOpenClaw(t, server, token)
	if unassigned.ThreadID != assigned.ThreadID ||
		unassigned.AssignmentState != AgentAssignmentStateStopped ||
		unassigned.CommentKind != AgentTaskCommentStoppedByUserRequest {
		t.Fatalf("unassign response = %+v", unassigned)
	}
	assertParticipantStoppedThread(t, server, token)
	assertParticipantReplayOmitsStaleQueued(t, server, token)
}
