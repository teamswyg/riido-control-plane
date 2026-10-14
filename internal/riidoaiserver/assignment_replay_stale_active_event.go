package riidoaiserver

import (
	"time"

	"github.com/teamswyg/riido-contracts/metadatakeys"
)

func staleReplayActiveAssignmentEvent(assignment Assignment, seq int64, at time.Time) TaskEvent {
	metadata := staleReplayAssignmentMetadata("stale_replay_active_assignment")
	if assignment.LeaseToken != "" {
		metadata["lease_token"] = assignment.LeaseToken
	}
	return TaskEvent{
		Seq:          seq,
		TaskID:       assignment.TaskID,
		AssignmentID: assignment.ID,
		AgentID:      assignment.AgentID,
		Type:         EventAssignmentFailed,
		State:        AssignmentFailed,
		Message:      "active assignment replay repaired after stale projection drift",
		Metadata:     metadata,
		At:           at,
	}
}

func staleReplayQueuedAssignmentEvent(assignment Assignment, seq int64, at time.Time) TaskEvent {
	event := staleQueuedAssignmentEvent(assignment, seq, at)
	event.Message = "queued assignment replay repaired after stale queue timeout"
	event.Metadata = staleReplayAssignmentMetadata("stale_replay_queued_assignment")
	return event
}

func staleQueuedAssignmentEvent(assignment Assignment, seq int64, at time.Time) TaskEvent {
	return TaskEvent{
		Seq:          seq,
		TaskID:       assignment.TaskID,
		AssignmentID: assignment.ID,
		AgentID:      assignment.AgentID,
		Type:         EventAssignmentFailed,
		State:        AssignmentFailed,
		Message:      "queued assignment repaired after stale queue timeout",
		Metadata:     staleReplayAssignmentMetadata("stale_queued_assignment"),
		At:           at,
	}
}

func staleReplayAssignmentMetadata(category string) map[string]string {
	return map[string]string{
		metadatakeys.AssignmentFailureCategory.String(): category,
		metadatakeys.AssignmentResultStatus.String():    "failed",
	}
}
