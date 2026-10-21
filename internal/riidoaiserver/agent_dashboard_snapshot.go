package riidoaiserver

import (
	"context"
	"sort"
	"time"
)

type agentAssignmentHistoryReader interface {
	LoadAgentAssignmentOperations(context.Context, string) ([]AssignmentOperationRecord, error)
}

func agentDashboardSnapshotFromState(state *storeState, agentID string, now time.Time) StoreSnapshot {
	out := StoreSnapshot{SavedAt: now, Events: map[string][]TaskEvent{}}
	ids := map[string]bool{}
	for _, id := range state.agentAssignments[agentID] {
		a, ok := state.assignments[id]
		if ok && a.AgentID == agentID {
			out.Assignments = append(out.Assignments, a)
			ids[id] = true
		}
	}
	tasks := map[string]bool{}
	for _, a := range out.Assignments {
		tasks[a.TaskID] = true
	}
	for task := range tasks {
		for _, e := range state.events[task] {
			if ids[e.AssignmentID] {
				out.Events[task] = append(out.Events[task], e)
			}
		}
	}
	return out
}

func mergeAgentDashboardOperations(snapshot StoreSnapshot, operations []AssignmentOperationRecord, agentID string) StoreSnapshot {
	sort.SliceStable(operations, func(i, j int) bool {
		a, b := assignmentOperationLastEventSeq(operations[i]), assignmentOperationLastEventSeq(operations[j])
		if a != b {
			return a < b
		}
		return operations[i].RecordedAt.Before(operations[j].RecordedAt)
	})
	byID := map[string]Assignment{}
	latestSeq := map[string]int64{}
	for _, a := range snapshot.Assignments {
		byID[a.ID] = a
	}
	seen := map[string]bool{}
	for _, events := range snapshot.Events {
		for _, e := range events {
			seen[dashboardEventKey(e)] = true
			if e.Seq > latestSeq[e.AssignmentID] {
				latestSeq[e.AssignmentID] = e.Seq
			}
		}
	}
	for _, op := range operations {
		if op.AgentID != agentID || op.Assignment.AgentID != agentID {
			continue
		}
		old, ok := byID[op.AssignmentID]
		if !ok || (!op.Assignment.UpdatedAt.Before(old.UpdatedAt) && assignmentOperationLastEventSeq(op) >= latestSeq[op.AssignmentID]) {
			byID[op.AssignmentID] = op.Assignment
			latestSeq[op.AssignmentID] = assignmentOperationLastEventSeq(op)
		}
		for _, e := range op.Events {
			if e.AgentID == agentID && !seen[dashboardEventKey(e)] {
				snapshot.Events[e.TaskID] = append(snapshot.Events[e.TaskID], e)
				seen[dashboardEventKey(e)] = true
			}
		}
	}
	snapshot.Assignments = nil
	for _, a := range byID {
		snapshot.Assignments = append(snapshot.Assignments, a)
	}
	return snapshot
}
func dashboardEventKey(e TaskEvent) string { return assignmentOperationEventReplayKey(e) }
