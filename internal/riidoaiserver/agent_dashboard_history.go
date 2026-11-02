package riidoaiserver

import "context"

func (s *Store) LoadAgentDashboardHistory(ctx context.Context, agentID string) (AgentDashboardHistory, error) {
	return s.loadAgentDashboard(ctx, agentID, false)
}

func (s *Store) LoadAgentDashboardActive(ctx context.Context, agentID string) (AgentDashboardHistory, error) {
	return s.loadAgentDashboard(ctx, agentID, true)
}

func (s *Store) loadAgentDashboard(ctx context.Context, agentID string, activeOnly bool) (AgentDashboardHistory, error) {
	reply := make(chan StoreSnapshot, 1)
	if err := s.send(ctx, agentDashboardSnapshotCmd{agentID: agentID, activeOnly: activeOnly, reply: reply}); err != nil {
		return AgentDashboardHistory{}, err
	}
	var snapshot StoreSnapshot
	select {
	case snapshot = <-reply:
	case <-ctx.Done():
		return AgentDashboardHistory{}, ctx.Err()
	}
	activeID := ""
	usesActiveLease := false
	if reader, ok := s.operationStore.(AssignmentActiveLeaseStore); ok {
		lease, found, err := reader.LoadAgentActiveAssignment(ctx, agentID)
		if err != nil {
			return AgentDashboardHistory{}, err
		}
		usesActiveLease = true
		if found {
			activeID = lease.ActiveAssignmentID
		}
	}
	detailReader, projected := s.operationStore.(agentDashboardProjectionReader)
	var missing map[string]int64
	if projected {
		var err error
		snapshot, missing, err = detailReader.dashboardSnapshot(ctx, agentID, snapshot, activeOnly, activeID)
		if err != nil {
			return AgentDashboardHistory{}, err
		}
	} else if reader, ok := s.operationStore.(agentAssignmentHistoryReader); ok && !activeOnly {
		operations, err := reader.LoadAgentAssignmentOperations(ctx, agentID)
		if err != nil {
			return AgentDashboardHistory{}, err
		}
		snapshot = mergeAgentDashboardOperations(snapshot, operations, agentID)
	}

	result := dashboardHistoryFromSnapshot(snapshot, agentID)
	result.UsesActiveLease = usesActiveLease
	result.ActiveAssignmentID = activeID
	if activeOnly {
		active := result.Runs[:0]
		for _, run := range result.Runs {
			if !isDashboardTerminal(run.Assignment.State) {
				active = append(active, run)
			}
		}
		result.Runs = active
	}
	result.activeOnly = activeOnly
	result.detailReader = detailReader
	if !activeOnly {
		result.missing = missing
	}
	result.Durable = s.operationStore != nil || s.snapshotStore != nil
	if err := s.loadDashboardQueue(ctx, agentID, &result); err != nil {
		return AgentDashboardHistory{}, err
	}

	return result, nil
}
