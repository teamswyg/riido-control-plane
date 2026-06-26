package riidoaiserver

import "testing"

func TestHTTPThreadMessageApprovalReplyWithoutPendingFollowsNeedsInputThread(t *testing.T) {
	handler, _, assignmentStore := newIntentDialogueHTTPTestServer(t)
	root := postIntentDialogueAssignment(t, handler, "agent-owned-codex")
	assertIntentGateAction(t, root)
	assertIntentDialoguePollNone(t, assignmentStore, root)

	reply := postIntentDialogueThreadMessage(t, handler, root, "승인할게 진행해줘")

	if reply.Message == clientMessageToolApprovalUnavailable {
		t.Fatalf("needs-input approval-like reply was blocked: %+v", reply)
	}
	if reply.AssignmentID == root.AssignmentID || reply.ThreadID == root.ThreadID {
		t.Fatalf("needs-input followup did not create new run: root=%+v reply=%+v", root, reply)
	}
	if reply.AssignmentState == AgentAssignmentStateFailed {
		t.Fatalf("needs-input followup failed instead of continuing: %+v", reply)
	}
}
