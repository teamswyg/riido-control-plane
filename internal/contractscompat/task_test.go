package contractscompat

import (
	"testing"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-contracts/task"
)

func TestTaskFSMBaseline(t *testing.T) {
	if !ir.EventTaskQueued.IsTransition() {
		t.Fatal("TaskQueued must remain a transition event")
	}
	if task.FSMSchemaVersion != 1 {
		t.Fatalf("FSMSchemaVersion = %d", task.FSMSchemaVersion)
	}
	if !task.ValidateTransition(task.StateCreated, task.StateQueued, ir.EventTaskQueued) {
		t.Fatal("Created -> Queued must remain legal")
	}
	if task.GeneratedTaskFSMServiceProvider().TaskFSM().Name() != "task" {
		t.Fatal("task FSM service provider must return the generated task FSM")
	}
	if !task.GeneratedTaskFSM().CanTransition(task.TaskStateCodeCreated, task.TaskStateCodeQueued, ir.EventTypeCodeTaskQueued) {
		t.Fatal("Generated task FSM must keep Created -> Queued transition")
	}
}
