package riidoaiserver

import (
	"time"

	"github.com/teamswyg/riido-contracts/assignment"
)

func recoveryProjectionEvent() TaskEvent {
	return TaskEvent{
		TaskID: "task", AssignmentID: "assignment", AgentID: "agent-owned-codex", State: AssignmentRunning,
		Type: EventRiidoLog, At: time.Now().UTC(), Metadata: (assignment.RecoveryContext{
			State: "checking", Reason: "transient_preflight", Provider: "codex", Attempt: 1,
			Circuit: "closed", OutcomeCertainty: "not_started",
		}).Metadata(),
	}
}
