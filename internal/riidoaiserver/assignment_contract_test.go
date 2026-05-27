package riidoaiserver

import (
	"testing"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
)

func TestAssignmentContractImportsSharedSurface(t *testing.T) {
	if SchemaVersion != assignmentcontract.SchemaVersion {
		t.Fatalf("SchemaVersion = %q, want %q", SchemaVersion, assignmentcontract.SchemaVersion)
	}

	statePairs := map[AssignmentState]assignmentcontract.AssignmentState{
		AssignmentQueued:     assignmentcontract.AssignmentQueued,
		AssignmentLeased:     assignmentcontract.AssignmentLeased,
		AssignmentReady:      assignmentcontract.AssignmentReady,
		AssignmentRunning:    assignmentcontract.AssignmentRunning,
		AssignmentCancelling: assignmentcontract.AssignmentCancelling,
		AssignmentCancelled:  assignmentcontract.AssignmentCancelled,
		AssignmentCompleted:  assignmentcontract.AssignmentCompleted,
		AssignmentFailed:     assignmentcontract.AssignmentFailed,
	}
	for local, shared := range statePairs {
		if local != shared {
			t.Fatalf("assignment state alias drift: local %q shared %q", local, shared)
		}
		if isTerminal(local) != assignmentcontract.IsTerminal(shared) {
			t.Fatalf("terminal classification drift for %q", local)
		}
		if isAgentActive(local) != assignmentcontract.IsAgentActive(shared) {
			t.Fatalf("agent-active classification drift for %q", local)
		}
	}

	actionPairs := map[PollAction]assignmentcontract.PollAction{
		PollNone:   assignmentcontract.PollNone,
		PollStart:  assignmentcontract.PollStart,
		PollCancel: assignmentcontract.PollCancel,
		PollActive: assignmentcontract.PollActive,
	}
	for local, shared := range actionPairs {
		if local != shared {
			t.Fatalf("poll action alias drift: local %q shared %q", local, shared)
		}
	}

	eventPairs := map[string]string{
		EventAssignmentQueued:       assignmentcontract.EventAssignmentQueued,
		EventAssignmentLeased:       assignmentcontract.EventAssignmentLeased,
		EventAssignmentReady:        assignmentcontract.EventAssignmentReady,
		EventAssignmentRunning:      assignmentcontract.EventAssignmentRunning,
		EventAssignmentCancelling:   assignmentcontract.EventAssignmentCancelling,
		EventAssignmentCancelled:    assignmentcontract.EventAssignmentCancelled,
		EventAssignmentCompleted:    assignmentcontract.EventAssignmentCompleted,
		EventAssignmentFailed:       assignmentcontract.EventAssignmentFailed,
		EventAssignmentStateUpdated: assignmentcontract.EventAssignmentStateUpdated,
		EventRiidoLog:               assignmentcontract.EventRiidoLog,
		EventProviderLog:            assignmentcontract.EventProviderLog,
		EventProviderWarning:        assignmentcontract.EventProviderWarning,
		EventProviderError:          assignmentcontract.EventProviderError,
	}
	for local, shared := range eventPairs {
		if local != shared {
			t.Fatalf("task event alias drift: local %q shared %q", local, shared)
		}
	}
}

func TestAssignmentTransitionDelegatesToSharedContract(t *testing.T) {
	for _, from := range assignmentcontract.AllAssignmentStates() {
		for _, to := range assignmentcontract.AllAssignmentStates() {
			if got, want := canTransitionAssignment(from, to), assignmentcontract.CanTransition(from, to); got != want {
				t.Fatalf("canTransitionAssignment(%q,%q) = %v, want %v", from, to, got, want)
			}
		}
	}
}
