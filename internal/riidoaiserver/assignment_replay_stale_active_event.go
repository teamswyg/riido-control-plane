package riidoaiserver

import (
	"time"

	"github.com/teamswyg/riido-contracts/metadatakeys"
)

func staleReplayActiveAssignmentEvent(assignment Assignment, seq int64, at time.Time) TaskEvent {
	metadata := map[string]string{
		metadatakeys.AssignmentFailureCategory.String(): "stale_replay_active_assignment",
		metadatakeys.AssignmentResultStatus.String():    "failed",
	}
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
