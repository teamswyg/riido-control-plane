package riidoaiserver

import (
	"sort"
	"strconv"
)

func stateFromAssignmentOperations(records []AssignmentOperationRecord) (storeState, error) {
	ordered := append([]AssignmentOperationRecord(nil), records...)
	sort.SliceStable(ordered, func(i, j int) bool {
		leftSeq := assignmentOperationLastEventSeq(ordered[i])
		rightSeq := assignmentOperationLastEventSeq(ordered[j])
		if leftSeq != rightSeq {
			return leftSeq < rightSeq
		}
		if !ordered[i].RecordedAt.Equal(ordered[j].RecordedAt) {
			return ordered[i].RecordedAt.Before(ordered[j].RecordedAt)
		}
		return ordered[i].OperationID < ordered[j].OperationID
	})

	state := newStoreState()
	seenEvents := map[string]bool{}
	for _, record := range ordered {
		if err := validateAssignmentOperationRecord(record); err != nil {
			return storeState{}, err
		}
		for _, event := range record.Events {
			eventKey := assignmentOperationEventReplayKey(event)
			if seenEvents[eventKey] {
				continue
			}
			seenEvents[eventKey] = true
			state.events[event.TaskID] = append(state.events[event.TaskID], event)
			if event.Seq > state.nextEventSeq {
				state.nextEventSeq = event.Seq
			}
			if event.AssignmentID != "" && event.State != "" {
				assignment := state.assignments[event.AssignmentID]
				if assignment.ID != "" {
					assignment.State = event.State
					if !event.At.IsZero() {
						assignment.UpdatedAt = event.At
					}
					state.assignments[assignment.ID] = assignment
				}
			}
		}
		state.assignments[record.Assignment.ID] = record.Assignment
		if seq := assignmentSequence(record.Assignment.ID); seq > state.nextAssignmentSeq {
			state.nextAssignmentSeq = seq
		}
	}
	for taskID := range state.events {
		sort.Slice(state.events[taskID], func(i, j int) bool {
			return state.events[taskID][i].Seq < state.events[taskID][j].Seq
		})
	}
	rebuildAssignmentIndexes(&state)
	return state, nil
}

func rebuildAssignmentIndexes(state *storeState) {
	state.tasks = map[string]taskRecord{}
	state.agentAssignments = map[string][]string{}
	assignments := make([]Assignment, 0, len(state.assignments))
	for _, assignment := range state.assignments {
		assignments = append(assignments, assignment)
	}
	sort.Slice(assignments, func(i, j int) bool {
		if !assignments[i].CreatedAt.Equal(assignments[j].CreatedAt) {
			return assignments[i].CreatedAt.Before(assignments[j].CreatedAt)
		}
		return assignments[i].ID < assignments[j].ID
	})
	for _, assignment := range assignments {
		state.agentAssignments[assignment.AgentID] = append(state.agentAssignments[assignment.AgentID], assignment.ID)
		task := state.tasks[assignment.TaskID]
		current := state.assignments[task.currentAssignmentID]
		if task.id == "" || assignmentIsNewer(assignment, current) {
			state.tasks[assignment.TaskID] = taskRecord{
				id:                  assignment.TaskID,
				componentID:         assignment.ComponentID,
				currentAssignmentID: assignment.ID,
			}
		}
	}
}

func assignmentIsNewer(candidate, current Assignment) bool {
	if current.ID == "" {
		return true
	}
	if !candidate.CreatedAt.Equal(current.CreatedAt) {
		return candidate.CreatedAt.After(current.CreatedAt)
	}
	return candidate.ID > current.ID
}

func assignmentOperationEventReplayKey(event TaskEvent) string {
	return event.TaskID + "\x00" + strconv.FormatInt(event.Seq, 10)
}
