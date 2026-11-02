package riidoaiserver

import (
	"context"
	"fmt"
)

// The startup journal already contains historical runs. Discover runs created
// on another replica through creation records, one key per run, then read STATE
// only for missing or non-terminal runs. Progress volume never affects this query.
func (s *DynamoDBAssignmentOperationStore) dashboardSnapshot(ctx context.Context, agentID string, local StoreSnapshot, activeOnly bool, activeID string) (StoreSnapshot, map[string]int64, error) {
	credentials, err := s.credentialsProvider.Credentials(ctx)
	if err != nil {
		return local, nil, err
	}
	ids := map[string]bool{}
	if !activeOnly && s.enableAgentHistory {
		err = s.dashboardMemberIDs(ctx, credentials, agentID, ids)
		if err != nil {
			return local, nil, err
		}
	}
	// An active lease can point to a run which this replica has never seen.
	if activeID != "" {
		ids[activeID] = true
	}

	positions := map[string]int{}
	seenSeq := map[string]int64{}
	for i, a := range local.Assignments {
		positions[a.ID] = i
		if isDashboardTerminal(a.State) {
			delete(ids, a.ID)
		} else if !activeOnly || a.ID == activeID {
			ids[a.ID] = true
		}
	}
	for _, events := range local.Events {
		for _, e := range events {
			if e.Seq > seenSeq[e.AssignmentID] {
				seenSeq[e.AssignmentID] = e.Seq
			}
		}
	}
	missing := map[string]int64{}
	for id := range ids {
		projection, found, err := s.loadAssignmentProjection(ctx, id, credentials)
		if err != nil {
			return local, nil, err
		}
		if !found {
			if _, known := positions[id]; known {
				continue
			}
			projection, err = s.dashboardLegacyProjection(ctx, credentials, agentID, id)
			if err != nil {
				return local, nil, err
			}
		}
		a := projection.Assignment
		if a.ID != id || a.AgentID != agentID {
			return local, nil, fmt.Errorf("dashboard projection scope mismatch")
		}
		pos, exists := positions[id]
		if exists && projection.LastEventSeq < seenSeq[id] {
			continue
		}
		if exists {
			local.Assignments[pos] = a
		} else {
			local.Assignments = append(local.Assignments, a)
		}
		if !exists || projection.LastEventSeq > seenSeq[id] {
			missing[id] = projection.LastEventSeq
			if isDashboardTerminal(a.State) && !activeOnly {
				at, err := s.firstDashboardTerminal(ctx, credentials, a, projection.LastEventSeq)
				if err != nil {
					return local, nil, err
				}
				if !at.IsZero() {
					local.Events[a.TaskID] = append(local.Events[a.TaskID], TaskEvent{AssignmentID: id, TaskID: a.TaskID, AgentID: agentID, State: a.State, At: at})
				}
			}
		}
	}
	return local, missing, nil
}
