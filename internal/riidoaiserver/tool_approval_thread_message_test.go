package riidoaiserver

import (
	"testing"
	"time"
)

const (
	approvalChatBody          = "<p>너가 직접 진행해줘 승인할게</p>"
	approvalExecutionChatBody = "<p>go 명령 실행도 해줘</p>"
)

func TestHTTPThreadMessageApprovesPendingToolApproval(t *testing.T) {
	assertHTTPThreadMessageApprovesPendingToolApproval(t, approvalChatBody)
}

func TestHTTPThreadMessageApprovesPendingToolApprovalExecutionReply(t *testing.T) {
	assertHTTPThreadMessageApprovesPendingToolApproval(t, approvalExecutionChatBody)
}

func assertHTTPThreadMessageApprovesPendingToolApproval(t *testing.T, body string) {
	t.Helper()
	server := newApprovalChatTestServer(t)
	assigned := assignApprovalRoundTripTask(t, server)
	pollApprovalRoundTripTask(t, server, assigned.AssignmentID)
	createApprovalRoundTripRequest(t, server, assigned.AssignmentID)

	waitDone := make(chan ToolApprovalWaitResponse, 1)
	go func() {
		waitDone <- waitApprovalRoundTripDecision(t, server, assigned.AssignmentID)
	}()
	time.Sleep(50 * time.Millisecond)

	reply := postApprovalChatThreadMessage(t, server, assigned, body)
	if reply.AssignmentID != assigned.AssignmentID || reply.ThreadID != assigned.ThreadID {
		t.Fatalf("approval reply created new work: assigned=%+v reply=%+v", assigned, reply)
	}
	assertApprovalChatWaitApproved(t, waitDone)
	assertApprovalChatHistoryMessage(t, server, assigned.ThreadID, body)
}
