package riidoaiserver

import "testing"

func TestHTTPThreadMessageApprovalReplyWithoutPendingApprovalDoesNotCreateWork(t *testing.T) {
	server := newApprovalChatTestServer(t)
	assigned := assignApprovalRoundTripTask(t, server)

	reply := postApprovalChatThreadMessage(t, server, assigned, approvalChatBody)

	if reply.AssignmentID != assigned.AssignmentID || reply.ThreadID != assigned.ThreadID {
		t.Fatalf("stale approval reply created new work: assigned=%+v reply=%+v", assigned, reply)
	}
	if reply.WorkStatus != AgentWorkStatusFailed ||
		reply.AssignmentState != AgentAssignmentStateFailed ||
		reply.CommentKind != AgentTaskCommentTaskFailed {
		t.Fatalf("stale approval reply state = %+v", reply)
	}
	if reply.Message != clientMessageToolApprovalUnavailable ||
		reply.ResultMessage != clientMessageToolApprovalUnavailable ||
		reply.ActiveStream != nil {
		t.Fatalf("stale approval reply response = %+v", reply)
	}
	assertApprovalChatHistoryMessage(t, server, assigned.ThreadID, approvalChatBody)
	assertApprovalChatHistoryBody(t, server, assigned.ThreadID, clientMessageToolApprovalUnavailable)
}
