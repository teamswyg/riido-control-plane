package riidoaiserver

import (
	"testing"
	"time"
)

func TestHTTPToolApprovalRoundTrip(t *testing.T) {
	server := newAIAgentClientHTTPTestServer(t, []StaticTokenCredential{{
		PrincipalID: "user-1",
		Token:       "round-trip-token",
		Scopes: []string{
			"ai-agent:*",
			"agent:agent-public-openclaw:*",
		},
	}})
	assigned := assignApprovalRoundTripTask(t, server)
	pollApprovalRoundTripTask(t, server, assigned.AssignmentID)
	createApprovalRoundTripRequest(t, server, assigned.AssignmentID)
	assertApprovalRoundTripList(t, server, assigned.AssignmentID)

	waitDone := make(chan ToolApprovalWaitResponse, 1)
	go func() {
		waitDone <- waitApprovalRoundTripDecision(t, server, assigned.AssignmentID)
	}()
	time.Sleep(50 * time.Millisecond)
	decision := decideApprovalRoundTrip(t, server, assigned.AssignmentID)
	if decision.Result.Status != ApprovalApproved || decision.Decision == nil {
		t.Fatalf("decision response = %+v", decision)
	}
	select {
	case waited := <-waitDone:
		if waited.Result.Status != ApprovalApproved || waited.Decision == nil ||
			waited.Decision.DecidedBy != "user-1" {
			t.Fatalf("wait response = %+v", waited)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for approval decision")
	}
}
