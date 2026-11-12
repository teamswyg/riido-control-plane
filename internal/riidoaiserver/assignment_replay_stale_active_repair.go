package riidoaiserver

import (
	"context"
	"sort"
	"time"
)

func repairStaleReplayAssignments(ctx context.Context, state *storeState, now time.Time, activeLeaseDuration time.Duration, leases AssignmentActiveLeaseStore) ([]AssignmentOperationRecord, error) {
	if state == nil {
		return nil, nil
	}
	if activeLeaseDuration <= 0 {
		activeLeaseDuration = time.Duration(DefaultAssignmentActiveLeaseSeconds) * time.Second
	}
	ids := replayStaleActiveAssignmentIDs(state, now.UTC(), activeLeaseDuration)
	ids, err := replayExpiredActiveAssignmentIDs(ctx, state, ids, now, leases)
	if err != nil {
		return nil, err
	}
	repairs := make([]AssignmentOperationRecord, 0, len(ids))
	for _, id := range ids {
		repairs = append(repairs, repairReplayAssignmentAsFailed(
			state,
			state.assignments[id],
			staleReplayActiveAssignmentEvent,
			now.UTC(),
		))
	}
	for _, id := range replayStaleQueuedAssignmentIDs(state, now.UTC()) {
		repairs = append(repairs, repairReplayAssignmentAsFailed(
			state,
			state.assignments[id],
			staleReplayQueuedAssignmentEvent,
			now.UTC(),
		))
	}
	if len(repairs) > 0 {
		rebuildAssignmentIndexes(state)
		rebuildStateMetricsFromHistory(state)
	}
	return repairs, nil
}

func replayStaleActiveAssignmentIDs(state *storeState, now time.Time, activeLeaseDuration time.Duration) []string {
	ids := []string{}
	for id, assignment := range state.assignments {
		if !assignmentHoldsActiveLease(assignment.State) || assignment.UpdatedAt.IsZero() {
			continue
		}
		if assignment.UpdatedAt.Add(activeLeaseDuration).After(now) {
			continue
		}
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}
