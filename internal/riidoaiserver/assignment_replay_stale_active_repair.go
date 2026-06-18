package riidoaiserver

import (
	"sort"
	"time"
)

func repairStaleReplayActiveAssignments(state *storeState, now time.Time, activeLeaseDuration time.Duration) []AssignmentOperationRecord {
	if state == nil {
		return nil
	}
	if activeLeaseDuration <= 0 {
		activeLeaseDuration = time.Duration(DefaultAssignmentActiveLeaseSeconds) * time.Second
	}
	ids := replayStaleActiveAssignmentIDs(state, now.UTC(), activeLeaseDuration)
	repairs := make([]AssignmentOperationRecord, 0, len(ids))
	for _, id := range ids {
		repairs = append(repairs, repairStaleReplayActiveAssignment(state, state.assignments[id], now.UTC()))
	}
	if len(repairs) > 0 {
		rebuildAssignmentIndexes(state)
		rebuildStateMetricsFromHistory(state)
	}
	return repairs
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

func repairStaleReplayActiveAssignment(state *storeState, assignment Assignment, now time.Time) AssignmentOperationRecord {
	state.nextEventSeq++
	assignment.State = AssignmentFailed
	assignment.UpdatedAt = now
	state.assignments[assignment.ID] = assignment
	event := staleReplayActiveAssignmentEvent(assignment, state.nextEventSeq, now)
	state.events[assignment.TaskID] = append(state.events[assignment.TaskID], event)
	return AssignmentOperationRecord{
		SchemaVersion: AssignmentOperationSchemaVersion,
		OperationID:   assignmentOperationID(AssignmentOperationAgentEvent, assignment, []TaskEvent{event}),
		OperationType: AssignmentOperationAgentEvent,
		TaskID:        assignment.TaskID,
		AssignmentID:  assignment.ID,
		AgentID:       assignment.AgentID,
		Assignment:    assignment,
		Events:        []TaskEvent{event},
		RecordedAt:    now,
	}
}
