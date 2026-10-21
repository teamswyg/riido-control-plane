package riidoaiserver

import (
	"errors"
	"fmt"

	"github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-contracts/metadatakeys"
)

func (s *Store) checkRecoveryWatermark(current Assignment, req AgentEventRequest) (TaskEvent, bool, error) {
	context := assignment.RecoveryContextFromMetadata(req.Metadata)
	watermark := s.recoveryFence
	if context == nil || watermark == nil {
		return TaskEvent{}, false, nil
	}
	keyHash := recoveryHash(req.Metadata[metadatakeys.AssignmentEventKey.String()])
	for _, receipt := range watermark.Receipts {
		if receipt.KeyHash != keyHash {
			continue
		}
		if receipt.ContextHash != recoveryContextHash(*context) {
			return TaskEvent{}, false, errors.New("recovery event key was reused with another transition")
		}
		return TaskEvent{EventID: fmt.Sprintf("%s:%d", current.ID, receipt.Seq), WorkspaceID: current.WorkspaceID, TeamID: current.TeamID, CreatedBy: current.CreatedBy, ConfirmationReason: receipt.ConfirmationReason, Seq: receipt.Seq, TaskID: current.TaskID, AssignmentID: current.ID, AgentID: current.AgentID, Type: EventRiidoLog, State: receipt.State, Metadata: cloneMetadata(req.Metadata), At: receipt.At}, true, nil
	}
	if len(watermark.Receipts) >= maxRecoveryReceipts {
		return TaskEvent{}, false, errors.New("provider recovery journal budget exhausted")
	}
	if recoveryMovedBackwards(watermark.Context, *context) {
		return TaskEvent{}, false, errors.New("recovery event belongs to an earlier transition")
	}
	return TaskEvent{}, false, nil
}

func recoveryMovedBackwards(previous, next assignment.RecoveryContext) bool {
	if next.Attempt < previous.Attempt || next.SwitchCount < previous.SwitchCount {
		return true
	}
	if next.State == "cancelled" {
		return false
	}
	switch previous.State {
	case "recovered", "cancelled", "unavailable", "user_action_required":
		return true
	case "dispatched":
		return next.State != "dispatched" && next.State != "recovered" && next.State != "user_action_required" && next.State != "unavailable"
	}
	return false
}
