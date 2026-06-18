package riidoaiserver

import (
	"sort"
	"time"
)

const staleReplayQueuedAssignmentMaxAge = 24 * time.Hour

func replayStaleQueuedAssignmentIDs(state *storeState, now time.Time) []string {
	ids := []string{}
	for id, assignment := range state.assignments {
		if !assignmentQueuedPastMaxAge(assignment, now) {
			continue
		}
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func assignmentQueuedPastMaxAge(assignment Assignment, now time.Time) bool {
	if assignment.State.Code() != AssignmentStateCodeQueued || assignment.UpdatedAt.IsZero() {
		return false
	}
	return !assignment.UpdatedAt.Add(staleReplayQueuedAssignmentMaxAge).After(now)
}
