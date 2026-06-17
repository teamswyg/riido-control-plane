package riidoaiserver

import (
	"context"
	"sort"
)

func loadReplayAssignmentProjections(ctx context.Context, state *storeState, reader AssignmentProjectionReader) ([]AssignmentProjection, error) {
	if state == nil || reader == nil {
		return nil, nil
	}
	ids := replayProjectionRefreshIDs(state)
	projections := make([]AssignmentProjection, 0, len(ids))
	for _, id := range ids {
		projection, ok, err := reader.LoadAssignmentProjection(ctx, id)
		if err != nil {
			return nil, err
		}
		if ok {
			projections = append(projections, projection)
		}
	}
	return projections, nil
}

func overlayAssignmentProjections(state *storeState, projections []AssignmentProjection) {
	if state == nil || len(projections) == 0 {
		return
	}
	ordered := append([]AssignmentProjection(nil), projections...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].LastEventSeq != ordered[j].LastEventSeq {
			return ordered[i].LastEventSeq < ordered[j].LastEventSeq
		}
		return ordered[i].Assignment.ID < ordered[j].Assignment.ID
	})
	for _, projection := range ordered {
		assignment := projection.Assignment
		if assignment.ID == "" {
			continue
		}
		state.assignments[assignment.ID] = assignment
		if seq := assignmentSequence(assignment.ID); seq > state.nextAssignmentSeq {
			state.nextAssignmentSeq = seq
		}
	}
	rebuildAssignmentIndexes(state)
	rebuildStateMetricsFromHistory(state)
}

func replayProjectionRefreshIDs(state *storeState) []string {
	ids := []string{}
	for id, assignment := range state.assignments {
		if !isTerminal(assignment.State) {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	return ids
}
