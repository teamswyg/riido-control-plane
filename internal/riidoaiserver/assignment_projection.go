package riidoaiserver

import (
	"time"

	"github.com/teamswyg/riido-contracts/assignment"
)

const workspacePreparationFailedDiagnostic = "workspace_preparation_failed"

func isWorkspacePreparationFailure(previous Assignment, event TaskEvent) bool {
	return previous.State == AssignmentFailed &&
		event.Type == EventAssignmentFailed && event.State == AssignmentFailed &&
		event.AssignmentID == previous.ID && event.TaskID == previous.TaskID &&
		event.AgentID == previous.AgentID &&
		event.Metadata[assignment.ExecutionDiagnosticKey] == workspacePreparationFailedDiagnostic
}

type AssignmentProjection struct {
	Assignment                 Assignment
	LastEventSeq               int64
	Recovery                   *recoveryWatermark          `json:"-"`
	EventACKs                  []assignmentEventACKReceipt `json:"-"`
	WorkspacePreparationFailed bool                        `json:"-"`
	ContextTerminalAt          time.Time                   `json:"-"`
}
