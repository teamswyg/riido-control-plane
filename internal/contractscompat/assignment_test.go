package contractscompat

import (
	"testing"

	"github.com/teamswyg/riido-contracts/assignment"
)

func TestAssignmentFSMBaseline(t *testing.T) {
	if assignment.SchemaVersion != "riido-ai-server.v1" {
		t.Fatalf("assignment SchemaVersion = %q", assignment.SchemaVersion)
	}
	if !assignment.CanTransition(assignment.AssignmentQueued, assignment.AssignmentLeased) {
		t.Fatal("Queued -> Leased assignment transition must remain legal")
	}
	if assignment.GeneratedAssignmentFSMServiceProvider().AssignmentFSM().Name() != "assignment" {
		t.Fatal("assignment FSM service provider must return the generated assignment FSM")
	}
	if !assignment.GeneratedAssignmentFSM().CanTransition(assignment.AssignmentStateCodeQueued, assignment.AssignmentStateCodeLeased) {
		t.Fatal("Generated assignment FSM must keep queued -> leased transition")
	}
}

func TestAssignmentApprovalBaseline(t *testing.T) {
	if assignment.ApprovalTimeoutTerminalStatus != assignment.ApprovalTimedOut {
		t.Fatal("approval timeout must resolve to the timed_out terminal status")
	}
	if !assignment.ApprovalTimedOut.IsTerminal() || assignment.ApprovalPending.IsTerminal() {
		t.Fatal("approval terminal predicates drifted")
	}
	if assignment.ApprovalDecisionApprove.Code() != assignment.ApprovalDecisionCodeApprove {
		t.Fatal("approval decision enum drifted")
	}
}
